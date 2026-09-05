package authhandler

import (
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	auth_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/auth"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/cache"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	authapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/auth"

	"github.com/labstack/echo/v4"
	"google.golang.org/grpc"
)

type DepsAuth struct {
	Client     *grpc.ClientConn
	E          *echo.Echo
	Logger     logger.LoggerInterface
	Cache      *cache.CacheStore
	ApiHandler errors.ApiHandler
}

func RegisterAuthHandler(deps *DepsAuth) {
	mapper := authapimapper.NewAuthResponseMapper()

	cache := auth_cache.NewMencache(deps.Cache)

	NewHandlerAuth(&authHandleParams{
		client:     pb.NewAuthServiceClient(deps.Client),
		router:     deps.E,
		logger:     deps.Logger,
		mapper:     mapper,
		cache:      cache,
		apiHandler: deps.ApiHandler,
	})
}
