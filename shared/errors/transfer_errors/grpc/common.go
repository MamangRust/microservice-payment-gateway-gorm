package transfergrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcTransferNotFound  = errors.NewGrpcError("Transfer not found", int(codes.NotFound))
	ErrGrpcTransferInvalidID = errors.NewGrpcError("Invalid Transfer ID", int(codes.InvalidArgument))
	ErrGrpcInvalidUserID     = errors.NewGrpcError("Invalid user ID", int(codes.InvalidArgument))
	ErrGrpcInvalidCardNumber = errors.NewGrpcError("Invalid card number", int(codes.InvalidArgument))
	ErrGrpcInvalidMonth      = errors.NewGrpcError("Invalid month", int(codes.InvalidArgument))
	ErrGrpcInvalidYear       = errors.NewGrpcError("Invalid year", int(codes.InvalidArgument))
)
