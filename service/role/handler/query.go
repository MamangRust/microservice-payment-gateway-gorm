package handler

import (
	"context"
	"math"
	"time"

	pbutils "github.com/MamangRust/microservice-payment-gateway-grpc/pb/common"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/domain/requests"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	role_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/role_errors/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type roleQueryHandleGrpc struct {
	pb.UnimplementedRoleQueryServiceServer
	roleQuery service.RoleQueryService
}

func NewRoleQueryHandleGrpc(roleQuery service.RoleQueryService) RoleQueryHandlerGrpc {
	return &roleQueryHandleGrpc{
		roleQuery: roleQuery,
	}
}

func (s *roleQueryHandleGrpc) FindAllRole(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRole, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindAll(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb.RoleResponse, len(roles))
	for i, role := range roles {
		protoRoles[i] = &pb.RoleResponse{
			Id:        role.RoleID,
			Name:      role.RoleName,
			CreatedAt: formatCreatedAt(role.CreatedAt),
			UpdatedAt: formatUpdatedAt(role.UpdatedAt),
		}
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationRole{
		Status:         "success",
		Message:        "Successfully fetched role records",
		Data:           protoRoles,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *roleQueryHandleGrpc) FindByActive(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindByActiveRole(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb.RoleResponseDeleteAt, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		protoRoles = append(protoRoles, mapRoleResponseDeleteAtFromPtrs(
			role.RoleID,
			role.RoleName,
			role.CreatedAt,
			role.UpdatedAt,
			role.DeletedAt,
		))
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationRoleDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched active roles",
		Data:           protoRoles,
		PaginationMeta: paginationMeta,
	}, nil
}

func (s *roleQueryHandleGrpc) FindByTrashed(ctx context.Context, req *pb.FindAllRoleRequest) (*pb.ApiResponsePaginationRoleDeleteAt, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllRoles{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	roles, totalRecords, err := s.roleQuery.FindByTrashedRole(ctx, &reqService)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pb.RoleResponseDeleteAt, 0, len(roles))
	for _, role := range roles {
		if role == nil {
			continue
		}
		protoRoles = append(protoRoles, mapRoleResponseDeleteAtFromPtrs(
			role.RoleID,
			role.RoleName,
			role.CreatedAt,
			role.UpdatedAt,
			role.DeletedAt,
		))
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbutils.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	return &pb.ApiResponsePaginationRoleDeleteAt{
		Status:         "success",
		Message:        "Successfully fetched trashed roles",
		Data:           protoRoles,
		PaginationMeta: paginationMeta,
	}, nil
}

func formatCreatedAt(t interface{}) string {
	switch v := t.(type) {
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.Format("2006-01-02")
	case time.Time:
		return v.Format("2006-01-02")
	default:
		return ""
	}
}

func formatUpdatedAt(t interface{}) string {
	switch v := t.(type) {
	case *time.Time:
		if v == nil {
			return ""
		}
		return v.Format("2006-01-02")
	case time.Time:
		return v.Format("2006-01-02")
	default:
		return ""
	}
}

func mapRoleResponseDeleteAtFromPtrs(id int32, name string, createdAt, updatedAt, deletedAt interface{}) *pb.RoleResponseDeleteAt {
	response := &pb.RoleResponseDeleteAt{
		Id:        id,
		Name:      name,
		CreatedAt: formatCreatedAt(createdAt),
		UpdatedAt: formatUpdatedAt(updatedAt),
	}

	if deletedAt != nil {
		formatted := formatCreatedAt(deletedAt)
		if formatted != "" {
			response.DeletedAt = wrapperspb.String(formatted)
		}
	}

	return response
}

func (s *roleQueryHandleGrpc) FindByIdRole(ctx context.Context, req *pb.FindByIdRoleRequest) (*pb.ApiResponseRole, error) {
	roleID := int(req.GetRoleId())

	if roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleQuery.FindById(ctx, roleID)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pb.RoleResponse{
		Id:        role.RoleID,
		Name:      role.RoleName,
		CreatedAt: formatCreatedAt(role.CreatedAt),
		UpdatedAt: formatUpdatedAt(role.UpdatedAt),
	}

	return &pb.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role",
		Data:    protoRole,
	}, nil
}

func (s *roleQueryHandleGrpc) FindByNameRole(ctx context.Context, req *pb.FindByNameRoleRequest) (*pb.ApiResponseRole, error) {
	name := req.GetName()

	if name == "" {
		return nil, role_errors.ErrGrpcRoleInvalidName
	}

	role, err := s.roleQuery.FindByName(ctx, name)

	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pb.RoleResponse{
		Id:        role.RoleID,
		Name:      role.RoleName,
		CreatedAt: formatCreatedAt(role.CreatedAt),
		UpdatedAt: formatUpdatedAt(role.UpdatedAt),
	}

	return &pb.ApiResponseRole{
		Status:  "success",
		Message: "Successfully fetched role",
		Data:    protoRole,
	}, nil
}
