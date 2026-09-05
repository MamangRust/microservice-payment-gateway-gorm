package mencache

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/database/models"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
)

type SaldoQueryCache interface {
	GetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldos(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
	GetCachedSaldoById(ctx context.Context, saldo_id int) (*repository.SaldoResult, bool)
	SetCachedSaldoById(ctx context.Context, saldo_id int, data *repository.SaldoResult)
	GetCachedSaldoByCardNumber(ctx context.Context, card_number string) (*models.Saldo, bool)
	SetCachedSaldoByCardNumber(ctx context.Context, card_number string, data *models.Saldo)
	GetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldoByActive(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
	GetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos) ([]*repository.SaldoResult, *int, bool)
	SetCachedSaldoByTrashed(ctx context.Context, req *requests.FindAllSaldos, data []*repository.SaldoResult, totalRecords *int)
}

type SaldoCommandCache interface {
	DeleteSaldoCache(ctx context.Context, saldo_id int)
	DeleteSaldoCacheByCardNumber(ctx context.Context, card_number string)
}
