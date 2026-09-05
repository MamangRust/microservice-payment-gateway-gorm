package handler

import (
	"context"
	"math"
	"time"

	pbutils "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"

	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	user_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/user_errors/grpc"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/convert"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type userQueryHandleGrpc struct {
	pb.UnimplementedUserQueryServiceServer

	userQueryService service.UserQueryService
}

func NewUserQueryHandleGrpc(query service.UserQueryService) UserQueryHandleGrpc {
	return &userQueryHandleGrpc{
		userQueryService: query,
	}
}


func (s *userQueryHandleGrpc) FindAll(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUser, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindAll(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
		}
	}

	return &pb.ApiResponsePaginationUser{
		Status:         "success",
		Message:        "Successfully fetched users",
		Data:           userResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *userQueryHandleGrpc) FindById(ctx context.Context, request *pb.FindByIdUserRequest) (*pb.ApiResponseUser, error) {
	id := int(request.GetId())

	if id == 0 {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.userQueryService.FindByID(ctx, id)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) FindByActive(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUserDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindByActive(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponseDeleteAt, len(users))
	for i, user := range users {
		var deletedAt *wrapperspb.StringValue
		if user.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: user.DeletedAt.Format(time.RFC3339)}
		}
		userResponses[i] = &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
			DeletedAt: deletedAt,
		}
	}

	return &pb.ApiResponsePaginationUserDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched active users",
		Data:           userResponses,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *userQueryHandleGrpc) FindByTrashed(ctx context.Context, request *pb.FindAllUserRequest) (*pb.ApiResponsePaginationUserDeleteAt, error) {
	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllUsers{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	users, totalRecords, err := s.userQueryService.FindByTrashed(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	userResponses := make([]*pb.UserResponseDeleteAt, len(users))
	for i, user := range users {
		var deletedAt *wrapperspb.StringValue
		if user.DeletedAt != nil {
			deletedAt = &wrapperspb.StringValue{Value: user.DeletedAt.Format(time.RFC3339)}
		}
		userResponses[i] = &pb.UserResponseDeleteAt{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
			DeletedAt: deletedAt,
		}
	}

	return &pb.ApiResponsePaginationUserDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched trashed users",
		Data:           userResponses,
		PaginationMeta: paginationMeta,
	}, nil
}
func (s *userQueryHandleGrpc) FindByEmail(ctx context.Context, request *pb.FindByEmailUserRequest) (*pb.ApiResponseUser, error) {
	email := request.GetEmail()

	if email == "" {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.userQueryService.FindByEmail(ctx, email)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
		},
	}, nil
}

func (s *userQueryHandleGrpc) FindByVerificationCode(ctx context.Context, request *pb.FindByVerificationCodeUserRequest) (*pb.ApiResponseUser, error) {
	code := request.GetVerificationCode()

	if code == "" {
		return nil, user_errors.ErrGrpcUserInvalidId
	}

	user, err := s.userQueryService.FindByVerificationCode(ctx, code)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pb.ApiResponseUser{
		Status:  "success",
		Message: "Successfully fetched user",
		Data: &pb.UserResponse{
			Id:        user.UserID,
			Firstname: user.Firstname,
			Lastname:  user.Lastname,
			Email:     user.Email,
			Password:  user.Password,
			CreatedAt: convert.FormatTimeRFC3339(user.CreatedAt),
			UpdatedAt: convert.FormatTimeRFC3339(user.UpdatedAt),
		},
	}, nil
}
