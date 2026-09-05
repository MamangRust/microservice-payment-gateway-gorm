package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo/service"
)

type handler struct {
	SaldoQueryHandleGrpc
	SaldoCommandHandleGrpc
}

type Handler interface {
	SaldoQueryHandleGrpc
	SaldoCommandHandleGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		SaldoQueryHandleGrpc:   NewSaldoQueryHandleGrpc(service),
		SaldoCommandHandleGrpc: NewSaldoCommandHandleGrpc(service),
	}
}
