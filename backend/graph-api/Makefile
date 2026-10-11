.PHONY: sql-gen load-spec init up down test lint migrate new-migrate

WEB_PORT=8080

sql-gen:
	cd db && go tool sqlc generate

new-migrate:
	@if [ -z "$(name)" ]; then \
			echo "Error: Use: make create n=name"; \
			exit 1; \
		fi
		cd db/migrations && go tool goose create $(name) sql
load-spec:
	curl http://localhost:${WEB_PORT}/openapi.json -o openapi.json

up:
	 docker compose up -d --build

migrate:
	docker compose run --rm migrate

down:
	docker compose down

test :
	go test ./...

lint:
	golangci-lint run ./...