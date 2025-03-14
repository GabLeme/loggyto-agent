FROM --platform=$BUILDPLATFORM golang:1.24 as builder
WORKDIR /app

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH CGO_ENABLED=0 go build -o log-agent main.go

FROM --platform=$TARGETPLATFORM debian:bullseye-slim
WORKDIR /root/

COPY --from=builder /app/log-agent .

CMD ["./log-agent"]
