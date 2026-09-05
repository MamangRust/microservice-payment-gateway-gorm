package rolegrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var (
	ErrGrpcValidateCreateRole = errors.NewGrpcError("validation failed: invalid create Role request", int(codes.InvalidArgument))
	ErrGrpcValidateUpdateRole = errors.NewGrpcError("validation failed: invalid update Role request", int(codes.InvalidArgument))
)
