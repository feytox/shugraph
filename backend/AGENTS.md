# backend

- Клиент graph-api в graph-admin-api генерируется из `graph-api/openapi.yaml`. Если меняешь API graph-api, обнови спеку (`make openapi` в graph-api), перегенерируй клиент (`make gen` в graph-admin-api) и закоммить оба изменения.
- Интеграционные тесты под build-тегом `integration`, Postgres через testcontainers (нужен Docker). Обычный `go test` их не запускает, используй `make integration-test`.
