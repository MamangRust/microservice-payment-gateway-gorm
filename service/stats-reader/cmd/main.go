package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pbCard "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card/stats"
	pbCardBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/card"
	pbMerchant "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant/stats"
	pbMerchantBase "github.com/MamangRust/microservice-payment-gateway-grpc/pb/merchant"
	pbSaldo "github.com/MamangRust/microservice-payment-gateway-grpc/pb/saldo/stats"
	pbTopup "github.com/MamangRust/microservice-payment-gateway-grpc/pb/topup/stats"
	pbTransaction "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transaction/stats"
	pbTransfer "github.com/MamangRust/microservice-payment-gateway-grpc/pb/transfer/stats"
	pbWithdraw "github.com/MamangRust/microservice-payment-gateway-grpc/pb/withdraw/stats"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/clickhouse"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/dotenv"
	"github.com/MamangRust/microservice-payment-gateway-grpc/pkg/logger"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/handler"
	"github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader/repository"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
	}
	log, _ := logger.NewLogger("stats-reader", nil)
	
	chConn, err := clickhouse.NewClient(log)
	if err != nil {
		log.Fatal("Failed to connect to ClickHouse", zap.Error(err))
	}

	repo := repository.NewClickHouseReaderRepository(chConn)
	cardStatsHandler := handler.NewCardStatsHandler(repo, log)
	merchantStatsHandler := handler.NewMerchantStatsHandler(repo, log)
	saldoStatsHandler := handler.NewSaldoStatsHandler(repo, log)
	topupStatsHandler := handler.NewTopupStatsHandler(repo, log)
	transactionStatsHandler := handler.NewTransactionStatsHandler(repo, log)
	transferStatsHandler := handler.NewTransferStatsHandler(repo, log)
	withdrawStatsHandler := handler.NewWithdrawStatsHandler(repo, log)

	grpcServer := grpc.NewServer()
	
	pbCard.RegisterCardStatsBalanceServiceServer(grpcServer, cardStatsHandler)
	pbCard.RegisterCardStatsTopupServiceServer(grpcServer, cardStatsHandler)
	pbCard.RegisterCardStatsTransactionServiceServer(grpcServer, cardStatsHandler)
	pbCard.RegisterCardStatsTransferServiceServer(grpcServer, cardStatsHandler)
	pbCard.RegisterCardStatsWithdrawServiceServer(grpcServer, cardStatsHandler)
	pbCardBase.RegisterCardDashboardServiceServer(grpcServer, cardStatsHandler)

	pbMerchant.RegisterMerchantStatsAmountServiceServer(grpcServer, merchantStatsHandler)
	pbMerchant.RegisterMerchantStatsMethodServiceServer(grpcServer, merchantStatsHandler)
	pbMerchant.RegisterMerchantStatsTotalAmountServiceServer(grpcServer, merchantStatsHandler)
	pbMerchantBase.RegisterMerchantTransactionServiceServer(grpcServer, merchantStatsHandler)

	pbSaldo.RegisterSaldoStatsBalanceServiceServer(grpcServer, saldoStatsHandler)
	pbSaldo.RegisterSaldoStatsTotalBalanceServer(grpcServer, saldoStatsHandler)

	pbTopup.RegisterTopupStatsAmountServiceServer(grpcServer, topupStatsHandler)
	pbTopup.RegisterTopupStatsMethodServiceServer(grpcServer, topupStatsHandler)
	pbTopup.RegisterTopupStatsStatusServiceServer(grpcServer, topupStatsHandler)

	pbTransaction.RegisterTransactionStatsAmountServiceServer(grpcServer, transactionStatsHandler)
	pbTransaction.RegisterTransactionStatsMethodServiceServer(grpcServer, transactionStatsHandler)
	pbTransaction.RegisterTransactionStatsStatusServiceServer(grpcServer, transactionStatsHandler)

	pbTransfer.RegisterTransferStatsAmountServiceServer(grpcServer, transferStatsHandler)
	pbTransfer.RegisterTransferStatsStatusServiceServer(grpcServer, transferStatsHandler)

	pbWithdraw.RegisterWithdrawStatsAmountServiceServer(grpcServer, withdrawStatsHandler)
	pbWithdraw.RegisterWithdrawStatsStatusServiceServer(grpcServer, withdrawStatsHandler)

	reflection.Register(grpcServer)

	port := ":50062"
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("Failed to listen", zap.Error(err))
	}

	log.Info("Stats Reader starting", zap.String("port", port))

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("Failed to serve", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down Stats Reader...")
	grpcServer.GracefulStop()
}
