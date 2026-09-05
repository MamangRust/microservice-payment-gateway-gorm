package handler

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbCardBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbCardStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	"google.golang.org/protobuf/types/known/emptypb"
)

type CardRepository interface {
	GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	GetYearlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, startYear, endYear int) ([]repository.YearlyAmount, error)
}

type CardStatsHandler struct {
	pbCardStats.UnimplementedCardStatsBalanceServiceServer
	pbCardStats.UnimplementedCardStatsTopupServiceServer
	pbCardStats.UnimplementedCardStatsTransactionServiceServer
	pbCardStats.UnimplementedCardStatsTransferServiceServer
	pbCardStats.UnimplementedCardStatsWithdrawServiceServer
	pbCardBase.UnimplementedCardDashboardServiceServer
	repo CardRepository
	log  logger.LoggerInterface
}

func NewCardStatsHandler(repo CardRepository, log logger.LoggerInterface) *CardStatsHandler {
	return &CardStatsHandler{
		repo: repo,
		log:  log,
	}
}

// --- Card Dashboard Service ---

func (h *CardStatsHandler) DashboardCard(ctx context.Context, req *emptypb.Empty) (*card.ApiResponseDashboardCard, error) {
	year := time.Now().Year()

	type SaldoRepo interface {
		GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]repository.YearlyAmount, error)
	}

	var balanceData []repository.YearlyAmount
	if sr, ok := h.repo.(SaldoRepo); ok {
		balanceData, _ = sr.GetYearlyTotalSaldo(ctx, year, year)
	} else {
		balanceData, _ = h.repo.GetYearlyAmounts(ctx, "saldo_events", "", nil, year, year)
	}

	topupData, _ := h.repo.GetYearlyAmounts(ctx, "topup_events", "", nil, year, year)
	withdrawData, _ := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "", nil, year, year)
	transactionData, _ := h.repo.GetYearlyAmounts(ctx, "transaction_events", "", nil, year, year)
	transferData, _ := h.repo.GetYearlyAmounts(ctx, "transfer_events", "", nil, year, year)

	totalBalance := int64(0)
	if len(balanceData) > 0 {
		totalBalance = balanceData[0].TotalAmount
	}
	totalTopup := int64(0)
	if len(topupData) > 0 {
		totalTopup = topupData[0].TotalAmount
	}
	totalWithdraw := int64(0)
	if len(withdrawData) > 0 {
		totalWithdraw = withdrawData[0].TotalAmount
	}
	totalTransaction := int64(0)
	if len(transactionData) > 0 {
		totalTransaction = transactionData[0].TotalAmount
	}
	totalTransfer := int64(0)
	if len(transferData) > 0 {
		totalTransfer = transferData[0].TotalAmount
	}

	return &card.ApiResponseDashboardCard{
		Status:  "success",
		Message: "Retrieved global card dashboard",
		Data: &card.CardResponseDashboard{
			TotalBalance:     totalBalance,
			TotalTopup:       totalTopup,
			TotalWithdraw:    totalWithdraw,
			TotalTransaction: totalTransaction,
			TotalTransfer:    totalTransfer,
		},
	}, nil
}

func (h *CardStatsHandler) DashboardCardNumber(ctx context.Context, req *card.FindByCardNumberRequest) (*card.ApiResponseDashboardCardNumber, error) {
	year := time.Now().Year()
	balanceData, _ := h.repo.GetYearlyAmounts(ctx, "saldo_events", "card_number", req.CardNumber, year, year)
	topupData, _ := h.repo.GetYearlyAmounts(ctx, "topup_events", "card_number", req.CardNumber, year, year)
	withdrawData, _ := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "card_number", req.CardNumber, year, year)
	transactionData, _ := h.repo.GetYearlyAmounts(ctx, "transaction_events", "card_number", req.CardNumber, year, year)
	transferSendData, _ := h.repo.GetYearlyAmounts(ctx, "transfer_events", "source_card", req.CardNumber, year, year)
	transferRecvData, _ := h.repo.GetYearlyAmounts(ctx, "transfer_events", "destination_card", req.CardNumber, year, year)

	totalBalance := int64(0)
	if len(balanceData) > 0 {
		totalBalance = balanceData[0].TotalAmount
	}
	totalTopup := int64(0)
	if len(topupData) > 0 {
		totalTopup = topupData[0].TotalAmount
	}
	totalWithdraw := int64(0)
	if len(withdrawData) > 0 {
		totalWithdraw = withdrawData[0].TotalAmount
	}
	totalTransaction := int64(0)
	if len(transactionData) > 0 {
		totalTransaction = transactionData[0].TotalAmount
	}
	totalTransferSend := int64(0)
	if len(transferSendData) > 0 {
		totalTransferSend = transferSendData[0].TotalAmount
	}
	totalTransferRecv := int64(0)
	if len(transferRecvData) > 0 {
		totalTransferRecv = transferRecvData[0].TotalAmount
	}

	return &card.ApiResponseDashboardCardNumber{
		Status:  "success",
		Message: "Retrieved card-specific dashboard",
		Data: &card.CardResponseDashboardCardNumber{
			TotalBalance:          totalBalance,
			TotalTopup:            totalTopup,
			TotalWithdraw:         totalWithdraw,
			TotalTransaction:      totalTransaction,
			TotalTransferSend:     totalTransferSend,
			TotalTransferReceiver: totalTransferRecv,
		},
	}, nil
}

// --- Card Stats Balance ---

func (h *CardStatsHandler) FindMonthlyBalance(ctx context.Context, req *pbCardStats.FindYearBalance) (*pbCardStats.ApiResponseMonthlyBalance, error) {
	type SaldoRepo interface {
		GetMonthlyTotalSaldo(ctx context.Context, year int) ([]repository.MonthlyAmount, error)
	}

	var data []repository.MonthlyAmount
	var err error
	if sr, ok := h.repo.(SaldoRepo); ok {
		data, err = sr.GetMonthlyTotalSaldo(ctx, int(req.Year))
	} else {
		data, err = h.repo.GetMonthlyAmounts(ctx, "saldo_events", "", nil, int(req.Year))
	}
	if err != nil {
		return nil, err
	}
	return &pbCardStats.ApiResponseMonthlyBalance{
		Status:  "success",
		Message: "Retrieved monthly balances",
		Data:    h.mapToCardMonthlyBalance(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyBalance(ctx context.Context, req *pbCardStats.FindYearBalance) (*pbCardStats.ApiResponseYearlyBalance, error) {
	type SaldoRepo interface {
		GetYearlyTotalSaldo(ctx context.Context, startYear, endYear int) ([]repository.YearlyAmount, error)
	}

	var data []repository.YearlyAmount
	var err error
	if sr, ok := h.repo.(SaldoRepo); ok {
		data, err = sr.GetYearlyTotalSaldo(ctx, int(req.Year), int(req.Year))
	} else {
		data, err = h.repo.GetYearlyAmounts(ctx, "saldo_events", "", nil, int(req.Year), int(req.Year))
	}
	if err != nil {
		return nil, err
	}
	return &pbCardStats.ApiResponseYearlyBalance{
		Status:  "success",
		Message: "Retrieved yearly balances",
		Data:    h.mapToCardYearlyBalance(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyBalanceByCardNumber(ctx context.Context, req *pbCardStats.FindYearBalanceCardNumber) (*pbCardStats.ApiResponseMonthlyBalance, error) {
	type SaldoRepo interface {
		GetMonthlyAmounts(ctx context.Context, table string, filterField string, filterValue interface{}, year int) ([]repository.MonthlyAmount, error)
	}

	var data []repository.MonthlyAmount
	var err error
	if sr, ok := h.repo.(SaldoRepo); ok {
		data, err = sr.GetMonthlyAmounts(ctx, "saldo_events", "card_number", req.CardNumber, int(req.Year))
	} else {
		data, err = h.repo.GetMonthlyAmounts(ctx, "saldo_events", "card_number", req.CardNumber, int(req.Year))
	}
	if err != nil {
		return nil, err
	}
	return &pbCardStats.ApiResponseMonthlyBalance{
		Status:  "success",
		Message: "Retrieved monthly balances for card",
		Data:    h.mapToCardMonthlyBalance(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyBalanceByCardNumber(ctx context.Context, req *pbCardStats.FindYearBalanceCardNumber) (*pbCardStats.ApiResponseYearlyBalance, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "saldo_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &pbCardStats.ApiResponseYearlyBalance{
		Status:  "success",
		Message: "Retrieved yearly balances for card",
		Data:    h.mapToCardYearlyBalance(data),
	}, nil
}

// --- Card Stats Topup ---

func (h *CardStatsHandler) FindMonthlyTopupAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "topup_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly topups",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTopupAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "topup_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly topups",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyTopupAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly topups for card",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTopupAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "topup_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly topups for card",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

// --- Card Stats Transaction ---

func (h *CardStatsHandler) FindMonthlyTransactionAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transactions",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransactionAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transactions",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyTransactionAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transactions for card",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransactionAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transaction_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transactions for card",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

// --- Card Stats Transfer ---

func (h *CardStatsHandler) FindMonthlyTransferSenderAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer sender amounts",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransferSenderAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer sender amounts",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyTransferReceiverAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer receiver amounts",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransferReceiverAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer receiver amounts",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyTransferSenderAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer sender amounts for card",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransferSenderAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "source_card", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer sender amounts for card",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyTransferReceiverAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "transfer_events", "destination_card", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly transfer receiver amounts for card",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyTransferReceiverAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "transfer_events", "destination_card", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly transfer receiver amounts for card",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

// --- Card Stats Withdraw ---

func (h *CardStatsHandler) FindMonthlyWithdrawAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "withdraw_events", "", nil, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly withdraws",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyWithdrawAmount(ctx context.Context, req *card.FindYearAmount) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "", nil, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly withdraws",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindMonthlyWithdrawAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseMonthlyAmount, error) {
	data, err := h.repo.GetMonthlyAmounts(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseMonthlyAmount{
		Status:  "success",
		Message: "Retrieved monthly withdraws for card",
		Data:    h.mapToCardMonthlyAmount(data),
	}, nil
}

func (h *CardStatsHandler) FindYearlyWithdrawAmountByCardNumber(ctx context.Context, req *card.FindYearAmountCardNumber) (*card.ApiResponseYearlyAmount, error) {
	data, err := h.repo.GetYearlyAmounts(ctx, "withdraw_events", "card_number", req.CardNumber, int(req.Year), int(req.Year))
	if err != nil {
		return nil, err
	}
	return &card.ApiResponseYearlyAmount{
		Status:  "success",
		Message: "Retrieved yearly withdraws for card",
		Data:    h.mapToCardYearlyAmount(data),
	}, nil
}

// --- Mappers ---

func (h *CardStatsHandler) mapToCardMonthlyBalance(data []repository.MonthlyAmount) []*pbCardStats.CardResponseMonthlyBalance {
	var results []*pbCardStats.CardResponseMonthlyBalance
	for _, d := range data {
		results = append(results, &pbCardStats.CardResponseMonthlyBalance{
			Month:        d.Month,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}

func (h *CardStatsHandler) mapToCardYearlyBalance(data []repository.YearlyAmount) []*pbCardStats.CardResponseYearlyBalance {
	var results []*pbCardStats.CardResponseYearlyBalance
	for _, d := range data {
		results = append(results, &pbCardStats.CardResponseYearlyBalance{
			Year:         d.Year,
			TotalBalance: d.TotalAmount,
		})
	}
	return results
}

func (h *CardStatsHandler) mapToCardMonthlyAmount(data []repository.MonthlyAmount) []*card.CardResponseMonthlyAmount {
	var results []*card.CardResponseMonthlyAmount
	for _, d := range data {
		results = append(results, &card.CardResponseMonthlyAmount{
			Month:       d.Month,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}

func (h *CardStatsHandler) mapToCardYearlyAmount(data []repository.YearlyAmount) []*card.CardResponseYearlyAmount {
	var results []*card.CardResponseYearlyAmount
	for _, d := range data {
		results = append(results, &card.CardResponseYearlyAmount{
			Year:        d.Year,
			TotalAmount: d.TotalAmount,
		})
	}
	return results
}
