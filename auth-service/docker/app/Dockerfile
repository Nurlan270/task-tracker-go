FROM golang:1.26.3-alpine AS base

WORKDIR /app

RUN apk add --no-cache ca-certificates git

COPY go.work ./

COPY common/go.mod common/go.sum ./common/
COPY auth-service/go.mod auth-service/go.sum ./auth-service/

RUN cd auth-service && go mod download

COPY common ./common
COPY auth-service ./auth-service

FROM base AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app/auth-service

EXPOSE 7070

CMD ["air", "-c", "./docker/app/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM alpine:3.22 AS prod

WORKDIR /app/auth_service

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/auth_service ./bin/auth_service

EXPOSE 7070

CMD ["./bin/auth_service"]