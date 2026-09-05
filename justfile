set shell := ["bash", "-c"]

COMPOSE_FILE := "deployments/local/docker-compose.yml"
SERVICES := "apigateway auth user role card merchant saldo topup transaction transfer withdraw email ai-security stats-reader stats-writer"
DOCKER_COMPOSE := "docker compose"
PROTO_DIR := "proto"
OUTDIR_PROTO := "pb"

# List all recipes
default:
    @just --list


# Validate all modules (build and unit test)
validate:
    @echo "🔍 Validating all workspace modules..."
    @go list -m -f '{{ "{{" }}.Dir{{ "}}" }}/...' | grep -v 'tests' | xargs go build
    @go list -m -f '{{ "{{" }}.Dir{{ "}}" }}/...' | grep -v 'tests' | xargs go test -short
    @echo "✅ Validation successful."


# Generate protocol buffers
generate-proto:
    protoc \
        --proto_path={{PROTO_DIR}} \
        --go_out={{OUTDIR_PROTO}} --go_opt=paths=source_relative \
        --go-grpc_out={{OUTDIR_PROTO}} --go-grpc_opt=paths=source_relative \
        $(find {{PROTO_DIR}} -name "*.proto")

# Generate swagger documentation
generate-swagger:
    swag init -g service/apigateway/cmd/main.go -o service/apigateway/docs

# Build docker images for all services
build-image:
    @for service in {{SERVICES}}; do \
        echo "🔨 Building microservice-payment-gateway-grpc/$service:latest..."; \
        docker build -t microservice-payment-gateway-grpc/$service:latest -f service/$service/Dockerfile . || exit 1; \
    done
    @echo "✅ All services built successfully."

# Load images to minikube
image-load:
    @for service in {{SERVICES}}; do \
        echo "🚚 Loading microservice-payment-gateway-grpc/$service:latest..."; \
        minikube image load microservice-payment-gateway-grpc/$service:latest || exit 1; \
    done
    @echo "✅ All services loaded successfully."

# Delete images from minikube
image-delete:
    @for service in {{SERVICES}}; do \
        echo "🗑️ Deleting microservice-payment-gateway-grpc/$service:latest image..."; \
        minikube image rm microservice-payment-gateway-grpc/$service:latest || echo "⚠️ Failed to delete image (maybe not found)"; \
    done
    @echo "✅ All requested images deleted (if they existed)."

# Show docker compose process status
ps:
    {{DOCKER_COMPOSE}} -f {{COMPOSE_FILE}} ps

# Start docker compose services
up:
    {{DOCKER_COMPOSE}} -f {{COMPOSE_FILE}} up -d

# Stop docker compose services
down:
    {{DOCKER_COMPOSE}} -f {{COMPOSE_FILE}} down

# Build images and start docker compose
build-up: build-image up

# Start minikube with docker driver
kube-start:
    minikube start --driver=docker

# Apply kubernetes manifests
kube-up:
    kubectl apply -f deployments/kubernetes/namespace.yaml
    kubectl apply -f deployments/kubernetes

# Delete kubernetes manifests
kube-down:
    kubectl delete -f deployments/kubernetes --ignore-not-found
    kubectl delete -f deployments/kubernetes/namespace.yaml --ignore-not-found

# Show kubernetes status
kube-status:
    @echo "🔍 Checking Pods in payment-gateway..."
    @kubectl get pods -n payment-gateway
    @echo -e "\n🔍 Checking Services in payment-gateway..."
    @kubectl get svc -n payment-gateway
    @echo -e "\n🔍 Checking PVCs in payment-gateway..."
    @kubectl get pvc -n payment-gateway
    @echo -e "\n🔍 Checking Jobs in payment-gateway..."
    @kubectl get jobs -n payment-gateway

# Tunnel minikube services
kube-tunnel:
    minikube tunnel

# Run unit tests in pkg/
test-unit:
    @echo "🧪 Running unit tests in pkg/..."
    @cd pkg && go test ./... -v

# Run integration tests in tests/
test-integration:
    @echo "🧪 Running integration tests in tests/..."
    @cd tests && APP_ENV=development go test ./... -v

# Run all tests (unit and integration)
test-all: test-unit test-integration

# Run auth service integration tests
test-auth:
    @echo "🧪 Running auth integration tests..."
    @cd tests && APP_ENV=development go test ./auth/... -v

# Run ai-security tests (Python)
test-ai-security:
    @echo "🧪 Running ai-security tests..."
    @PYTHONPATH=service/ai-security:service/ai-security/generated/ai_security pytest service/ai-security/tests -v

# Build all Go service binaries (from api-gateway to withdraw)
build:
    @mkdir -p bin
    @for mod in service/*/go.mod; do \
        dir=$(dirname $mod); \
        service=$(basename $dir); \
        echo "🔨 Building $service..."; \
        (cd $dir && go build -o ../../bin/$service ./cmd/main.go) || exit 1; \
    done
    @echo "✅ All services built successfully in bin/ folder."

# Run go mod tidy for all services
tidy-all:
    @echo "🧹 Tidying all service modules..."
    @for service in {{SERVICES}}; do \
        if [ -d "service/$service" ]; then \
            echo "📦 Tidying $service..."; \
            (cd service/$service && go mod tidy) || echo "⚠️ Failed to tidy $service"; \
        fi \
    done
    @echo "✅ All services tidied."
