package service

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type SaldoQueryService interface {
	ListReconciliationQueue(ctx context.Context, status string, limit int32) ([]*repository.ReconciliationQueueResult, error)
	ListLedgerEntries(ctx context.Context, cardNumber string, limit int32) ([]*repository.LedgerEntryResult, error)
	FindAll(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error)
	FindById(ctx context.Context, saldo_id int) (*repository.SaldoResult, error)
	FindByCardNumber(ctx context.Context, card_number string) (*models.Saldo, error)
	FindByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, error)
}

type SaldoCommandService interface {
	CreateSaldo(ctx context.Context, request *requests.CreateSaldoRequest) (*repository.SaldoMutationResult, error)
	CreateSaldoIfNotExists(ctx context.Context, request *requests.CreateSaldoRequest) error
	UpdateSaldo(ctx context.Context, request *requests.UpdateSaldoRequest) (*repository.SaldoMutationResult, error)
	UpdateSaldoWithdraw(ctx context.Context, request *requests.UpdateSaldoWithdraw) (*repository.SaldoMutationResult, error)
	ApplySaldoAdjustment(ctx context.Context, request *requests.ApplySaldoAdjustmentRequest) (*repository.SaldoAdjustmentResult, error)
	ResolveReconciliation(ctx context.Context, queueID int64, operationID, note string) error
	DebitSaldo(ctx context.Context, request *requests.DebitSaldoRequest) (*repository.SaldoMutationResult, error)
	CreditSaldo(ctx context.Context, request *requests.CreditSaldoRequest) (*repository.SaldoMutationResult, error)
	TrashSaldo(ctx context.Context, saldo_id int) (*models.Saldo, error)
	RestoreSaldo(ctx context.Context, saldo_id int) (*models.Saldo, error)
	DeleteSaldoPermanent(ctx context.Context, saldo_id int) (bool, error)
	RestoreAllSaldo(ctx context.Context) (bool, error)
	DeleteAllSaldoPermanent(ctx context.Context) (bool, error)
}
