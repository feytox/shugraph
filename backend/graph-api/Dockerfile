FROM golang:1.27-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build  -o "/build/graph-api" "./cmd"

FROM alpine:3.21

WORKDIR /app

COPY --from=builder /build/graph-api /app/graph-api
EXPOSE 8080

ENTRYPOINT ["/app/graph-api"]