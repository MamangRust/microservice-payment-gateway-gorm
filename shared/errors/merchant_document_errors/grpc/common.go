package merchantdocumentgrpcerrors

import (
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"google.golang.org/grpc/codes"
)

var ErrGrpcMerchantInvalidID = errors.NewGrpcError("Invalid merchant id", int(codes.InvalidArgument))
