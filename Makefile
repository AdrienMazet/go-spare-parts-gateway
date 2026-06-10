.PHONY: init-env
init-env:
	cp -n .env.example .env || echo "Configuration file .env already exists"

OAPI_CODEGEN_VERSION := v2.5.0
OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)

.PHONY: generate-api-models
generate-api-models:
	$(OAPI_CODEGEN) --config api/config.oapi-codegen.yaml api/api-definition.yaml

.PHONY: generate-mocks
generate-mocks:
	go install github.com/golang/mock/mockgen@v1.6.0
	go generate ./...

.PHONY: run
run:
	docker-compose up -d --wait
	go run .

.PHONY: build
build:
	go build -v ./...

.PHONY: run-tests
run-tests:
	go test ./... -v -short -count=1

.PHONY: coverage
coverage:
	go test ./... -coverprofile=coverage.out
	grep -Ev '(/mock_[^/]+\.go:|/api\.gen\.go:|/main\.go:|/routes\.go:)' coverage.out > coverage.filtered.out
	go tool cover -func=coverage.filtered.out

.PHONY: lint
lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
	@golangci-lint run ./...

.PHONY: tidy
tidy:
	go install golang.org/x/tools/cmd/goimports@latest
	goimports -w .
	go mod tidy

.PHONY: tidy-check
tidy-check:
	go mod tidy
	git diff --exit-code
