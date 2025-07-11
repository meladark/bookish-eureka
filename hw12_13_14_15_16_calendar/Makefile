BIN := "./bin/calendar"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

POSTGRES_CONTAINER=calendar_temp_postgres
POSTGRES_PORT=5499
POSTGRES_USER=calendar
POSTGRES_PASSWORD=calendar
POSTGRES_DB=calendar
POSTGRES_VOLUME=calendar_temp_pgdata

.PHONY: build run build-img run-img version test lint run-db stop-db run-server test-server

run-server:
	@$(BIN) --config ./configs/config.toml & echo $$! > server.pid
	@sleep 2

test-server:
	@if curl -s http://localhost:881 | grep -q "Hello, Calendar!"; then \
		echo "✅ Сервер отвечает корректно."; \
		kill -2 $$(cat server.pid); \
		rm server.pid; \
	else \
		echo "❌ Сервер не отвечает как ожидалось."; \
		kill -2 $$(cat server.pid); \
		rm server.pid; \
		exit 1; \
	fi

run-db:
	docker run -d --name $(POSTGRES_CONTAINER) \
		-e POSTGRES_USER=$(POSTGRES_USER) \
		-e POSTGRES_PASSWORD=$(POSTGRES_PASSWORD) \
		-e POSTGRES_DB=$(POSTGRES_DB) \
		-p $(POSTGRES_PORT):5432 \
		-v $(POSTGRES_VOLUME):/var/lib/postgresql/data \
		postgres:latest

stop-db:
	-docker stop $(POSTGRES_CONTAINER)
	-docker rm $(POSTGRES_CONTAINER)
	-docker volume rm $(POSTGRES_VOLUME)

full-cycle: run-db
	@sleep 5 # дать время на старт базы
	$(MAKE) do-something
	$(MAKE) stop-db

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar

run: build
	$(BIN) --config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run -p 80:8080 $(DOCKER_IMG) 

version: build
	$(BIN) version

test: run-server test-server run-db
	@sleep 5 
	go test -race ./internal/...
	$(MAKE) stop-db
	
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.57.2

lint: install-lint-deps
	golangci-lint run ./cmd/... ./internal/...
