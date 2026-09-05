package handler

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant/service"
)

type Handler interface {
	MerchantQueryHandleGrpc
	MerchantCommandHandleGrpc
	MerchantDocumentQueryHandleGrpc
	MerchantDocumentCommandHandleGrpc
	MerchantTransactionHandleGrpc
}

// Handler contains the gRPC handlers for merchant and merchant document operations.
type handler struct {
	MerchantQueryHandleGrpc
	MerchantCommandHandleGrpc
	MerchantDocumentQueryHandleGrpc
	MerchantDocumentCommandHandleGrpc
	MerchantTransactionHandleGrpc
}

func NewHandler(service service.Service) Handler {
	return &handler{
		MerchantQueryHandleGrpc:           NewMerchantQueryHandleGrpc(service.MerchantQueryService()),
		MerchantCommandHandleGrpc:         NewMerchantCommandHandleGrpc(service.MerchantCommandService()),
		MerchantDocumentQueryHandleGrpc:   NewMerchantDocumentQueryHandleGrpc(service.MerchantDocumentQueryService()),
		MerchantDocumentCommandHandleGrpc: NewMerchantDocumentCommandHandleGrpc(service.MerchantDocumentCommandService()),
		MerchantTransactionHandleGrpc: NewMerchantTransactionHandleGrpc(
			service.MerchantTransactionService()),
	}
}
