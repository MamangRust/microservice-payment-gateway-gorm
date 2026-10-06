package apps

import (
	"fmt"
	"time"

	pb_role "github.com/MamangRust/microservice-payment-gateway-grpc/pb/role"
	pb "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user"
	pbuserrole "github.com/MamangRust/microservice-payment-gateway-grpc/pb/user_role"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/adapter"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/hash"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/resilience"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/server"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/repository"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/user/service"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	roleConn, err := grpc.NewClient(viper.GetString("GRPC_ROLE_ADDR"), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Role service: %w", err)
	}

	roleQueryClient := pb_role.NewRoleQueryServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)

	repos := repository.NewRepositories(&repository.Deps{
		Db:              srv.GormDB,
		RoleQueryClient: roleQueryClient,
		UserRoleClient:  userRoleClient,
		Guard: repository.GuardOptions{
			Role: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
			UserRole: []adapter.GuardOption{
				adapter.WithDependencyGuard(resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, srv.Logger)),
			},
		},
	})
	hasher := hash.NewHashingPassword()
	svc := service.NewService(&service.Deps{
		Cache:        srv.CacheStore,
		Logger:       srv.Logger,
		Repositories: repos,
		Hash:         hasher,
	})
	h := handler.NewHandler(svc)

	srv.RegisterServices = func(gs *grpc.Server) {
		pb.RegisterUserQueryServiceServer(gs, h)
		pb.RegisterUserCommandServiceServer(gs, h)
	}

	return srv, nil
}
