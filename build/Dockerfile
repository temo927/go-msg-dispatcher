FROM golang:1.22.2-alpine AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/server

CMD ["./server"]
