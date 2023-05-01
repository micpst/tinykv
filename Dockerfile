# syntax = docker/dockerfile:1
# --- Base image ---
FROM golang:1.20-buster AS base

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# --- Development image ---
FROM golangci/golangci-lint:v1.51 AS dev

WORKDIR /app

COPY --from=base /go/pkg/mod /go/pkg/mod

ENTRYPOINT ["tail", "-f", "/dev/null"]

# --- Development watch image ---
FROM cosmtrek/air AS watch

RUN apt update && \
    apt install -y nginx

# --- Build image ---
FROM base AS build

COPY api ./api
COPY cmd ./cmd
COPY pkg ./pkg

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -o ./bin/master ./cmd/master

# --- Master release image ---
FROM gcr.io/distroless/base-debian11:debug AS master

WORKDIR /app

SHELL ["/busybox/sh", "-c"]

RUN mkdir indexdb && \
    chown -R nonroot:nonroot indexdb

COPY --from=build /app/bin/master .

USER nonroot:nonroot

EXPOSE 3000

ENTRYPOINT ["./master", "--db", "indexdb", "--port", "3000"]

CMD ["--volumes", ""]

# --- Volume node release image ---
FROM nginx:1.23 AS volume

COPY volume/setup.sh .

RUN chmod +x setup.sh

CMD ["./setup.sh", "-g", "daemon off;"]
