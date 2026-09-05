package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw/service"
)

type Handler interface {
	WithdrawQueryHandlerGrpc
	WithdrawCommandHandlerGrpc
}

type handler struct {
	WithdrawQueryHandlerGrpc
	WithdrawCommandHandlerGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		WithdrawQueryHandlerGrpc:   NewWithdrawQueryHandleGrpc(service),
		WithdrawCommandHandlerGrpc: NewWithdrawCommandHandleGrpc(service),
	}
}
