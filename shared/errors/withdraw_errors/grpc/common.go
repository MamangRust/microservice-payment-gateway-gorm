package withdrawgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcWithdrawInvalidID = errors.NewGrpcError("Invalid Withdraw ID", int(codes.InvalidArgument))
	ErrGrpcInvalidUserID     = errors.NewGrpcError("Invalid user ID", int(codes.InvalidArgument))
	ErrGrpcInvalidCardNumber = errors.NewGrpcError("Invalid card number", int(codes.InvalidArgument))
	ErrGrpcInvalidMonth      = errors.NewGrpcError("Invalid month", int(codes.InvalidArgument))
	ErrGrpcInvalidYear       = errors.NewGrpcError("Invalid year", int(codes.InvalidArgument))
)
