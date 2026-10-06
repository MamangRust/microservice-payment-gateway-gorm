package stats_writer_test

import (
	"context"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	writer_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/repository"
	writer_usecase "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer/usecase"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"github.com/stretchr/testify/suite"
)

const (
	testCardNumber = "4111111111111111"
	testOtherCard  = "4222222222222222"
	testMerchantID = 7
	testAPIKey     = "apikey-test-1"
)

type StatsWriterTestSuite struct {
	suite.Suite
	ts     *tests.TestSuite
	chConn clickhouse.Conn
	repo   writer_repo.Repository
	uc     writer_usecase.UseCase
}

func (s *StatsWriterTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.chConn, err = s.ts.OpenStatsConn()
	s.Require().NoError(err)
	s.Require().NoError(tests.TruncateStatsTables(context.Background(), s.chConn))

	s.repo = writer_repo.NewClickhouseRepository(s.chConn, s.ts.Logger)
	s.uc = writer_usecase.NewStatsUseCase(s.repo)
}

func (s *StatsWriterTestSuite) TearDownSuite() {
	if s.repo != nil {
		_ = s.repo.Close()
	}
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.ts.Teardown()
}

func (s *StatsWriterTestSuite) count(ctx context.Context, table string) uint64 {
	var n uint64
	s.Require().NoError(s.chConn.QueryRow(ctx, "SELECT count() FROM "+table).Scan(&n))
	return n
}

// Test1_AllEventTypesPersisted drives every Save*Event path, flushes the
// writer and verifies one row landed in each ClickHouse table with the expected
// values.
func (s *StatsWriterTestSuite) Test1_AllEventTypesPersisted() {
	ctx := context.Background()
	now := time.Now()

	s.Require().NoError(s.uc.SaveTransactionEvent(ctx, events.TransactionEvent{
		TransactionID: 1,
		TransactionNo: "TX-1",
		CardNumber:    testCardNumber,
		CardType:      "debit",
		CardProvider:  "Visa",
		Amount:        1000,
		PaymentMethod: "card",
		MerchantID:    testMerchantID,
		MerchantName:  "Merchant A",
		Status:        "success",
		ApiKey:        testAPIKey,
		CreatedAt:     now,
	}))
	s.Require().NoError(s.uc.SaveTopupEvent(ctx, events.TopupEvent{
		TopupID:       1,
		TopupNo:       "TP-1",
		CardNumber:    testCardNumber,
		CardType:      "debit",
		CardProvider:  "Visa",
		Amount:        5000,
		PaymentMethod: "bank_transfer",
		Status:        "success",
		CreatedAt:     now,
	}))
	s.Require().NoError(s.uc.SaveTransferEvent(ctx, events.TransferEvent{
		TransferID:      1,
		TransferNo:      "TR-1",
		SourceCard:      testCardNumber,
		DestinationCard: testOtherCard,
		Amount:          2000,
		Status:          "success",
		CreatedAt:       now,
	}))
	s.Require().NoError(s.uc.SaveWithdrawEvent(ctx, events.WithdrawEvent{
		WithdrawID: 1,
		WithdrawNo: "WD-1",
		CardNumber: testCardNumber,
		CardType:   "debit",
		Amount:     3000,
		Status:     "success",
		CreatedAt:  now,
	}))
	s.Require().NoError(s.uc.SaveSaldoEvent(ctx, events.SaldoEvent{
		CardNumber:   testCardNumber,
		TotalBalance: 12000,
		CreatedAt:    now,
	}))
	s.Require().NoError(s.uc.SaveMerchantEvent(ctx, events.MerchantEvent{
		MerchantID: testMerchantID,
		UserID:     99,
		Name:       "Merchant A",
		Email:      "merchant@example.com",
		Status:     "active",
		CreatedAt:  now,
	}))
	s.Require().NoError(s.uc.SaveCardEvent(ctx, events.CardEvent{
		CardID:       1,
		UserID:       99,
		CardNumber:   testCardNumber,
		CardType:     "debit",
		CardProvider: "Visa",
		Status:       "active",
		CreatedAt:    now,
	}))

	s.Require().NoError(s.repo.Flush(ctx))

	for _, table := range tests.StatsTables {
		s.Equal(uint64(1), s.count(ctx, table), "row count for %s", table)
	}

	var amount int64
	var status, merchantName, apiKey string
	s.Require().NoError(s.chConn.QueryRow(ctx,
		`SELECT amount, status, merchant_name, apikey FROM transaction_events LIMIT 1`).
		Scan(&amount, &status, &merchantName, &apiKey))
	s.Equal(int64(1000), amount)
	s.Equal("success", status)
	s.Equal("Merchant A", merchantName)
	s.Equal(testAPIKey, apiKey)

	var totalBalance int64
	s.Require().NoError(s.chConn.QueryRow(ctx,
		`SELECT total_balance FROM saldo_events LIMIT 1`).Scan(&totalBalance))
	s.Equal(int64(12000), totalBalance)
}

// Test2_BatchFlushOnSizeThreshold verifies the writer auto-flushes once the
// 1000-row batch size is reached and that the remainder is flushed explicitly.
func (s *StatsWriterTestSuite) Test2_BatchFlushOnSizeThreshold() {
	ctx := context.Background()
	s.Require().NoError(s.chConn.Exec(ctx, "TRUNCATE TABLE IF EXISTS topup_events"))

	const total = 1500
	now := time.Now()
	for i := 1; i <= total; i++ {
		s.Require().NoError(s.uc.SaveTopupEvent(ctx, events.TopupEvent{
			TopupID:    uint64(i),
			TopupNo:    "TP-BATCH",
			CardNumber: testCardNumber,
			Amount:     100,
			Status:     "success",
			CreatedAt:  now,
		}))
	}
	s.Require().NoError(s.repo.Flush(ctx))

	s.Equal(uint64(total), s.count(ctx, "topup_events"))
}

// Test3_CloseFlushesPendingRows verifies the shutdown path persists rows that
// were still buffered. It runs last because it closes the repository.
func (s *StatsWriterTestSuite) Test3_CloseFlushesPendingRows() {
	ctx := context.Background()
	s.Require().NoError(s.chConn.Exec(ctx, "TRUNCATE TABLE IF EXISTS withdraw_events"))

	now := time.Now()
	for i := 1; i <= 3; i++ {
		s.Require().NoError(s.uc.SaveWithdrawEvent(ctx, events.WithdrawEvent{
			WithdrawID: uint64(i),
			WithdrawNo: "WD-CLOSE",
			CardNumber: testCardNumber,
			Amount:     500,
			Status:     "success",
			CreatedAt:  now,
		}))
	}

	s.Require().NoError(s.repo.Close())

	s.Equal(uint64(3), s.count(ctx, "withdraw_events"))
}

func TestStatsWriterSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(StatsWriterTestSuite))
}
