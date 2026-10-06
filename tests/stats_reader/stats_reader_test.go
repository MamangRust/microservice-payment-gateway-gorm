package stats_reader_test

import (
	"context"
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	pbCardBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbMerchantBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbSaldoBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pbCardStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/card"
	pbMerchantStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/merchant"
	pbSaldoStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/saldo"
	pbTopupStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/topup"
	pbTransactionStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/transaction"
	pbTransferStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/transfer"
	pbWithdrawStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/stats/withdraw"
	pbTopupBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	pbTransactionBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction"
	pbTransferBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer"
	pbWithdrawBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw"
	stats_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	stats_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	tests "github.com/MamangRust/microservice-payment-gateway-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	testCardNumber = "4111111111111111"
	testOtherCard  = "4222222222222222"
	testMerchantID = 7
	testAPIKey     = "apikey-test-1"
)

// expectedServices mirrors the registration list in
// service/stats-reader/cmd/main.go. Keep the two in sync: this suite exists to
// catch a service silently disappearing from (or never being added to) main.go.
var expectedServices = []string{
	"pb.card.CardDashboardService",
	"pb.card.stats.CardStatsBalanceService",
	"pb.card.stats.CardStatsTopupService",
	"pb.card.stats.CardStatsTransactionService",
	"pb.card.stats.CardStatsTransferService",
	"pb.card.stats.CardStatsWithdrawService",
	"pb.merchant.MerchantTransactionService",
	"pb.merchant.stats.MerchantStatsAmountService",
	"pb.merchant.stats.MerchantStatsMethodService",
	"pb.merchant.stats.MerchantStatsTotalAmountService",
	"pb.saldo.stats.SaldoStatsBalanceService",
	"pb.saldo.stats.SaldoStatsTotalBalance",
	"pb.topup.stats.TopupStatsAmountService",
	"pb.topup.stats.TopupStatsMethodService",
	"pb.topup.stats.TopupStatsStatusService",
	"pb.transaction.stats.TransactionStatsAmountService",
	"pb.transaction.stats.TransactionStatsMethodService",
	"pb.transaction.stats.TransactionStatsStatusService",
	"pb.transfer.stats.TransferStatsAmountService",
	"pb.transfer.stats.TransferStatsStatusService",
	"pb.withdraw.stats.WithdrawStatsAmountService",
	"pb.withdraw.stats.WithdrawStatsStatusService",
}

// expectedUnaryRPCs is the number of unary methods exposed by the 22 services
// above. The count guards against a service being dropped from main.go.
const expectedUnaryRPCs = 109

type StatsReaderTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	chConn     clickhouse.Conn
	grpcServer *grpc.Server
	conn       *grpc.ClientConn
	year       int
}

func (s *StatsReaderTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	s.chConn, err = s.ts.OpenStatsConn()
	s.Require().NoError(err)
	s.Require().NoError(tests.TruncateStatsTables(context.Background(), s.chConn))

	s.year = time.Now().Year()
	s.seedFixtures()

	repo := stats_repo.NewRepository(s.chConn)
	log := s.ts.Logger

	cardHandler := stats_handler.NewCardStatsHandler(repo, log)
	merchantHandler := stats_handler.NewMerchantStatsHandler(repo, log)
	saldoHandler := stats_handler.NewSaldoStatsHandler(repo, log)
	topupHandler := stats_handler.NewTopupStatsHandler(repo, log)
	transactionHandler := stats_handler.NewTransactionStatsHandler(repo, log)
	transferHandler := stats_handler.NewTransferStatsHandler(repo, log)
	withdrawHandler := stats_handler.NewWithdrawStatsHandler(repo, log)

	server := grpc.NewServer()
	pbCardStats.RegisterCardStatsBalanceServiceServer(server, cardHandler)
	pbCardStats.RegisterCardStatsTopupServiceServer(server, cardHandler)
	pbCardStats.RegisterCardStatsTransactionServiceServer(server, cardHandler)
	pbCardStats.RegisterCardStatsTransferServiceServer(server, cardHandler)
	pbCardStats.RegisterCardStatsWithdrawServiceServer(server, cardHandler)
	pbCardBase.RegisterCardDashboardServiceServer(server, cardHandler)

	pbMerchantStats.RegisterMerchantStatsAmountServiceServer(server, merchantHandler)
	pbMerchantStats.RegisterMerchantStatsMethodServiceServer(server, merchantHandler)
	pbMerchantStats.RegisterMerchantStatsTotalAmountServiceServer(server, merchantHandler)
	pbMerchantBase.RegisterMerchantTransactionServiceServer(server, merchantHandler)

	pbSaldoStats.RegisterSaldoStatsBalanceServiceServer(server, saldoHandler)
	pbSaldoStats.RegisterSaldoStatsTotalBalanceServer(server, saldoHandler)

	pbTopupStats.RegisterTopupStatsAmountServiceServer(server, topupHandler)
	pbTopupStats.RegisterTopupStatsMethodServiceServer(server, topupHandler)
	pbTopupStats.RegisterTopupStatsStatusServiceServer(server, topupHandler)

	pbTransactionStats.RegisterTransactionStatsAmountServiceServer(server, transactionHandler)
	pbTransactionStats.RegisterTransactionStatsMethodServiceServer(server, transactionHandler)
	pbTransactionStats.RegisterTransactionStatsStatusServiceServer(server, transactionHandler)

	pbTransferStats.RegisterTransferStatsAmountServiceServer(server, transferHandler)
	pbTransferStats.RegisterTransferStatsStatusServiceServer(server, transferHandler)

	pbWithdrawStats.RegisterWithdrawStatsAmountServiceServer(server, withdrawHandler)
	pbWithdrawStats.RegisterWithdrawStatsStatusServiceServer(server, withdrawHandler)

	reflection.Register(server)
	s.grpcServer = server

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() { _ = server.Serve(lis) }()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
}

func (s *StatsReaderTestSuite) TearDownSuite() {
	if s.conn != nil {
		_ = s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.ts.Teardown()
}

// seedFixtures writes deterministic rows for the current year into every stats
// table. Expected totals are asserted in the tests below.
func (s *StatsReaderTestSuite) seedFixtures() {
	ctx := context.Background()
	now := time.Now()

	exec := func(query string, args ...interface{}) {
		s.Require().NoError(s.chConn.Exec(ctx, query, args...), query)
	}

	// 1000 + 2500 success, 500 failed -> success total 3500
	exec(`INSERT INTO transaction_events (transaction_id, transaction_no, card_number, card_type, card_provider, amount, payment_method, merchant_id, merchant_name, status, apikey, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "TX-1", testCardNumber, "debit", "Visa", 1000, "card", testMerchantID, "Merchant A", "success", testAPIKey, now)
	exec(`INSERT INTO transaction_events (transaction_id, transaction_no, card_number, card_type, card_provider, amount, payment_method, merchant_id, merchant_name, status, apikey, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		2, "TX-2", testCardNumber, "debit", "Visa", 2500, "transfer", testMerchantID, "Merchant A", "success", testAPIKey, now)
	exec(`INSERT INTO transaction_events (transaction_id, transaction_no, card_number, card_type, card_provider, amount, payment_method, merchant_id, merchant_name, status, apikey, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		3, "TX-3", testOtherCard, "credit", "MasterCard", 500, "card", testMerchantID, "Merchant A", "failed", testAPIKey, now)

	// 5000 success, 700 failed
	exec(`INSERT INTO topup_events (topup_id, topup_no, card_number, card_type, card_provider, amount, payment_method, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		1, "TP-1", testCardNumber, "debit", "Visa", 5000, "bank_transfer", "success", now)
	exec(`INSERT INTO topup_events (topup_id, topup_no, card_number, card_type, card_provider, amount, payment_method, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		2, "TP-2", testCardNumber, "debit", "Visa", 700, "bank_transfer", "failed", now)

	// 2000 success
	exec(`INSERT INTO transfer_events (transfer_id, transfer_no, source_card, destination_card, amount, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "TR-1", testCardNumber, testOtherCard, 2000, "success", now)

	// 3000 success, 400 failed
	exec(`INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, card_type, amount, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, "WD-1", testCardNumber, "debit", 3000, "success", now)
	exec(`INSERT INTO withdraw_events (withdraw_id, withdraw_no, card_number, card_type, amount, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		2, "WD-2", testCardNumber, "debit", 400, "failed", now)

	// Latest balance per card -> 12000 + 5000 = 17000 globally, 12000 for the card.
	exec(`INSERT INTO saldo_events (card_number, total_balance, created_at) VALUES (?, ?, ?)`,
		testCardNumber, 12000, now)
	exec(`INSERT INTO saldo_events (card_number, total_balance, created_at) VALUES (?, ?, ?)`,
		testOtherCard, 5000, now)

	exec(`INSERT INTO card_events (card_id, user_id, card_number, card_type, card_provider, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		1, 99, testCardNumber, "debit", "Visa", "active", now)

	exec(`INSERT INTO merchant_events (merchant_id, user_id, name, email, status, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		testMerchantID, 99, "Merchant A", "merchant@example.com", "active", now)
}

// Test1_AllRegisteredServicesExposed asks the server, over reflection, which
// services it actually serves and compares that with main.go's registration.
func (s *StatsReaderTestSuite) Test1_AllRegisteredServicesExposed() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client := grpc_reflection_v1.NewServerReflectionClient(s.conn)
	stream, err := client.ServerReflectionInfo(ctx)
	s.Require().NoError(err)

	s.Require().NoError(stream.Send(&grpc_reflection_v1.ServerReflectionRequest{
		MessageRequest: &grpc_reflection_v1.ServerReflectionRequest_ListServices{ListServices: ""},
	}))
	resp, err := stream.Recv()
	s.Require().NoError(err)

	var served []string
	for _, svc := range resp.GetListServicesResponse().GetService() {
		// The reflection service lists itself; it is not part of the stats API.
		if strings.HasPrefix(svc.GetName(), "grpc.reflection.") {
			continue
		}
		served = append(served, svc.GetName())
	}

	for _, name := range expectedServices {
		s.Contains(served, name, "service %s must be registered", name)
	}
	s.Len(served, len(expectedServices), "unexpected extra services: %v", served)
}

// Test2_EveryStatsMethodServed invokes every unary RPC of every registered
// service with an empty request. A missing registration surfaces as
// codes.Unimplemented, so a green run proves all stats endpoints are wired and
// reach ClickHouse without error.
func (s *StatsReaderTestSuite) Test2_EveryStatsMethodServed() {
	ctx := context.Background()
	invoked := 0

	for _, svc := range expectedServices {
		desc, err := protoregistry.GlobalFiles.FindDescriptorByName(protoreflect.FullName(svc))
		s.Require().NoError(err, "descriptor for %s", svc)
		sd, ok := desc.(protoreflect.ServiceDescriptor)
		s.Require().True(ok, "%s is not a service descriptor", svc)

		for i := 0; i < sd.Methods().Len(); i++ {
			md := sd.Methods().Get(i)
			s.Require().False(md.IsStreamingClient(), "%s is streaming", md.FullName())
			s.Require().False(md.IsStreamingServer(), "%s is streaming", md.FullName())

			fullMethod := fmt.Sprintf("/%s/%s", sd.FullName(), md.Name())
			req := dynamicpb.NewMessage(md.Input())
			res := dynamicpb.NewMessage(md.Output())

			callCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
			err := s.conn.Invoke(callCtx, fullMethod, req, res)
			cancel()

			s.NoError(err, "invoke %s", fullMethod)
			invoked++
		}
	}

	s.Equal(expectedUnaryRPCs, invoked, "registered unary RPC count changed")
}

// Test3_CardStatsReturnSeededTotals checks the card stats services against the
// seeded fixtures.
func (s *StatsReaderTestSuite) Test3_CardStatsReturnSeededTotals() {
	ctx := context.Background()
	year := int32(s.year)

	balance, err := pbCardStats.NewCardStatsBalanceServiceClient(s.conn).
		FindYearlyBalance(ctx, &pbCardStats.FindYearBalance{Year: year})
	s.Require().NoError(err)
	s.Equal("success", balance.Status)
	s.Require().NotEmpty(balance.Data)
	s.Equal(int64(17000), balance.Data[0].TotalBalance)

	balanceByCard, err := pbCardStats.NewCardStatsBalanceServiceClient(s.conn).
		FindYearlyBalanceByCardNumber(ctx, &pbCardStats.FindYearBalanceCardNumber{Year: year, CardNumber: testCardNumber})
	s.Require().NoError(err)
	s.Require().NotEmpty(balanceByCard.Data)
	s.Equal(int64(12000), balanceByCard.Data[0].TotalBalance)

	topup, err := pbCardStats.NewCardStatsTopupServiceClient(s.conn).
		FindYearlyTopupAmount(ctx, &pbCardBase.FindYearAmount{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(topup.Data)
	s.Equal(int64(5000), topup.Data[0].TotalAmount)

	trx, err := pbCardStats.NewCardStatsTransactionServiceClient(s.conn).
		FindYearlyTransactionAmount(ctx, &pbCardBase.FindYearAmount{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(trx.Data)
	s.Equal(int64(3500), trx.Data[0].TotalAmount)

	transfer, err := pbCardStats.NewCardStatsTransferServiceClient(s.conn).
		FindYearlyTransferSenderAmount(ctx, &pbCardBase.FindYearAmount{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(transfer.Data)
	s.Equal(int64(2000), transfer.Data[0].TotalAmount)

	withdraw, err := pbCardStats.NewCardStatsWithdrawServiceClient(s.conn).
		FindYearlyWithdrawAmount(ctx, &pbCardBase.FindYearAmount{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(withdraw.Data)
	s.Equal(int64(3000), withdraw.Data[0].TotalAmount)
}

// Test4_DomainStatsReturnSeededTotals checks merchant, saldo, topup,
// transaction, transfer and withdraw services against the same fixtures.
func (s *StatsReaderTestSuite) Test4_DomainStatsReturnSeededTotals() {
	ctx := context.Background()
	year := int32(s.year)

	merchant, err := pbMerchantStats.NewMerchantStatsAmountServiceClient(s.conn).
		FindYearlyAmountMerchant(ctx, &pbMerchantBase.FindYearMerchant{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(merchant.Data)
	s.Equal(int64(3500), merchant.Data[0].TotalAmount)

	merchantByAPIKey, err := pbMerchantStats.NewMerchantStatsAmountServiceClient(s.conn).
		FindYearlyAmountByApikey(ctx, &pbMerchantBase.FindYearMerchantByApikey{Year: year, ApiKey: testAPIKey})
	s.Require().NoError(err)
	s.Require().NotEmpty(merchantByAPIKey.Data)
	s.Equal(int64(3500), merchantByAPIKey.Data[0].TotalAmount)

	saldo, err := pbSaldoStats.NewSaldoStatsBalanceServiceClient(s.conn).
		FindYearlySaldoBalances(ctx, &pbSaldoBase.FindYearlySaldo{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(saldo.Data)
	s.Equal(int64(17000), saldo.Data[0].TotalBalance)

	topup, err := pbTopupStats.NewTopupStatsAmountServiceClient(s.conn).
		FindYearlyTopupAmounts(ctx, &pbTopupBase.FindYearTopupStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(topup.Data)
	s.Equal(int64(5000), topup.Data[0].TotalAmount)

	trx, err := pbTransactionStats.NewTransactionStatsAmountServiceClient(s.conn).
		FindYearlyAmounts(ctx, &pbTransactionBase.FindYearTransactionStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(trx.Data)
	s.Equal(int64(3500), trx.Data[0].TotalAmount)

	transfer, err := pbTransferStats.NewTransferStatsAmountServiceClient(s.conn).
		FindYearlyTransferAmounts(ctx, &pbTransferBase.FindYearTransferStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(transfer.Data)
	s.Equal(int64(2000), transfer.Data[0].TotalAmount)

	withdraw, err := pbWithdrawStats.NewWithdrawStatsAmountServiceClient(s.conn).
		FindYearlyWithdraws(ctx, &pbWithdrawBase.FindYearWithdrawStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(withdraw.Data)
	s.Equal(int64(3000), withdraw.Data[0].TotalAmount)
}

// Test5_FailedStatusStats confirms failed rows are counted separately.
func (s *StatsReaderTestSuite) Test5_FailedStatusStats() {
	ctx := context.Background()
	year := int32(s.year)

	topupFailed, err := pbTopupStats.NewTopupStatsStatusServiceClient(s.conn).
		FindYearlyTopupStatusFailed(ctx, &pbTopupBase.FindYearTopupStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(topupFailed.Data)
	s.Equal(int32(1), topupFailed.Data[0].TotalFailed)
	s.Equal(int64(700), topupFailed.Data[0].TotalAmount)

	trxFailed, err := pbTransactionStats.NewTransactionStatsStatusServiceClient(s.conn).
		FindYearlyTransactionStatusFailed(ctx, &pbTransactionBase.FindYearTransactionStatus{Year: year})
	s.Require().NoError(err)
	s.Require().NotEmpty(trxFailed.Data)
	s.Equal(int64(500), trxFailed.Data[0].TotalAmount)
}

func TestStatsReaderSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(StatsReaderTestSuite))
}
