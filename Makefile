GO := go

TEST := $(GO) test

DOCKER_COMPOSE := docker-compose

OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen

CONFIG_FILE := dto_generator_cfg.yaml
INPUT_FILE := swagger.yaml

generate:
	@$(GO) install github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@v2.2.0
	@$(GO) get github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen
	@$(OAPI_CODEGEN) -config $(CONFIG_FILE) $(INPUT_FILE)

linter:
	golangci-lint run ./cmd/config/...
	golangci-lint run ./cmd/initdb/...
	golangci-lint run ./internal/...
	golangci-lint run ./test/integration/...


unit:
	@$(TEST) ./internal/usecase/...

integration:
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up -d
	@$(TEST) ./test/integration/...
	@$(DOCKER_COMPOSE) down

coverage_t:
	@$(DOCKER_COMPOSE) -f docker-compose.test.yml up -d
	@$(TEST) -coverprofile "coverage/cover.out" ./...
	@$(GO) run coverage/filter_coverage.go coverage/cover.out coverage/filtered_coverage.out
	@$(GO) tool cover -func "coverage/filtered_coverage.out"
	@$(DOCKER_COMPOSE) down

up:
	@$(DOCKER_COMPOSE) up --build -d

up_log:
	@$(DOCKER_COMPOSE) up --build

down:
	@$(DOCKER_COMPOSE) down

start: unit integration up
start_log:  unit integration up_log
stop: down
