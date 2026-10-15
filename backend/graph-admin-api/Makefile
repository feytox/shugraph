.PHONY: openapi test lint new-migration

openapi:
	go run ./cmd openapi > openapi.yaml

test:
	go vet ./... && go test -race ./...

lint:
	golangci-lint run ./...

new-migration:
	@test -n "$(name)" || (echo "usage: make new-migration name=<name>" && exit 1)
	cd db/migrations && go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 create $(name) sql
