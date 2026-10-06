package handler

import (
	"context"

	pbrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/role/service"
	"github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors"
	role_errors "github.com/MamangRust/microservice-payment-gateway-grpc/shared/errors/role_errors/grpc"
)

// userRoleHandleGrpc implements pbuserrole.UserRoleServiceServer. Role
// assignment lives in its own service, but the underlying work is delegated to
// the role query/command services.
type userRoleHandleGrpc struct {
	pbuserrole.UnimplementedUserRoleServiceServer

	roleQuery   service.RoleQueryService
	roleCommand service.RoleCommandService
}

// UserRoleHandleGrpc is the public interface for the user-role handler.
type UserRoleHandleGrpc interface {
	pbuserrole.UserRoleServiceServer

	CreateUserRole(ctx context.Context, request *pbuserrole.CreateUserRoleRequest) (*pbrole.ApiResponseRole, error)
	DeleteUserRole(ctx context.Context, request *pbuserrole.DeleteUserRoleRequest) (*pbrole.ApiResponseRole, error)
	FindByUserId(ctx context.Context, request *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error)
}

func NewUserRoleHandleGrpc(roleQuery service.RoleQueryService, roleCommand service.RoleCommandService) UserRoleHandleGrpc {
	return &userRoleHandleGrpc{
		roleQuery:   roleQuery,
		roleCommand: roleCommand,
	}
}

func (s *userRoleHandleGrpc) FindByUserId(ctx context.Context, request *pbuserrole.FindByIdUserRoleRequest) (*pbrole.ApiResponsesRole, error) {
	userID := int(request.GetUserId())

	if userID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	roles, err := s.roleQuery.FindByUserId(ctx, userID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRoles := make([]*pbrole.RoleResponse, len(roles))
	for i, role := range roles {
		protoRoles[i] = &pbrole.RoleResponse{
			Id:        role.RoleID,
			Name:      role.RoleName,
			CreatedAt: formatCreatedAt(role.CreatedAt),
			UpdatedAt: formatUpdatedAt(role.UpdatedAt),
		}
	}

	return &pbrole.ApiResponsesRole{
		Status:  "success",
		Message: "Successfully fetched role by user id",
		Data:    protoRoles,
	}, nil
}

func (s *userRoleHandleGrpc) CreateUserRole(ctx context.Context, request *pbuserrole.CreateUserRoleRequest) (*pbrole.ApiResponseRole, error) {
	userID := int(request.GetUserId())
	roleID := int(request.GetRoleId())

	if userID == 0 || roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	role, err := s.roleCommand.CreateUserRole(ctx, userID, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	protoRole := &pbrole.RoleResponse{
		Id:        role.RoleID,
		Name:      role.RoleName,
		CreatedAt: formatCreatedAt(role.CreatedAt),
		UpdatedAt: formatUpdatedAt(role.UpdatedAt),
	}

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully associated role with user",
		Data:    protoRole,
	}, nil
}

func (s *userRoleHandleGrpc) DeleteUserRole(ctx context.Context, request *pbuserrole.DeleteUserRoleRequest) (*pbrole.ApiResponseRole, error) {
	userID := int(request.GetUserId())
	roleID := int(request.GetRoleId())

	if userID == 0 || roleID == 0 {
		return nil, role_errors.ErrGrpcRoleInvalidId
	}

	_, err := s.roleCommand.DeleteUserRole(ctx, userID, roleID)
	if err != nil {
		return nil, errors.ToGrpcError(err)
	}

	return &pbrole.ApiResponseRole{
		Status:  "success",
		Message: "Successfully removed role from user",
	}, nil
}
