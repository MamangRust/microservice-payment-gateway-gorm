package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction/service"
)

type Handler interface {
	TransactionQueryHandleGrpc
	TransactionCommandHandleGrpc
}

type handler struct {
	TransactionQueryHandleGrpc
	TransactionCommandHandleGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		TransactionQueryHandleGrpc:   NewTransactionQueryHandleGrpc(service),
		TransactionCommandHandleGrpc: NewTransactionCommandHandleGrpc(service),
	}
}
