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
RUN cd task-service && go mod download -x

# Actual dependencies
COPY common/ common/
COPY protos/ protos/

# Current service
COPY task-service/ task-service/

FROM base AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app/task-service

EXPOSE 7070

CMD ["air", "-c", "./docker/app/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/task-service ./cmd/task-service

FROM alpine:3.22 AS prod

WORKDIR /app/task-service

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/task-service ./bin/task-service

EXPOSE 7070

CMD ["./bin/task-service"]