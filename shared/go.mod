module github.com/MamangRust/microservice-payment-gateway-grpc/shared

go 1.25.1

require (
	github.com/MamangRust/microservice-payment-gateway-grpc/pb v0.0.0-00010101000000-000000000000
	github.com/MamangRust/microservice-payment-gateway-grpc/pkg v0.0.0
	github.com/go-playground/validator/v10 v10.30.1
	github.com/google/uuid v1.6.0
	github.com/labstack/echo/v4 v4.15.0
	github.com/redis/go-redis/v9 v9.17.3
	go.opentelemetry.io/otel v1.43.0
	go.opentelemetry.io/otel/metric v1.43.0
	go.opentelemetry.io/otel/trace v1.43.0
	go.uber.org/zap v1.27.1
	google.golang.org/grpc v1.79.1
	google.golang.org/protobuf v1.36.11
	gorm.io/gorm v1.31.2
)

require (
	github.com/IBM/sarama v1.46.3 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/eapache/go-resiliency v1.7.0 // indirect
	github.com/eapache/go-xerial-snappy v0.0.0-20230731223053-c322873962e3 // indirect
	github.com/eapache/queue v1.1.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.13 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang/snappy v1.0.0 // indirect
	github.com/hashicorp/go-uuid v1.0.3 // indirect
	github.com/jcmturner/aescts/v2 v2.0.0 // indirect
	github.com/jcmturner/dnsutils/v2 v2.0.0 // indirect
	github.com/jcmturner/gofork v1.7.6 // indirect
	github.com/jcmturner/gokrb5/v8 v8.4.4 // indirect
	github.com/jcmturner/rpc/v2 v2.0.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.5 // indirect
	github.com/labstack/gommon v0.4.2 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.21 // indirect
	github.com/pierrec/lz4/v4 v4.1.26 // indirect
	github.com/rcrowley/go-metrics v0.0.0-20250401214520-65e299d6c5c9 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.opentelemetry.io/auto/sdk v1.2.1 // indirect
	go.opentelemetry.io/otel/log v0.19.0 // indirect
	go.opentelemetry.io/otel/sdk v1.43.0 // indirect
	go.opentelemetry.io/otel/sdk/log v0.19.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260420184626-e10c466a9529 // indirect
)

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
