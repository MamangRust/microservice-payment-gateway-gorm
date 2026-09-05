package repository

import (
	"context"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/events"
)

type Repository interface {
	InsertTransactionEvent(ctx context.Context, event events.TransactionEvent) error
	InsertTopupEvent(ctx context.Context, event events.TopupEvent) error
	InsertTransferEvent(ctx context.Context, event events.TransferEvent) error
	InsertWithdrawEvent(ctx context.Context, event events.WithdrawEvent) error
	InsertSaldoEvent(ctx context.Context, event events.SaldoEvent) error
	InsertMerchantEvent(ctx context.Context, event events.MerchantEvent) error
	InsertCardEvent(ctx context.Context, event events.CardEvent) error

	Flush(ctx context.Context) error
	Close() error
}
