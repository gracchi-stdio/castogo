# Build stage
FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
RUN go run github.com/a-h/templ/cmd/templ generate

RUN CGO_ENABLED=0 GOOS=linux go build -o /podlog ./cmd/server

# Runtime stage
FROM alpine:3.23

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /podlog /app/podlog
COPY static /app/static

EXPOSE 8080

CMD ["/app/podlog"]
