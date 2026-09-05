package rolegrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var ErrGrpcRoleInvalidId = errors.NewGrpcError("Invalid Role ID", int(codes.NotFound))
var ErrGrpcRoleInvalidName = errors.NewGrpcError("Invalid Role Name", int(codes.InvalidArgument))
