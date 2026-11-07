FROM golang:1.26.10-alpine AS base

WORKDIR /app

RUN apk add --no-cache ca-certificates git

ENV GOWORK=/app/go.work

COPY go.work ./

COPY common/go.mod common/go.sum ./common/
COPY protos/go.mod protos/go.sum ./protos/
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY auth-service/go.mod auth-service/go.sum ./auth-service/

RUN cd gateway && go mod download -x

COPY common ./common
COPY protos ./protos
COPY gateway ./gateway
COPY auth-service ./auth-service

FROM base AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app/gateway

EXPOSE 8080

CMD ["air", "-c", "./docker/app/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/app

FROM alpine:3.22 AS prod

WORKDIR /app/gateway

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway ./bin/gateway

EXPOSE 8080

CMD ["./bin/gateway"]