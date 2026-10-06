package adapter

import (
	"context"
	"time"

	pbcard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbsaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
)

type SaldoAdapter interface {
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.SaldoMutationResult, error)
	DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*models.SaldoMutationResult, error)
	CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*models.SaldoMutationResult, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.SaldoMutationResult, error)
}

type saldoGRPCAdapter struct {
	QueryClient   pbsaldo.SaldoQueryServiceClient
	CommandClient pbsaldo.SaldoCommandServiceClient
	guard         *resilience.DependencyGuard
}

func (a *saldoGRPCAdapter) SetGuard(g *resilience.DependencyGuard) {
	a.guard = g
}

func NewSaldoAdapter(queryClient pbsaldo.SaldoQueryServiceClient, commandClient pbsaldo.SaldoCommandServiceClient, opts ...GuardOption) SaldoAdapter {
	a := &saldoGRPCAdapter{
		QueryClient:   queryClient,
		CommandClient: commandClient,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

func (a *saldoGRPCAdapter) FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.QueryClient.FindByCardNumber(callCtx, &pbcard.FindByCardNumberRequest{
			CardNumber: card_number,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}

	return mapSaldoResponse(resp.Data)
}

func mapSaldoResponse(s *pbsaldo.SaldoResponse) (*models.Saldo, error) {
	if s == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	saldo := &models.Saldo{
		SaldoID:      s.SaldoId,
		CardNumber:   s.CardNumber,
		TotalBalance: s.TotalBalance,
	}
	if s.WithdrawAmount != 0 {
		amount := s.WithdrawAmount
		saldo.WithdrawAmount = &amount
	}
	saldo.WithdrawTime = parseTime(s.WithdrawTime)
	saldo.CreatedAt = parseTime(s.CreatedAt)
	saldo.UpdatedAt = parseTime(s.UpdatedAt)
	return saldo, nil
}

func mapMutationResponse(resp *pbsaldo.SaldoResponse) *models.SaldoMutationResult {
	if resp == nil {
		return nil
	}
	return &models.SaldoMutationResult{
		SaldoID:      resp.SaldoId,
		CardNumber:   resp.CardNumber,
		TotalBalance: resp.TotalBalance,
	}
}

func (a *saldoGRPCAdapter) UpdateSaldoBalance(ctx context.Context, request *requests.UpdateSaldoBalance) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldo(callCtx, &pbsaldo.UpdateSaldoRequest{
			CardNumber:   request.CardNumber,
			TotalBalance: int64(request.TotalBalance),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMutationResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.DebitSaldo(callCtx, &pbsaldo.DebitSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int64(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.CreditSaldo(callCtx, &pbsaldo.CreditSaldoRequest{
			CardNumber:  request.CardNumber,
			Amount:      int64(request.Amount),
			OperationId: request.OperationID,
			SourceType:  request.SourceType,
			SourceId:    request.SourceID,
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, sharedErrors.NewBadRequestError("saldo response is required")
	}
	return mapMutationResponse(resp.Data), nil
}

func (a *saldoGRPCAdapter) UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*models.SaldoMutationResult, error) {
	var resp *pbsaldo.ApiResponseSaldo
	err := callGuarded(ctx, a.guard, func(callCtx context.Context) error {
		var callErr error
		resp, callErr = a.CommandClient.UpdateSaldoWithdraw(callCtx, &pbsaldo.UpdateSaldoWithdrawRequest{
			CardNumber:     request.CardNumber,
			TotalBalance:   int64(request.TotalBalance),
			WithdrawAmount: int64(*request.WithdrawAmount),
			WithdrawTime:   request.WithdrawTime.Format(time.RFC3339),
		})
		return callErr
	})
	if err != nil {
		return nil, err
	}
	return mapMutationResponse(resp.Data), nil
}
