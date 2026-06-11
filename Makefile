.PHONY: init-env
init-env:
	cp -n .env.example .env || echo "Configuration file .env already exists"

OAPI_CODEGEN_VERSION := v2.5.0
OAPI_CODEGEN := go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@$(OAPI_CODEGEN_VERSION)
K8S_NODE ?= desktop-control-plane
K6_BASE_URL ?= http://localhost:18080
K6_DOCKER_BASE_URL ?= http://host.docker.internal:18080

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

.PHONY: run-local
run-local:
	docker-compose up -d --wait postgres kafka atdistri globalcar hexapart partspec otel-collector
	go run .

.PHONY: build
build:
	go build -v ./...

.PHONY: k8s-build-images
k8s-build-images:
	docker build -t spare-parts-api:local --build-arg SERVICE=. .
	docker build -t offer-price-worker:local --build-arg SERVICE=./cmd/offer-price-worker .
	docker build -t spare-parts-provider:local --build-arg SERVICE=./cmd/provider .

.PHONY: k8s-load-images
k8s-load-images:
	docker save spare-parts-api:local -o /tmp/spare-parts-api-local.tar
	docker save offer-price-worker:local -o /tmp/offer-price-worker-local.tar
	docker save spare-parts-provider:local -o /tmp/spare-parts-provider-local.tar
	docker exec -i $(K8S_NODE) ctr -n k8s.io images import - < /tmp/spare-parts-api-local.tar
	docker exec -i $(K8S_NODE) ctr -n k8s.io images import - < /tmp/offer-price-worker-local.tar
	docker exec -i $(K8S_NODE) ctr -n k8s.io images import - < /tmp/spare-parts-provider-local.tar

.PHONY: k8s-apply
k8s-apply:
	kubectl apply -k k8s

.PHONY: k8s-local-up
k8s-local-up: k8s-build-images k8s-load-images k8s-reset
	kubectl rollout status -n spare-parts deploy/postgres --timeout=180s
	kubectl rollout status -n spare-parts deploy/kafka --timeout=180s
	kubectl rollout status -n spare-parts deploy/atdistri --timeout=120s
	kubectl rollout status -n spare-parts deploy/globalcar --timeout=120s
	kubectl rollout status -n spare-parts deploy/hexapart --timeout=120s
	kubectl rollout status -n spare-parts deploy/partspec --timeout=120s
	kubectl rollout status -n spare-parts deploy/spare-parts-api --timeout=180s
	kubectl rollout status -n spare-parts deploy/offer-price-worker --timeout=180s

.PHONY: k8s-reset
k8s-reset:
	kubectl delete namespace spare-parts --ignore-not-found
	kubectl apply -k k8s

.PHONY: k8s-status
k8s-status:
	kubectl get all,ingress,pvc -n spare-parts

.PHONY: k8s-watch-scale
k8s-watch-scale:
	kubectl get hpa -n spare-parts -w

.PHONY: k8s-port-forward
k8s-port-forward:
	kubectl port-forward --address 0.0.0.0 -n spare-parts svc/spare-parts-api 18080:8080

.PHONY: k8s-logs-api
k8s-logs-api:
	kubectl logs -n spare-parts deploy/spare-parts-api -f

.PHONY: load-spare-part
load-spare-part:
	BASE_URL=$(K6_BASE_URL) k6 run load-tests/spare-part.js

.PHONY: load-spare-part-docker
load-spare-part-docker:
	docker run --rm -e BASE_URL=$(K6_DOCKER_BASE_URL) -v $(CURDIR)/load-tests:/scripts grafana/k6:latest run /scripts/spare-part.js

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
