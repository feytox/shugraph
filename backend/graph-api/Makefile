.PHONY: gen openapi test integration-test lint new-migration

gen:
	cd db && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
	$(MAKE) openapi

openapi:
	go run ./cmd openapi > openapi.yaml

test:
	go vet ./... && go test -race ./...

integration-test:
	go test -race -tags=integration ./...

lint:
	golangci-lint run ./...

new-migration:
	@test -n "$(name)" || (echo "usage: make new-migration name=<name>" && exit 1)
	cd db/migrations && go run github.com/pressly/goose/v3/cmd/goose@v3.28.0 create $(name) sql
