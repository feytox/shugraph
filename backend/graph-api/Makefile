.PHONY: sql-gen new-migrate test integration-test lint openapi
sql-gen:
	cd db && go tool sqlc generate

new-migrate:
	@if [ -z "$(name)" ]; then \
			echo "Error: Use: make create n=name"; \
			exit 1; \
		fi
		cd db/migrations && go tool goose create $(name) sql

test :
	go test ./...

integration-test:
	go test -tags=integration ./...

lint:
	golangci-lint run ./...

openapi:
	go run ./cmd openapi > openapi.yaml