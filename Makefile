APP_NAME=log-agent
CMD_DIR=./cmd

build-linux:
	docker run --rm --platform=linux/amd64 -v "$(PWD)":/app -w /app golang:1.24-bookworm bash -c '\
		apt-get update && \
		apt-get install -y build-essential libsystemd-dev && \
		CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(APP_NAME) $(CMD_DIR) \
	'

clean:
	rm -f $(APP_NAME)
