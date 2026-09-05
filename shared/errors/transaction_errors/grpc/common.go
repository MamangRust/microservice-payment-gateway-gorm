package transactiongrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcTransactionInvalidID         = errors.NewGrpcError("Invalid Transaction ID", int(codes.InvalidArgument))
	ErrGrpcTransactionInvalidMerchantID = errors.NewGrpcError("Invalid Transaction Merchant ID", int(codes.InvalidArgument))
	ErrGrpcInvalidCardNumber            = errors.NewGrpcError("Invalid card number", int(codes.InvalidArgument))
	ErrGrpcInvalidMonth                 = errors.NewGrpcError("Invalid month", int(codes.InvalidArgument))
	ErrGrpcInvalidYear                  = errors.NewGrpcError("Invalid year", int(codes.InvalidArgument))
)
