# syntax=docker/dockerfile:1.7
FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
ARG TARGETOS
ARG TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags='-s -w' -o /out/editorial ./cmd/server

FROM alpine:3.21
RUN addgroup -S editorial && adduser -S -G editorial -u 10001 editorial
WORKDIR /app
COPY --from=build /out/editorial /app/editorial
COPY migrations /app/migrations
COPY api/openapi /app/api/openapi
USER editorial
EXPOSE 8080
ENTRYPOINT ["/app/editorial"]
