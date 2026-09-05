module github.com/MamangRust/microservice-payment-gateway-grpc/pb

go 1.25.0

replace github.com/MamangRust/microservice-payment-gateway-grpc/pb => ../pb

replace github.com/MamangRust/microservice-payment-gateway-grpc/pkg => ../pkg

replace github.com/MamangRust/microservice-payment-gateway-grpc/shared => ../shared

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway => ../service/apigateway

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/auth => ../service/auth

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/card => ../service/card

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/email => ../service/email

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant => ../service/merchant

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/migrate => ../service/migrate

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/role => ../service/role

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo => ../service/saldo

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/topup => ../service/topup

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction => ../service/transaction

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer => ../service/transfer

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/user => ../service/user

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw => ../service/withdraw

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader => ../service/stats-reader

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer => ../service/stats-writer

require (
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
)

require (
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/metric v1.43.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
)
