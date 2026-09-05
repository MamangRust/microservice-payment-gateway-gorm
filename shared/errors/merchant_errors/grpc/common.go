package merchantgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcMerchantInvalidID     = errors.NewGrpcError("Invalid Merchant ID", int(codes.InvalidArgument))
	ErrGrpcMerchantInvalidUserID = errors.NewGrpcError("Invalid Merchant User ID", int(codes.InvalidArgument))
	ErrGrpcMerchantInvalidApiKey = errors.NewGrpcError("Invalid Merchant Api Key", int(codes.InvalidArgument))
	ErrGrpcMerchantInvalidMonth  = errors.NewGrpcError("Invalid Merchant Month", int(codes.InvalidArgument))
	ErrGrpcMerchantInvalidYear   = errors.NewGrpcError("Invalid Merchant Year", int(codes.InvalidArgument))
)
