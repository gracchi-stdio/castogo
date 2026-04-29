# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git nodejs yarn

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
RUN go run github.com/a-h/templ/cmd/templ generate
RUN yarn install && yarn build

RUN CGO_ENABLED=0 GOOS=linux go build -o /castogo ./cmd/server

# Runtime stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata ffmpeg

WORKDIR /app

COPY --from=builder /castogo /app/castogo
COPY --from=builder /app/public /app/public

EXPOSE 8080

CMD ["/app/castogo"]
