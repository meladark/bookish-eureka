BIN := "./bin/calendar"
BIN_SCHEDULER := "bin/calendar_scheduler"
BIN_SENDER := "bin/calendar_sender"
DOCKER_IMG="calendar:develop"

GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

POSTGRES_CONTAINER=calendar_temp_postgres
POSTGRES_PORT=5499
POSTGRES_USER=calendar
POSTGRES_PASSWORD=calendar
POSTGRES_DB=calendar
POSTGRES_VOLUME=calendar_temp_pgdata
EXIT_CODE=127

.PHONY: test-grpc generate build run build-img run-img version test lint run-db stop-db run-server test-server up down integration-tests

up:
	docker-compose -f deployments/docker-compose.yaml up -d --build

down:
	docker-compose -f deployments/docker-compose.yaml down -v --remove-orphans

integration-tests: up
	@echo "🚀 Запуск интеграционных тестов..."
	@docker-compose -f deployments/docker-compose.test.yml up \
		--abort-on-container-exit \
		--exit-code-from tests \
		--force-recreate \
		tests || \
		TEST_EXIT_CODE=$$?; \
	make down; \
	exit $$TEST_EXIT_CODE

run-server:
	@$(BIN) --config ./configs/config.toml & echo $$! > server.pid
	@sleep 2

test-server:
	@if curl -s http://localhost:8888 | grep -q "Hello, Calendar!"; then \
		echo "✅ Сервер отвечает корректно."; \
		kill -2 $$(cat server.pid); \
		rm server.pid; \
	else \
		echo "❌ Сервер не отвечает как ожидалось."; \
		kill -2 $$(cat server.pid); \
		rm server.pid; \
		exit 1; \
	fi

test-http:
	@echo "⏳ Проверка, что сервер отвечает..."
	@sleep 1
	@if ! curl -s http://localhost:8888 | grep -q "Hello, Calendar!"; then \
		echo "❌ Сервер не отвечает на /"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Сервер работает."

	@echo "\n=== Test 1: Create Event ==="
	@code=$$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8888/events \
		-H "Content-Type: application/json" \
		-d '{"id": "event-1", "title": "HTTP Test Event", "startTime": "2025-07-23T10:00:00Z", "endTime": "2025-07-23T11:00:00Z"}'); \
	if [ "$$code" != "201" ]; then \
		echo "❌ Первое создание события не удалось. Код: $$code"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Событие успешно создано."

	@echo "\n=== Test 1.1: Duplicate Event Creation ==="
	@code=$$(curl -s -o /dev/null -w "%{http_code}" -X POST http://localhost:8888/events \
		-H "Content-Type: application/json" \
		-d '{"id": "event-1", "title": "HTTP Test Event", "startTime": "2025-07-23T10:00:00Z", "endTime": "2025-07-23T11:00:00Z"}'); \
	if [ "$$code" = "201" ]; then \
		echo "❌ Повторное создание прошло успешно, хотя должно было завершиться ошибкой."; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Повторное создание корректно завершилось ошибкой."


	@echo "\n=== Test 2: List Events By Day ==="
	@response=$$(curl -s "http://localhost:8888/events/day?date=2025-07-23T00:00:00Z"); \
	echo "$$response" | grep -q "event-1"; \
	if [ $$? -ne 0 ]; then \
		echo "❌ Событие не найдено в списке за день."; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Событие найдено по дню."

	@echo "\n=== Test 3: Exclude Events Outside Day ==="
	@curl -s -X POST http://localhost:8888/events \
		-H "Content-Type: application/json" \
		-d '{"id": "event-2", "title": "Outside Day", "startTime": "2025-07-24T10:00:00Z", "endTime": "2025-07-24T11:00:00Z"}' >/dev/null
	@response=$$(curl -s "http://localhost:8888/events/day?date=2025-07-23T00:00:00Z"); \
	echo "$$response" | grep -q "event-2"; \
	if [ $$? -eq 0 ]; then \
		echo "❌ Событие за другой день попало в выборку."; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Фильтрация по дню работает корректно."

	@echo "\n=== Test 4: Invalid Date Format ==="
	@code=$$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8888/events/day?date=invalid-date"); \
	if [ "$$code" != "400" ]; then \
		echo "❌ Неверная дата не вызвала ошибку. Код ответа: $$code"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Ошибка при неверной дате обрабатывается."

	@echo "\n=== Test 5: Method Not Allowed ==="
	@code=$$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8888/events"); \
	if [ "$$code" != "405" ]; then \
		echo "❌ GET /events не вернул 405. Код ответа: $$code"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi
	@echo "✅ Метод не разрешён — проверка пройдена."

	@echo "\n✅ Все HTTP-тесты успешно выполнены."
	@kill -2 $$(cat server.pid) && rm -f server.pid


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

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/calendar
	go build -v -o $(BIN_SCHEDULER) -ldflags "$(LDFLAGS)" ./cmd/calendar_scheduler
	go build -v -o $(BIN_SENDER) -ldflags "$(LDFLAGS)" ./cmd/calendar_sender

run: build
	$(BIN) --config ./configs/config.toml

build-img:
	docker build \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run --entrypoint /opt/calendar/calendar -p 80:8888 -p 99:9999 $(DOCKER_IMG) --config /etc/calendar/config.toml

version: build
	$(BIN) version

test: run-server test-grpc test-http run-db 
	@sleep 5 
	go test -race ./internal/...
	$(MAKE) stop-db
	
install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v2.2.2

lint: install-lint-deps
	golangci-lint run ./cmd/... ./internal/... --disable=depguard --skip-dirs="vendor|migrations|build|docs|bin|deployments|api"

generate:
	go generate ./api

test-grpc: check-grpcurl
	@echo "Running gRPC service tests using grpcurl..."
	@$(eval SERVER_ADDR := localhost:9999)

	@echo -e "\n=== Test 1: Create Event ==="
	@response=$$(grpcurl -plaintext -d '{"event": {"id": "test-1", "title": "Test Event", "description": "Test Description", "startTime": 1721664000, "endTime": 1721667600, "userId": "test-user", "notifyAt": 1721660400}}' $(SERVER_ADDR) event.EventService/CreateEvent); \
	if [ $$? -ne 0 ]; then \
		echo "❌ Ошибка: событие 'test-1' не создано. Код возврата: $$?"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n=== Test 2: Get Event ==="
	@response=$$(grpcurl -plaintext -d '{"id": "test-1"}' $(SERVER_ADDR) event.EventService/GetEvent); \
	if [ $$? -ne 0 ]; then \
		echo "❌ Ошибка: событие 'test-1' не найдено. Код возврата: $$?"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n=== Test 3: List Events By Day ==="
	@response=$$(grpcurl -plaintext -d '{"date": 1721664000}' $(SERVER_ADDR) event.EventService/ListEventsByDay); \
	if [ $$? -ne 0 ]; then \
		echo "❌ Ошибка: список событий за день не получен. Код возврата: $$?"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n=== Test 4: Update Event ==="
	@response=$$(grpcurl -plaintext -d '{"event": {"id": "test-1", "title": "Updated Event", "description": "Updated Description", "startTime": 1721664000, "endTime": 1721667600, "userId": "test-user", "notifyAt": 1721660400}}' $(SERVER_ADDR) event.EventService/UpdateEvent); \
	if [ $$? -ne 0 ]; then \
		echo "❌ Ошибка: событие 'test-1' не обновлено. Код возврата: $$?"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n=== Test 5: Delete Event ==="
	@response=$$(grpcurl -plaintext -d '{"id": "test-1"}' $(SERVER_ADDR) event.EventService/DeleteEvent); \
	if [ $$? -ne 0 ]; then \
		echo "❌ Ошибка: удаление события 'test-1' завершилось ошибкой. Код возврата: $$?"; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n=== Test 6: Verify Deletion ==="
	@response=$$(grpcurl -plaintext -d '{"id": "test-1"}' $(SERVER_ADDR) event.EventService/GetEvent); \
	if [ $$? -eq 0 ]; then \
		echo "❌ Ошибка: событие 'test-1' всё ещё существует после удаления."; \
		kill -2 $$(cat server.pid) && rm -f server.pid; \
		exit 1; \
	fi

	@echo -e "\n✅ Все gRPC тесты успешно выполнены."


check-grpcurl:
	which grpcurl >/dev/null || (echo "Installing grpcurl..." && go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest)
