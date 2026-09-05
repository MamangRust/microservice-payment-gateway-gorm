package topuphandler

import (
	"fmt"
	"net/http"
	"strconv"

	"google.golang.org/grpc/metadata"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup"
	topup_cache "github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway/redis/api/topup"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/idempotency"
	apimapper "github.com/MamangRust/microservice-payment-gateway-grpc/shared/mapper/topup"
	"github.com/go-playground/validator/v10"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/labstack/echo/v4"
)

type topupCommandHandleApi struct {
	client pb.TopupCommandServiceClient

	logger logger.LoggerInterface

	mapper apimapper.TopupCommandResponseMapper

	cache topup_cache.TopupMencache

	apiHandler errors.ApiHandler
}

type topupCommandHandleDeps struct {
	client pb.TopupCommandServiceClient

	router *echo.Echo

	logger logger.LoggerInterface

	mapper apimapper.TopupCommandResponseMapper

	cache topup_cache.TopupMencache

	apiHandler errors.ApiHandler
}

func NewTopupCommandHandleApi(params *topupCommandHandleDeps) *topupCommandHandleApi {

	topupHandler := &topupCommandHandleApi{
		client:     params.client,
		logger:     params.logger,
		mapper:     params.mapper,
		cache:      params.cache,
		apiHandler: params.apiHandler,
	}

	routerTopup := params.router.Group("/api/topup-command")

	routerTopup.POST("/create", params.apiHandler.Handle("create-topup", topupHandler.Create))
	routerTopup.POST("/update/:id", params.apiHandler.Handle("update-topup", topupHandler.Update))
	routerTopup.POST("/trashed/:id", params.apiHandler.Handle("trash-topup", topupHandler.TrashTopup))
	routerTopup.POST("/restore/:id", params.apiHandler.Handle("restore-topup", topupHandler.RestoreTopup))
	routerTopup.DELETE("/permanent/:id", params.apiHandler.Handle("delete-topup-permanent", topupHandler.DeleteTopupPermanent))

	routerTopup.POST("/restore/all", params.apiHandler.Handle("restore-all-topups", topupHandler.RestoreAllTopup))
	routerTopup.POST("/permanent/all", params.apiHandler.Handle("delete-all-topups-permanent", topupHandler.DeleteAllTopupPermanent))

	return topupHandler
}

// @Summary Create topup
// @Tags Topup Command
// @Security Bearer
// @Description Create a new topup record
// @Accept json
// @Produce json
// @Param CreateTopupRequest body requests.CreateTopupRequest true "Create topup request"
// @Success 200 {object} response.ApiResponseTopup "Created topup data"
// @Failure 400 {object} response.ErrorResponse "Bad Request: "
// @Failure 500 {object} response.ErrorResponse "Failed to create topup: "
// @Router /api/topup-command/create [post]
func (h *topupCommandHandleApi) Create(c echo.Context) error {
	var body requests.CreateTopupRequest

	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("Invalid request")
	}

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return errors.NewValidationError(validations)
	}

	key := c.Request().Header.Get("Idempotency-Key")
	if err := idempotency.ValidateKey(key); err != nil {
		return err
	}

	ctx := c.Request().Context()
	userID, ok := c.Get("user_id").(string)
	if !ok || userID == "" {
		return errors.ErrUnauthorized
	}
	ctx = metadata.AppendToOutgoingContext(ctx, "x-user-id", userID)

	res, err := h.client.CreateTopup(ctx, &pb.CreateTopupRequest{
		CardNumber:     body.CardNumber,
		TopupAmount:    int64(body.TopupAmount),
		TopupMethod:    body.TopupMethod,
		IdempotencyKey: key,
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseTopup(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Update topup
// @Tags Topup Command
// @Security Bearer
// @Description Update an existing topup record with the provided details
// @Accept json
// @Produce json
// @Param id path int true "Topup ID"
// @Param UpdateTopupRequest body requests.UpdateTopupRequest true "Update topup request"
// @Success 200 {object} response.ApiResponseTopup "Updated topup data"
// @Failure 400 {object} response.ErrorResponse "Bad Request: Invalid input data"
// @Failure 500 {object} response.ErrorResponse "Failed to update topup: "
// @Router /api/topup-command/update/{id} [post]
func (h *topupCommandHandleApi) Update(c echo.Context) error {
	idint, err := strconv.Atoi(c.Param("id"))

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	var body requests.UpdateTopupRequest

	if err := c.Bind(&body); err != nil {
		return errors.NewBadRequestError("Invalid request")
	}

	body.TopupID = &idint

	if err := body.Validate(); err != nil {
		validations := h.parseValidationErrors(err)
		return errors.NewValidationError(validations)
	}

	ctx := c.Request().Context()

	res, err := h.client.UpdateTopup(ctx, &pb.UpdateTopupRequest{
		TopupId:     int32(idint),
		CardNumber:  body.CardNumber,
		TopupAmount: int64(body.TopupAmount),
		TopupMethod: body.TopupMethod,
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseTopup(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Trash a topup
// @Tags Topup Command
// @Security Bearer
// @Description Trash a topup record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "Topup ID"
// @Success 200 {object} response.ApiResponseTopup "Successfully trashed topup record"
// @Failure 400 {object} response.ErrorResponse "Bad Request: Invalid ID"
// @Failure 500 {object} response.ErrorResponse "Failed to trash topup:"
// @Router /api/topup-command/trash/{id} [post]
func (h *topupCommandHandleApi) TrashTopup(c echo.Context) error {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := c.Request().Context()

	res, err := h.client.TrashedTopup(ctx, &pb.FindByIdTopupRequest{
		TopupId: int32(idInt),
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseTopupDeleteAt(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Restore a trashed topup
// @Tags Topup Command
// @Security Bearer
// @Description Restore a trashed topup record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "Topup ID"
// @Success 200 {object} response.ApiResponseTopup "Successfully restored topup record"
// @Failure 400 {object} response.ErrorResponse "Bad Request: Invalid ID"
// @Failure 500 {object} response.ErrorResponse "Failed to restore topup:"
// @Router /api/topup-command/restore/{id} [post]
func (h *topupCommandHandleApi) RestoreTopup(c echo.Context) error {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := c.Request().Context()

	res, err := h.client.RestoreTopup(ctx, &pb.FindByIdTopupRequest{
		TopupId: int32(idInt),
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseTopupDeleteAt(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Permanently delete a topup
// @Tags Topup Command
// @Security Bearer
// @Description Permanently delete a topup record by its ID.
// @Accept json
// @Produce json
// @Param id path int true "Topup ID"
// @Success 200 {object} response.ApiResponseTopupDelete "Successfully deleted topup record permanently"
// @Failure 400 {object} response.ErrorResponse "Bad Request: Invalid ID"
// @Failure 500 {object} response.ErrorResponse "Failed to delete topup:"
// @Router /api/topup-command/permanent/{id} [delete]
func (h *topupCommandHandleApi) DeleteTopupPermanent(c echo.Context) error {
	id := c.Param("id")

	idInt, err := strconv.Atoi(id)

	if err != nil {
		return errors.NewBadRequestError("id is required")
	}

	ctx := c.Request().Context()

	res, err := h.client.DeleteTopupPermanent(ctx, &pb.FindByIdTopupRequest{
		TopupId: int32(idInt),
	})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	so := h.mapper.ToApiResponseTopupDelete(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Restore all topup records
// @Tags Topup Command
// @Security Bearer
// @Description Restore all topup records that were previously deleted.
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTopupAll "Successfully restored all topup records"
// @Failure 500 {object} response.ErrorResponse "Failed to restore all topup records"
// @Router /api/topup-command/restore/all [post]
func (h *topupCommandHandleApi) RestoreAllTopup(c echo.Context) error {
	ctx := c.Request().Context()

	res, err := h.client.RestoreAllTopup(ctx, &emptypb.Empty{})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.logger.Debug("Successfully restored all topup")

	so := h.mapper.ToApiResponseTopupAll(res)

	return c.JSON(http.StatusOK, so)
}

// @Summary Permanently delete all topup records
// @Tags Topup Command
// @Security Bearer
// @Description Permanently delete all topup records from the database.
// @Accept json
// @Produce json
// @Success 200 {object} response.ApiResponseTopupAll "Successfully deleted all topup records permanently"
// @Failure 500 {object} response.ErrorResponse "Failed to permanently delete all topup records"
// @Router /api/topup-command/permanent/all [post]
func (h *topupCommandHandleApi) DeleteAllTopupPermanent(c echo.Context) error {
	ctx := c.Request().Context()

	res, err := h.client.DeleteAllTopupPermanent(ctx, &emptypb.Empty{})

	if err != nil {
		return errors.ParseGrpcError(err)
	}

	h.logger.Debug("Successfully deleted all topup permanently")

	so := h.mapper.ToApiResponseTopupAll(res)

	return c.JSON(http.StatusOK, so)
}

func (h *topupCommandHandleApi) parseValidationErrors(err error) []errors.ValidationError {
	var validationErrs []errors.ValidationError

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			validationErrs = append(validationErrs, errors.ValidationError{
				Field:   fe.Field(),
				Message: h.getValidationMessage(fe),
			})
		}
		return validationErrs
	}

	return []errors.ValidationError{
		{
			Field:   "general",
			Message: err.Error(),
		},
	}
}

func (h *topupCommandHandleApi) getValidationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return fmt.Sprintf("Must be at least %s", fe.Param())
	case "max":
		return fmt.Sprintf("Must be at most %s", fe.Param())
	case "gte":
		return fmt.Sprintf("Must be greater than or equal to %s", fe.Param())
	case "lte":
		return fmt.Sprintf("Must be less than or equal to %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("Must be one of: %s", fe.Param())
	default:
		return fmt.Sprintf("Validation failed on '%s' tag", fe.Tag())
	}
}
