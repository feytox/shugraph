FROM golang:1.27-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o "/build/graph-api" "./cmd"

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /build/graph-api /app/graph-api
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/graph-api"]