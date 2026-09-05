module github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader

go 1.25.1

require (
	github.com/ClickHouse/clickhouse-go/v2 v2.43.0
	github.com/MamangRust/microservice-payment-gateway-grpc/pb v0.0.0-00010101000000-000000000000
	github.com/MamangRust/microservice-payment-gateway-grpc/pkg v0.0.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.1
)

require (
	github.com/ClickHouse/ch-go v0.71.0 // indirect
	github.com/andybalholm/brotli v1.2.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/fsnotify/fsnotify v1.9.0 // indirect
	github.com/go-faster/city v1.0.1 // indirect
	github.com/go-faster/errors v0.7.1 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/paulmach/orb v0.13.0 // indirect
	github.com/pelletier/go-toml/v2 v2.2.4 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/sagikazarmark/locafero v0.12.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
	github.com/spf13/afero v1.15.0 // indirect
	github.com/spf13/cast v1.10.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/spf13/viper v1.21.0 // indirect
	github.com/subosito/gotenv v1.6.0 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel v1.43.0 // indirect
	go.opentelemetry.io/otel/log v0.19.0 // indirect
	go.opentelemetry.io/otel/metric v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.19.0 // indirect
	go.opentelemetry.io/otel/trace v1.43.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
	google.golang.org/protobuf v1.36.11
)

replace github.com/MamangRust/microservice-payment-gateway-grpc/pb => ../../pb

replace github.com/MamangRust/microservice-payment-gateway-grpc/pkg => ../../pkg

replace github.com/MamangRust/microservice-payment-gateway-grpc/shared => ../../shared

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/apigateway => ../../service/apigateway

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/auth => ../../service/auth

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/card => ../../service/card

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/email => ../../service/email

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/merchant => ../../service/merchant

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/migrate => ../../service/migrate

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/role => ../../service/role

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/saldo => ../../service/saldo

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/topup => ../../service/topup

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/transaction => ../../service/transaction

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/transfer => ../../service/transfer

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/user => ../../service/user

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/withdraw => ../../service/withdraw

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-reader => ./

replace github.com/MamangRust/microservice-payment-gateway-grpc/service/stats-writer => ../../service/stats-writer
