package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	"go.uber.org/zap"
)

const (
	defaultBatchSize     = 1000
	defaultFlushInterval = 5 * time.Second
)

// clickhouseBatch is a local interface matching the methods we use from driver.Batch.
type clickhouseBatch interface {
	Append(v ...interface{}) error
	Send() error
}

type batchEntry struct {
	batch clickhouseBatch
	count int
}

type clickhouseRepository struct {
	conn chDriver.Conn
	log  logger.LoggerInterface

	mu        sync.Mutex
	batches   map[string]*batchEntry
	batchSize int

	flushTicker *time.Ticker
	flushDone   chan struct{}
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewClickhouseRepository(conn chDriver.Conn, log logger.LoggerInterface) Repository {
	ctx, cancel := context.WithCancel(context.Background())

	r := &clickhouseRepository{
		conn:        conn,
		log:         log,
		batches:     make(map[string]*batchEntry),
		batchSize:   defaultBatchSize,
		flushTicker: time.NewTicker(defaultFlushInterval),
		flushDone:   make(chan struct{}),
		ctx:         ctx,
		cancel:      cancel,
	}

	go r.flushLoop()

	return r
}

func (r *clickhouseRepository) InsertTransactionEvent(ctx context.Context, event events.TransactionEvent) error {
	query := `INSERT INTO transaction_events (
		transaction_id, transaction_no, card_number, card_type, card_provider, amount,
		payment_method, merchant_id, merchant_name, status, apikey, created_at
	)`
	return r.appendToBatch(ctx, "transaction", query,
		event.TransactionID, event.TransactionNo, event.CardNumber, event.CardType, event.CardProvider, event.Amount,
		event.PaymentMethod, event.MerchantID, event.MerchantName, event.Status, event.ApiKey, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertTopupEvent(ctx context.Context, event events.TopupEvent) error {
	query := `INSERT INTO topup_events (
		topup_id, topup_no, card_number, card_type, card_provider, amount,
		payment_method, status, created_at
	)`
	return r.appendToBatch(ctx, "topup", query,
		event.TopupID, event.TopupNo, event.CardNumber, event.CardType, event.CardProvider,
		event.Amount, event.PaymentMethod, event.Status, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertTransferEvent(ctx context.Context, event events.TransferEvent) error {
	query := `INSERT INTO transfer_events (
		transfer_id, transfer_no, source_card, destination_card, amount, status, created_at
	)`
	return r.appendToBatch(ctx, "transfer", query,
		event.TransferID, event.TransferNo, event.SourceCard, event.DestinationCard,
		event.Amount, event.Status, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertWithdrawEvent(ctx context.Context, event events.WithdrawEvent) error {
	query := `INSERT INTO withdraw_events (
		withdraw_id, withdraw_no, card_number, card_type, amount, status, created_at
	)`
	return r.appendToBatch(ctx, "withdraw", query,
		event.WithdrawID, event.WithdrawNo, event.CardNumber, event.CardType,
		event.Amount, event.Status, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertSaldoEvent(ctx context.Context, event events.SaldoEvent) error {
	query := `INSERT INTO saldo_events (
		card_number, total_balance, created_at
	)`
	return r.appendToBatch(ctx, "saldo", query,
		event.CardNumber, event.TotalBalance, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertMerchantEvent(ctx context.Context, event events.MerchantEvent) error {
	query := `INSERT INTO merchant_events (
		merchant_id, user_id, name, email, status, created_at
	)`
	return r.appendToBatch(ctx, "merchant", query,
		event.MerchantID, event.UserID, event.Name, event.Email, event.Status, event.CreatedAt,
	)
}

func (r *clickhouseRepository) InsertCardEvent(ctx context.Context, event events.CardEvent) error {
	query := `INSERT INTO card_events (
		card_id, user_id, card_number, card_type, card_provider, status, created_at
	)`
	return r.appendToBatch(ctx, "card", query,
		event.CardID, event.UserID, event.CardNumber, event.CardType, event.CardProvider, event.Status, event.CreatedAt,
	)
}

func (r *clickhouseRepository) Flush(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var lastErr error
	for key, entry := range r.batches {
		if entry.count == 0 {
			continue
		}
		if err := entry.batch.Send(); err != nil {
			r.log.Error("Failed to flush ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
				zap.Error(err),
			)
			lastErr = fmt.Errorf("flush batch %s: %w", key, err)
		} else {
			r.log.Debug("Flushed ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
			)
		}
		delete(r.batches, key)
	}
	return lastErr
}

func (r *clickhouseRepository) Close() error {
	r.flushTicker.Stop()
	r.cancel()
	<-r.flushDone

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	return r.Flush(ctx)
}

func (r *clickhouseRepository) appendToBatch(ctx context.Context, key, query string, args ...interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	entry, exists := r.batches[key]
	if !exists {
		batch, err := r.conn.PrepareBatch(ctx, query)
		if err != nil {
			r.log.Error("Failed to prepare ClickHouse batch",
				zap.String("batch", key),
				zap.Error(err),
			)
			return fmt.Errorf("prepare batch %s: %w", key, err)
		}
		entry = &batchEntry{
			batch: batch,
		}
		r.batches[key] = entry
	}

	if err := entry.batch.Append(args...); err != nil {
		r.log.Error("Failed to append to ClickHouse batch",
			zap.String("batch", key),
			zap.Error(err),
		)
		return fmt.Errorf("append to batch %s: %w", key, err)
	}

	entry.count++
	if entry.count >= r.batchSize {
		if err := entry.batch.Send(); err != nil {
			r.log.Error("Failed to send full ClickHouse batch",
				zap.String("batch", key),
				zap.Int("rows", entry.count),
				zap.Error(err),
			)
			delete(r.batches, key)
			return fmt.Errorf("send batch %s: %w", key, err)
		}
		r.log.Debug("Sent full ClickHouse batch",
			zap.String("batch", key),
			zap.Int("rows", entry.count),
		)
		delete(r.batches, key)
	}

	return nil
}

func (r *clickhouseRepository) flushLoop() {
	defer close(r.flushDone)

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-r.flushTicker.C:
			ctx, cancel := context.WithTimeout(r.ctx, 30*time.Second)
			_ = r.Flush(ctx)
			cancel()
		}
	}
}
