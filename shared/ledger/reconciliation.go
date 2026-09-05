package ledger

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

// ReconciliationRow represents a balance-vs-ledger mismatch row.
// Defined locally so the ledger package has no dependency on saldo's sqlc schema.
type ReconciliationRow struct {
	SaldoID        int32
	CardNumber     string
	CurrentBalance int64
	LedgerBalance  int64
	Difference     int64
	LedgerEntries  int64
}

// ReconciliationStore is deliberately small so the worker can be tested with a fake.
type ReconciliationStore interface {
	ListReconciliationMismatches(context.Context, int32) ([]*ReconciliationRow, error)
}

type reconciliationQueueStore interface {
	EnqueueReconciliationMismatch(context.Context, *ReconciliationRow) error
}

type Reconciler struct {
	store           ReconciliationStore
	logger          logger.LoggerInterface
	interval        time.Duration
	batchSize       int32
	mismatchCounter metric.Int64Counter
}

func NewReconciler(store ReconciliationStore, log logger.LoggerInterface, interval time.Duration, batchSize int32) *Reconciler {
	if interval <= 0 {
		interval = 5 * time.Minute
	}
	if batchSize <= 0 {
		batchSize = 100
	}

	meter := otel.Meter("shared/ledger")
	mismatchCounter, err := meter.Int64Counter(
		"ledger_reconciliation_mismatch",
		metric.WithDescription("Total number of balance ledger mismatches detected"),
		metric.WithUnit("1"),
	)
	if err != nil {
		log.Error("ledger: failed to create mismatch counter", zap.Error(err))
	}

	return &Reconciler{store: store, logger: log, interval: interval, batchSize: batchSize, mismatchCounter: mismatchCounter}
}

// Start runs periodic reconciliation until ctx is cancelled.
func (r *Reconciler) Start(ctx context.Context) {
	r.RunOnce(ctx)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.RunOnce(ctx)
		}
	}
}

// RunOnce performs one reconciliation pass. It is exported for on-demand jobs
// and tests while the periodic worker uses the same code path.
func (r *Reconciler) RunOnce(ctx context.Context) {
	mismatches, err := r.store.ListReconciliationMismatches(ctx, r.batchSize)
	if err != nil {
		r.logError("ledger reconciliation failed", zap.Error(err))
		return
	}
	if len(mismatches) == 0 {
		r.logDebug("ledger reconciliation completed with no mismatches")
		return
	}

	queue, durable := r.store.(reconciliationQueueStore)
	for _, mismatch := range mismatches {
		if r.mismatchCounter != nil {
			r.mismatchCounter.Add(ctx, 1, metric.WithAttributes(
				attribute.Int("saldo_id", int(mismatch.SaldoID)),
				attribute.String("card_number", mismatch.CardNumber),
			))
		}
		if durable {
			if err := queue.EnqueueReconciliationMismatch(ctx, mismatch); err != nil {
				r.logError("failed to enqueue ledger mismatch", zap.Int32("saldo_id", mismatch.SaldoID), zap.Error(err))
			}
		}
		r.logError("balance ledger mismatch detected",
			zap.Int32("saldo_id", mismatch.SaldoID),
			zap.String("card_number", mismatch.CardNumber),
			zap.Int64("current_balance", mismatch.CurrentBalance),
			zap.Int64("ledger_balance", mismatch.LedgerBalance),
			zap.Int64("difference", mismatch.Difference),
			zap.Int64("ledger_entries", mismatch.LedgerEntries),
		)
	}
}

func (r *Reconciler) logError(message string, fields ...zap.Field) {
	if r.logger != nil {
		r.logger.Error(message, fields...)
	}
}

func (r *Reconciler) logDebug(message string, fields ...zap.Field) {
	if r.logger != nil {
		r.logger.Debug(message, fields...)
	}
}
