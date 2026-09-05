package cardhandler

import (
	"net/http"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	card_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/card"
	errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/card"
	"github.com/labstack/echo/v4"
)

type cardDashboardHandleApi struct {
	logger logger.LoggerInterface
	mapper apimapper.CardDashboardResponseMapper
	cache  card_cache.CardMencache

	apiHandler errors.ApiHandler
}

type cardDashboardHandleApiDeps struct {
	router     *echo.Echo
	logger     logger.LoggerInterface
	mapper     apimapper.CardDashboardResponseMapper
	cache      card_cache.CardMencache
	apiHandler errors.ApiHandler
}

func NewCardDashboardHandleApi(params *cardDashboardHandleApiDeps) *cardDashboardHandleApi {
	h := &cardDashboardHandleApi{
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	// Dashboard endpoints are now handled by StatsHandler's gRPC client to stats-reader.
	// This domain handler exists to maintain apigateway structure and registry compatibility.

	return h
}

// FindDashboard godoc
// @Summary Retrieve card dashboard (analytical)
// @Description This endpoint is deprecated in favor of /api/stats/card/dashboard
// @Tags Card Query
// @Security Bearer
// @Produce json
// @Success 200 {object} response.ApiResponseDashboardCard
// @Router /api/card-query/dashboard [get]
func (h *cardDashboardHandleApi) FindDashboard(c echo.Context) error {
	return h.apiHandler.Handle("find-dashboard", func(c echo.Context) error {
		return c.JSON(http.StatusMovedPermanently, map[string]string{
			"message": "Please use /api/stats/card/dashboard",
		})
	})(c)
}
