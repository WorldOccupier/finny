# syntax=docker/dockerfile:1
FROM golang:1.26-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . .

EXPOSE 8080

CMD ["go", "run", "./cmd/finny"]
