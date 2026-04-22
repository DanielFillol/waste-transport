APP=waste-api
CMD=./cmd/api

.PHONY: run build tidy lint test test-docker docker-up docker-down docker-logs

run:
	go run $(CMD)/main.go

build:
	go build -o bin/$(APP) $(CMD)/main.go

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

# Run e2e tests locally (requires postgres-test on :5435)
test:
	docker compose -f docker-compose.test.yml up -d postgres-test
	go test -v -timeout 300s ./tests/e2e/...

# Run e2e tests fully inside Docker
test-docker:
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit test
	docker compose -f docker-compose.test.yml down

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-logs:
	docker compose logs -f api
