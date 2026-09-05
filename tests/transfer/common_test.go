package transfer_test

import (
	"context"

	card_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/card/repository"
	saldo_repo "github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type transferSaldoRepoAdapter struct {
	saldoRepo saldo_repo.Repositories
}

func (a *transferSaldoRepoAdapter) FindByCardNumber(ctx context.Context, cardNumber string) (*models.Saldo, error) {
	return a.saldoRepo.FindByCardNumber(ctx, cardNumber)
}

func (a *transferSaldoRepoAdapter) UpdateSaldoBalance(ctx context.Context, req *requests.UpdateSaldoBalance) (*saldo_repo.SaldoMutationResult, error) {
	return a.saldoRepo.UpdateSaldoBalance(ctx, req)
}

type transferCardRepoAdapter struct {
	cardRepo card_repo.Repositories
}

func (a *transferCardRepoAdapter) FindUserCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error) {
	return a.cardRepo.CardQuery.FindUserCardByCardNumber(ctx, cardNumber)
}

func (a *transferCardRepoAdapter) FindCardByCardNumber(ctx context.Context, cardNumber string) (*models.Card, error) {
	return a.cardRepo.CardQuery.FindCardByCardNumber(ctx, cardNumber)
}
