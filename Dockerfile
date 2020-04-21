FROM golang:1.13-alpine as module_base

RUN addgroup -S caspr && \
  adduser -S caspr -G caspr --home /caspr

WORKDIR /caspr
USER caspr

# WORKDIR /go/src/application-service

# Force the go compiler to use modules
ENV GO111MODULE=on

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

FROM module_base as builder

COPY . .

RUN go build ./...
