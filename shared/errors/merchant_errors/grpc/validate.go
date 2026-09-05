package merchantgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateMerchant       = errors.NewGrpcError("Invalid input for create merchant", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateMerchant       = errors.NewGrpcError("Invalid input for update merchant", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateMerchantStatus = errors.NewGrpcError("Invalid input for update merchant status", int(codes.InvalidArgument))
)
