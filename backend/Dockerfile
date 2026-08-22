FROM golang:1.26-alpine AS builder

RUN go install github.com/air-verse/air@v1.67.4

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

CMD ["air", "-c", ".air.toml"]
