package saldohandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo"
	pbSaldoStats "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	stats_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	saldoapimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/saldo"
	"github.com/labstack/echo/v4"
)

type saldoStatsHandleApi struct {
	logger logger.LoggerInterface
	cache  stats_cache.StatsCache

	apiHandler errors.ApiHandler

	// Saldo Clients
	saldoBalance pbSaldoStats.SaldoStatsBalanceServiceClient
	saldoTotal   pbSaldoStats.SaldoStatsTotalBalanceClient
}

func setupSaldoStatsHandler(deps *DepsSaldo) func() {
	return func() {
		statsCache := stats_cache.NewStatsCache(deps.Cache)
		h := &saldoStatsHandleApi{
			logger:     deps.Logger,
			cache:      statsCache,
			apiHandler: deps.ApiHandler,

			saldoBalance: pbSaldoStats.NewSaldoStatsBalanceServiceClient(deps.StatsClient),
			saldoTotal:   pbSaldoStats.NewSaldoStatsTotalBalanceClient(deps.StatsClient),
		}

		g := deps.E.Group("/api/saldo/stats")

		g.GET("/balance/monthly", h.apiHandler.Handle("find-monthly-saldo-balances", h.FindMonthlySaldoBalances))
		g.GET("/balance/yearly", h.apiHandler.Handle("find-yearly-saldo-balances", h.FindYearlySaldoBalances))
		g.GET("/total/monthly", h.apiHandler.Handle("find-monthly-total-saldo-balance", h.FindMonthlyTotalSaldoBalance))
		g.GET("/total/yearly", h.apiHandler.Handle("find-year-total-saldo-balance", h.FindYearTotalSaldoBalance))
	}
}

// FindMonthlySaldoBalances godoc
// @Summary Get monthly saldo balances
// @Tags Stats Saldo
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseMonthSaldoBalances
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/saldo/stats/balance/monthly [get]
func (h *saldoStatsHandleApi) FindMonthlySaldoBalances(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:saldo:balance:monthly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.saldoBalance.FindMonthlySaldoBalances(ctx, &saldo.FindYearlySaldo{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := saldoapimapper.NewSaldoStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseMonthSaldoBalances(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearlySaldoBalances godoc
// @Summary Get yearly saldo balances
// @Tags Stats Saldo
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearSaldoBalances
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/saldo/stats/balance/yearly [get]
func (h *saldoStatsHandleApi) FindYearlySaldoBalances(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:saldo:balance:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.saldoBalance.FindYearlySaldoBalances(ctx, &saldo.FindYearlySaldo{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := saldoapimapper.NewSaldoStatsBalanceResponseMapper()
	apiRes := mapper.ToApiResponseYearSaldoBalances(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindMonthlyTotalSaldoBalance godoc
// @Summary Get monthly total saldo balance
// @Tags Stats Saldo
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Param month query int true "Month"
// @Success 200 {object} response.ApiResponseMonthSaldoBalances
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/saldo/stats/total/monthly [get]
func (h *saldoStatsHandleApi) FindMonthlyTotalSaldoBalance(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	// month is not used in the GRPC call currently in stats/saldo.go but prefixing cache key with it
	month, _ := strconv.Atoi(c.QueryParam("month"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:saldo:total:monthly:%d:%d", year, month)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.saldoTotal.FindMonthlyTotalSaldoBalance(ctx, &saldo.FindMonthlySaldoTotalBalance{
		Year:  int32(year),
		Month: int32(month),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := saldoapimapper.NewSaldoStatsTotalResponseMapper()
	apiRes := mapper.ToApiResponseMonthTotalSaldo(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}

// FindYearTotalSaldoBalance godoc
// @Summary Get yearly total saldo balance
// @Tags Stats Saldo
// @Accept json
// @Produce json
// @Param year query int true "Year"
// @Success 200 {object} response.ApiResponseYearTotalSaldo
// @Failure 500 {object} errors.ErrorResponse
// @Router /api/saldo/stats/total/yearly [get]
func (h *saldoStatsHandleApi) FindYearTotalSaldoBalance(c echo.Context) error {
	year, _ := strconv.Atoi(c.QueryParam("year"))
	ctx := c.Request().Context()
	cacheKey := fmt.Sprintf("stats:saldo:total:yearly:%d", year)

	if cached, found := h.cache.GetCache(ctx, cacheKey); found {
		return c.JSON(http.StatusOK, cached)
	}

	resp, err := h.saldoTotal.FindYearTotalSaldoBalance(ctx, &saldo.FindYearlySaldo{
		Year: int32(year),
	})
	if err != nil {
		return errors.ParseGrpcError(err)
	}

	mapper := saldoapimapper.NewSaldoStatsTotalResponseMapper()
	apiRes := mapper.ToApiResponseYearTotalSaldo(resp)

	h.cache.SetCache(ctx, cacheKey, apiRes)
	return c.JSON(http.StatusOK, apiRes)
}
