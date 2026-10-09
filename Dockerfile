FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git nodejs npm
RUN npm install --global bun

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 \
	&& go install github.com/a-h/templ/cmd/templ@v0.3.1020

COPY . .

RUN sqlc generate
RUN templ generate
RUN bun install --frozen-lockfile && bun run build

ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_TIME=unknown
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -trimpath -ldflags="-s -w -X github.com/gracchi-stdio/castogo/internal/version.Version=${VERSION} -X github.com/gracchi-stdio/castogo/internal/version.Commit=${COMMIT} -X github.com/gracchi-stdio/castogo/internal/version.BuildTime=${BUILD_TIME}" -o /castogo ./cmd/server
RUN go build -trimpath -ldflags="-s -w -X github.com/gracchi-stdio/castogo/internal/version.Version=${VERSION} -X github.com/gracchi-stdio/castogo/internal/version.Commit=${COMMIT} -X github.com/gracchi-stdio/castogo/internal/version.BuildTime=${BUILD_TIME}" -o /castogo-cli ./cmd/cli

FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata ffmpeg
RUN addgroup -S castogo && adduser -S -G castogo castogo

WORKDIR /app

COPY --from=builder /castogo /app/castogo
COPY --from=builder /castogo-cli /app/castogo-cli
COPY --from=builder /app/public /app/public
COPY --from=builder /app/sql/migrations /app/sql/migrations
COPY --chmod=755 scripts/entrypoint.sh /app/entrypoint.sh

EXPOSE 8080
STOPSIGNAL SIGTERM
HEALTHCHECK --interval=15s --timeout=5s --start-period=10s --retries=3 CMD wget -q -O /dev/null http://127.0.0.1:${PORT:-8080}/healthz || exit 1

USER castogo

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/castogo"]
