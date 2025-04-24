FROM --platform=$BUILDPLATFORM golang:1.24 AS builder
WORKDIR /app

# ✅ Instala as dependências necessárias para CGO e systemd
RUN apt-get update && \
    apt-get install -y gcc libsystemd-dev ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY . .

ARG TARGETOS
ARG TARGETARCH

# ✅ CGO habilitado e variáveis de compilação
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV CGO_ENABLED=1

RUN go build -o log-agent ./cmd/main.go

# Imagem final minimalista
FROM --platform=$TARGETPLATFORM debian:bookworm-slim
WORKDIR /root/

# ✅ Precisa do runtime libsystemd
RUN apt-get update && \
    apt-get install -y libsystemd0 ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/log-agent .

CMD ["./log-agent"]
