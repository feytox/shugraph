.PHONY: up down logs

export VERSION ?= $(shell git describe --tags --always --dirty)

.env:
	cp .env.example .env

up: .env
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f
