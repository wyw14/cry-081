.PHONY: fmt test race vet build run compose-up compose-down

fmt:
	test -z "$$(gofmt -l $$(find . -name '*.go' -not -path './vendor/*'))"

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

build:
	go build ./...

run:
	go run ./cmd/server

compose-up:
	docker compose up --build

compose-down:
	docker compose down
