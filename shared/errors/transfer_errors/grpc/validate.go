package transfergrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateTransferRequest = errors.NewGrpcError("Invalid input for create transfer", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateTransferRequest = errors.NewGrpcError("Invalid input for update transfer", int(codes.InvalidArgument))
)
