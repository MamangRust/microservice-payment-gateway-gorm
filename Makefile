COMPOSE_FILE=deployments/local/docker-compose.yml
SERVICES := apigateway migrate auth user role card merchant saldo topup transaction transfer withdraw email
DOCKER_COMPOSE=docker compose
PROTO_DIR=proto
OUTDIR_PROTO=pb

migrate:
	go run service/migrate/main.go up

migrate-down:
	go run service/migrate/main.go down


generate-proto:
	protoc \
		--proto_path=$(PROTO_DIR) \
		--go_out=$(OUTDIR_PROTO) --go_opt=paths=source_relative \
		--go-grpc_out=$(OUTDIR_PROTO) --go-grpc_opt=paths=source_relative \
		$$(find $(PROTO_DIR) -name "*.proto")


generate-sql:
	@for svc in user auth card merchant role saldo topup transaction transfer withdraw; do \
		echo "Generating sqlc for $$svc..."; \
		(cd service/$$svc/sqlc && sqlc generate) || exit 1; \
	done
	@echo "✅ All sqlc regenerated."


generate-swagger:
	swag init -g service/apigateway/cmd/main.go -o service/apigateway/docs

seeder:
	go run service/seeder/main.go


build-image:
	@for service in $(SERVICES); do \
		echo "🔨 Building microservice-payment-gateway-grpc/$$service:latest..."; \
		docker build -t microservice-payment-gateway-grpc/$$service:latest -f service/$$service/Dockerfile . || exit 1; \
	done
	@echo "✅ All services built successfully."

image-load:
	@for service in $(SERVICES); do \
		echo "🚚 Loading microservice-payment-gateway-grpc/$$service:latest..."; \
		minikube image load microservice-payment-gateway-grpc/$$service:latest || exit 1; \
	done
	@echo "✅ All services loaded successfully."


image-delete:
	@for service in $(SERVICES); do \
		echo "🗑️ Deleting microservice-payment-gateway-grpc/$$service:latest image..."; \
		minikube image rm microservice-payment-gateway-grpc/$$service:latest || echo "⚠️ Failed to delete (maybe not found)"; \
	done
	@echo "✅ All requested images deleted (if they existed)."


ps:
	${DOCKER_COMPOSE} -f $(COMPOSE_FILE) ps

up:
	${DOCKER_COMPOSE} -f $(COMPOSE_FILE) up -d

down:
	${DOCKER_COMPOSE} -f $(COMPOSE_FILE) down

build-up:
	make build-image && make up

kube-start:
	minikube start --driver=docker

kube-up:
	kubectl apply -f deployments/kubernetes/namespace.yaml
	kubectl apply -f deployments/kubernetes

kube-down:
	kubectl delete -f deployments/kubernetes --ignore-not-found
	kubectl delete -f deployments/kubernetes/namespace.yaml --ignore-not-found

kube-status:
	@echo "🔍 Checking Pods in payment-gateway..."
	@kubectl get pods -n payment-gateway

	@echo "\n🔍 Checking Services in payment-gateway..."
	@kubectl get svc -n payment-gateway

	@echo "\n🔍 Checking PVCs in payment-gateway..."
	@kubectl get pvc -n payment-gateway

	@echo "\n🔍 Checking Jobs in payment-gateway..."
	@kubectl get jobs -n payment-gateway

kube-tunnel:
	minikube tunnel


test-auth:
	@APP_ENV=development go test service/auth/tests/... -v
