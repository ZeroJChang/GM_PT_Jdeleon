FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o loans-service ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/loans-service .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./loans-service"]