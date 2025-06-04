# Etapa de build
FROM --platform=linux/amd64 golang:1.24-bookworm AS builder

WORKDIR /app

RUN apt-get update && \
    apt-get install -y build-essential libsystemd-dev ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o log-agent ./cmd/main.go

# Etapa final
FROM --platform=linux/amd64 debian:bookworm-slim

WORKDIR /root/

RUN apt-get update && \
    apt-get install -y libsystemd0 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/log-agent .

CMD ["./log-agent"]
