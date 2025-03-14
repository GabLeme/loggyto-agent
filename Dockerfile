FROM golang:1.21 as builder
WORKDIR /app

COPY . .

RUN go mod tidy && go build -o log-agent cmd/main.go

FROM debian:bullseye-slim
WORKDIR /root/

COPY --from=builder /app/log-agent .

CMD ["./log-agent"]