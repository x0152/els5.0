.PHONY: help up down logs ps rebuild restart sh psql run build test lint fmt tidy openapi openapi-all migrate-up migrate-down bootstrap sqlc-generate sqlc-vet image image-push new-app spacy security security-gosec security-vuln security-trivy security-gitleaks

APP_NAME := els-backend
BIN_DIR  := bin
OPENAPI_DIR := docs/openapi

IMAGE_REPO    ?= els/backend
IMAGE_TAG     ?= dev
IMAGE_PLATFORM?= linux/amd64
GIT_COMMIT    := $(shell git rev-parse --short HEAD 2>/dev/null || echo unknown)

GOBIN              := $(shell go env GOPATH)/bin
GOSEC_VERSION      ?= v2.26.1
GOVULNCHECK_VERSION?= v1.1.4
SECURITY_GOTOOLCHAIN ?= go1.25.11
TRIVY_HELM_KUBE_VERSION ?= 1.27.0

help: ## show all targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-14s\033[0m %s\n",$$1,$$2}'

up: ## start the dev stack (postgres + redis + air-based dev server)
	docker compose up --build --remove-orphans

up-d: ## start the dev stack in the background
	docker compose up --build --remove-orphans -d

down: ## stop and remove containers
	docker compose down

down-v: ## stop and remove volumes (wipe the DB)
	docker compose down -v

logs: ## tail logs of the dev server
	docker compose logs -f dev

ps: ## container status
	docker compose ps

rebuild: ## rebuild the dev server image (no go mod caches)
	docker compose build --no-cache dev

restart: ## restart the dev server
	docker compose restart dev

spacy: ## start the spaCy service (needed for film subtitle parsing)
	docker compose --profile spacy up --build -d spacy

speech: ## start the pronunciation assessment service (needed for the Speaking app)
	docker compose --profile speech up --build -d speech

sh: ## shell inside the dev server
	docker compose exec dev sh

psql: ## psql into postgres
	docker compose exec postgres sh -c 'psql -U "$$POSTGRES_USER" -d "$$POSTGRES_DB"'

run: ## run the dev server on the host (via air)
	air -c .air.toml

build: ## build the dev binary into ./bin (all-modules aggregator, local development only)
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/dev ./cmd

test: ## tests
	go test ./...

lint: ## go vet
	go vet ./...

fmt: ## go fmt
	go fmt ./...

tidy: ## go mod tidy
	go mod tidy

new-app: ## new HTTP app from template (make new-app MODULE=billing)
	@test -n "$(MODULE)" || (echo "usage: make new-app MODULE=<name>  (lowercase, e.g. billing)"; exit 2)
	go run ./tools/new-app -module=$(MODULE)

openapi: ## generate openapi for a module (make openapi MODULE=auth FMT=yaml)
	@mkdir -p $(OPENAPI_DIR)
	@MODULE=$${MODULE:-auth}; FMT=$${FMT:-yaml}; \
	go run ./tools/openapi-dump -module=$$MODULE -format=$$FMT -out=$(OPENAPI_DIR)/$$MODULE.$$FMT; \
	echo "wrote $(OPENAPI_DIR)/$$MODULE.$$FMT"

openapi-all: ## generate openapi for all registered modules (yaml + json)
	@mkdir -p $(OPENAPI_DIR)
	@for m in $$(go run ./tools/openapi-dump -list); do \
		for fmt in yaml json; do \
			go run ./tools/openapi-dump -module=$$m -format=$$fmt -out=$(OPENAPI_DIR)/$$m.$$fmt && echo "wrote $(OPENAPI_DIR)/$$m.$$fmt"; \
		done; \
	done

migrate-up: ## apply all up migrations (locally, on the host)
	go run ./cmd/migrator -direction=up

migrate-down: ## roll back one migration (locally, on the host)
	go run ./cmd/migrator -direction=down

bootstrap: ## create the default global administrator (values from .env)
	go run ./tools/bootstrap

sqlc-generate: ## regenerate sqlc code from queries/*.sql against the migration schema
	cd internal/infrastructure/postgres && sqlc generate

sqlc-vet: ## static check of sqlc config and queries
	cd internal/infrastructure/postgres && sqlc vet

image: ## build the prod image (IMAGE_REPO/IMAGE_TAG/IMAGE_PLATFORM are overridable)
	docker buildx build \
		--platform $(IMAGE_PLATFORM) \
		--file Dockerfile.prod \
		--build-arg VERSION=$(IMAGE_TAG) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(IMAGE_REPO):$(IMAGE_TAG) \
		--load \
		.

image-push: ## build and push the prod image (IMAGE_PLATFORM=linux/amd64,linux/arm64 for multi-arch)
	docker buildx build \
		--platform $(IMAGE_PLATFORM) \
		--file Dockerfile.prod \
		--build-arg VERSION=$(IMAGE_TAG) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		--tag $(IMAGE_REPO):$(IMAGE_TAG) \
		--push \
		.

security: security-gosec security-vuln security-trivy security-gitleaks ## full security scan (sast + vuln + secrets + fs)

security-gosec: ## SAST on sources (gosec)
	@test -x $(GOBIN)/gosec || go install github.com/securego/gosec/v2/cmd/gosec@$(GOSEC_VERSION)
	$(GOBIN)/gosec -exclude-dir=tools/new-app/_template -severity=medium -confidence=medium ./...

security-vuln: ## CVE scanner for modules and the current Go version (govulncheck)
	@if ! test -x $(GOBIN)/govulncheck || ! go version -m $(GOBIN)/govulncheck 2>/dev/null | grep -q ': $(SECURITY_GOTOOLCHAIN)$$'; then \
		GOTOOLCHAIN=$(SECURITY_GOTOOLCHAIN) go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \
	fi
	GOTOOLCHAIN=$(SECURITY_GOTOOLCHAIN) $(GOBIN)/govulncheck ./...

security-trivy: ## trivy fs (vulns + secrets + misconfig)
	@command -v trivy >/dev/null 2>&1 || (echo "trivy is not installed. See https://aquasecurity.github.io/trivy/"; exit 1)
	trivy fs --scanners vuln,secret,misconfig --severity HIGH,CRITICAL --ignore-unfixed --helm-kube-version $(TRIVY_HELM_KUBE_VERSION) --skip-files Dockerfile.dev .

security-gitleaks: ## gitleaks: search for secrets in the repo
	@command -v gitleaks >/dev/null 2>&1 || (echo "gitleaks is not installed. See https://github.com/gitleaks/gitleaks"; exit 1)
	gitleaks detect --no-banner --redact --source ..
