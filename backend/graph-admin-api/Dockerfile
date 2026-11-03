
FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
ARG VERSION=dev
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/graph-admin-api ./cmd

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/graph-admin-api /usr/local/bin/graph-admin-api
EXPOSE 8081
ENTRYPOINT ["/usr/local/bin/graph-admin-api"]
CMD ["serve"]
