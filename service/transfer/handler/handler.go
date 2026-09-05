package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer/service"
)

type Handler interface {
	TransferQueryHandleGrpc
	TransferCommandHandleGrpc
}

type handler struct {
	TransferQueryHandleGrpc
	TransferCommandHandleGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		TransferQueryHandleGrpc:   NewTransferQueryHandler(service),
		TransferCommandHandleGrpc: NewTransferCommandHandler(service),
	}
}
