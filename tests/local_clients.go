package tests

import (
	"context"

	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	role_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/role/handler"
	user_handler "github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// LocalUserClient implements both UserQueryServiceClient and UserCommandServiceClient
type LocalUserClient struct {
	Handler user_handler.Handler
}

func (c *LocalUserClient) FindAll(ctx context.Context, in *user.FindAllUserRequest, opts ...grpc.CallOption) (*user.ApiResponsePaginationUser, error) {
	return c.Handler.FindAll(ctx, in)
}

func (c *LocalUserClient) FindById(ctx context.Context, in *user.FindByIdUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.FindById(ctx, in)
}

func (c *LocalUserClient) FindByEmail(ctx context.Context, in *user.FindByEmailUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.FindByEmail(ctx, in)
}

func (c *LocalUserClient) FindByVerificationCode(ctx context.Context, in *user.FindByVerificationCodeUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.FindByVerificationCode(ctx, in)
}

func (c *LocalUserClient) FindByActive(ctx context.Context, in *user.FindAllUserRequest, opts ...grpc.CallOption) (*user.ApiResponsePaginationUserDeleteAt, error) {
	return c.Handler.FindByActive(ctx, in)
}

func (c *LocalUserClient) FindByTrashed(ctx context.Context, in *user.FindAllUserRequest, opts ...grpc.CallOption) (*user.ApiResponsePaginationUserDeleteAt, error) {
	return c.Handler.FindByTrashed(ctx, in)
}

func (c *LocalUserClient) Create(ctx context.Context, in *user.CreateUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.Create(ctx, in)
}

func (c *LocalUserClient) Update(ctx context.Context, in *user.UpdateUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.Update(ctx, in)
}

func (c *LocalUserClient) UpdateIsVerified(ctx context.Context, in *user.UpdateUserIsVerifiedRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.UpdateIsVerified(ctx, in)
}

func (c *LocalUserClient) UpdatePassword(ctx context.Context, in *user.UpdateUserPasswordRequest, opts ...grpc.CallOption) (*user.ApiResponseUser, error) {
	return c.Handler.UpdatePassword(ctx, in)
}

func (c *LocalUserClient) TrashedUser(ctx context.Context, in *user.FindByIdUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUserDeleteAt, error) {
	return c.Handler.TrashedUser(ctx, in)
}

func (c *LocalUserClient) RestoreUser(ctx context.Context, in *user.FindByIdUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUserDeleteAt, error) {
	return c.Handler.RestoreUser(ctx, in)
}

func (c *LocalUserClient) DeleteUserPermanent(ctx context.Context, in *user.FindByIdUserRequest, opts ...grpc.CallOption) (*user.ApiResponseUserDelete, error) {
	return c.Handler.DeleteUserPermanent(ctx, in)
}

func (c *LocalUserClient) RestoreAllUser(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*user.ApiResponseUserAll, error) {
	return c.Handler.RestoreAllUser(ctx, in)
}

func (c *LocalUserClient) DeleteAllUserPermanent(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*user.ApiResponseUserAll, error) {
	return c.Handler.DeleteAllUserPermanent(ctx, in)
}

// LocalRoleClient implements both RoleQueryServiceClient and RoleCommandServiceClient
type LocalRoleClient struct {
	Handler *role_handler.Handler
}

func (c *LocalRoleClient) FindAllRole(ctx context.Context, in *role.FindAllRoleRequest, opts ...grpc.CallOption) (*role.ApiResponsePaginationRole, error) {
	return c.Handler.RoleQuery.FindAllRole(ctx, in)
}

func (c *LocalRoleClient) FindByIdRole(ctx context.Context, in *role.FindByIdRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleQuery.FindByIdRole(ctx, in)
}

func (c *LocalRoleClient) CreateRole(ctx context.Context, in *role.CreateRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleCommand.CreateRole(ctx, in)
}

func (c *LocalRoleClient) UpdateRole(ctx context.Context, in *role.UpdateRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleCommand.UpdateRole(ctx, in)
}

func (c *LocalRoleClient) DeleteRolePermanent(ctx context.Context, in *role.FindByIdRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRoleDelete, error) {
	return c.Handler.RoleCommand.DeleteRolePermanent(ctx, in)
}

func (c *LocalRoleClient) CreateUserRole(ctx context.Context, in *role.CreateUserRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleCommand.CreateUserRole(ctx, in)
}

func (c *LocalRoleClient) DeleteUserRole(ctx context.Context, in *role.DeleteUserRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleCommand.DeleteUserRole(ctx, in)
}

func (c *LocalRoleClient) TrashedRole(ctx context.Context, in *role.FindByIdRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRoleDeleteAt, error) {
	return c.Handler.RoleCommand.TrashedRole(ctx, in)
}

func (c *LocalRoleClient) RestoreRole(ctx context.Context, in *role.FindByIdRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRoleDeleteAt, error) {
	return c.Handler.RoleCommand.RestoreRole(ctx, in)
}

func (c *LocalRoleClient) RestoreAllRole(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*role.ApiResponseRoleAll, error) {
	return c.Handler.RoleCommand.RestoreAllRole(ctx, in)
}

func (c *LocalRoleClient) DeleteAllRolePermanent(ctx context.Context, in *emptypb.Empty, opts ...grpc.CallOption) (*role.ApiResponseRoleAll, error) {
	return c.Handler.RoleCommand.DeleteAllRolePermanent(ctx, in)
}

func (c *LocalRoleClient) FindByUserId(ctx context.Context, in *role.FindByIdUserRoleRequest, opts ...grpc.CallOption) (*role.ApiResponsesRole, error) {
	return c.Handler.RoleQuery.FindByUserId(ctx, in)
}

func (c *LocalRoleClient) FindByActive(ctx context.Context, in *role.FindAllRoleRequest, opts ...grpc.CallOption) (*role.ApiResponsePaginationRoleDeleteAt, error) {
	return c.Handler.RoleQuery.FindByActive(ctx, in)
}

func (c *LocalRoleClient) FindByTrashed(ctx context.Context, in *role.FindAllRoleRequest, opts ...grpc.CallOption) (*role.ApiResponsePaginationRoleDeleteAt, error) {
	return c.Handler.RoleQuery.FindByTrashed(ctx, in)
}

func (c *LocalRoleClient) FindByNameRole(ctx context.Context, in *role.FindByNameRoleRequest, opts ...grpc.CallOption) (*role.ApiResponseRole, error) {
	return c.Handler.RoleQuery.FindByNameRole(ctx, in)
}
