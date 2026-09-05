package merchantdocumentgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcFailedCreateMerchantDocument = errors.NewGrpcError("Failed to create merchant document", int(codes.Internal))
	ErrGrpcFailedUpdateMerchantDocument = errors.NewGrpcError("Failed to update merchant document", int(codes.Internal))
)
