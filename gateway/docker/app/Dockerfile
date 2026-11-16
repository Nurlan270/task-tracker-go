FROM golang:1.26.10-alpine AS base

WORKDIR /app

RUN apk add --no-cache ca-certificates git

ENV GOWORK=/app/go.work

# Workspace
COPY go.work ./

# All module metadata required by go.work
COPY common/go.mod common/
COPY protos/go.mod protos/
COPY gateway/go.mod gateway/
COPY task-service/go.mod task-service/
COPY user-service/go.mod user-service/

# Download dependencies
RUN cd gateway && go mod download -x

# Actual dependencies
COPY common/ common/
COPY protos/ protos/

# Current service
COPY gateway/ gateway/

FROM base AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app/gateway

EXPOSE 8080

CMD ["air", "-c", "./docker/app/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/gateway ./cmd/gateway

FROM alpine:3.22 AS prod

WORKDIR /app/gateway

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/gateway ./bin/gateway

EXPOSE 8080

CMD ["./bin/gateway"]