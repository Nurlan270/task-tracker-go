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
RUN cd user-service && go mod download -x

# Actual dependencies
COPY common/ common/
COPY protos/ protos/

# Current service
COPY user-service/ user-service/

FROM base AS dev

RUN go install github.com/air-verse/air@latest

WORKDIR /app/user-service

EXPOSE 7070

CMD ["air", "-c", "./docker/app/.air.toml"]

FROM base AS builder

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/user-service ./cmd/user-service

FROM alpine:3.22 AS prod

WORKDIR /app/user-service

RUN apk add --no-cache ca-certificates

COPY --from=builder /bin/user-service ./bin/user-service

EXPOSE 7070

CMD ["./bin/user-service"]