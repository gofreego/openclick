build: clean build-ui
	go build -o application .
build-linux: clean build-ui
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o application .
run:
	go run main.go
test:
	go test -v ./...
clean:
	rm -f application

# UI build commands
build-ui:
	cd dashboard-ui && npm install && npm run build

dev-ui:
	cd dashboard-ui && npm run dev

clean-ui:
	cd dashboard-ui && rm -rf dist node_modules

docker: build-linux
	docker build -t openclick .
	rm -f application

docker-run: docker
	@echo "Tagging image as latest"
	docker tag openclick openclick:latest
	@echo "removing existing container named openclick if any"
	docker rm -f openclick || true
	@echo "Running image with name openclick, mapping ports 8085:8085 and 8086:8086"
	docker run -d --name openclick -p 8085:8085 -p 8086:8086 openclick:latest

install: 
	go mod tidy
	go get github.com/grpc-ecosystem/grpc-gateway/v2/internal/descriptor@v2.27.2
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2
	go install google.golang.org/protobuf/cmd/protoc-gen-go
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	go install github.com/envoyproxy/protoc-gen-validate@latest
	go install github.com/gofreego/goutils/cmd/sql-migrator@v1.3.9

setup:
	@echo "Compiling proto files..."
	sh ./api/protoc.sh
	go mod tidy

migrate:
	sql-migrator ./migrator.yaml

migrate-clickhouse:
	sql-migrator ./clickhouse-migrator.yaml

redeploy:
	@echo "Redeploying the application"
	@echo "Pulling latest changes from git"
	git pull
	@echo "Building the docker imamge"
	docker compose build
	@echo "Stopping existing docker containers"
	docker compose down
	@echo "Starting the docker containers"
	docker compose up -d
