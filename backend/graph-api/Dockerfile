
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/graph-api ./cmd

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/graph-api /usr/local/bin/graph-api
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/graph-api"]
CMD ["serve"]
