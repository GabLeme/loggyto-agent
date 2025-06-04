#!/bin/bash

set -e

# Help
usage() {
  echo "Usage: $0 -e <endpoint> -k <api_key> -s <api_secret> [-n] [-i <ignored_containers>] [-N <ignored_namespaces>]"
  echo "  -e   Endpoint (ex: https://loggyto-endpoint.com)"
  echo "  -k   API Key"
  echo "  -s   API Secret"
  echo "  -n   Disable TLS verification (optional)"
  echo "  -i   Comma-separated list of containers to ignore (optional)"
  echo "  -N   Comma-separated list of namespaces to ignore (optional)"
  exit 1
}

# Argumentos
NO_VERIFY="false"
IGNORED_CONTAINERS=""
IGNORED_NAMESPACES=""

while getopts ":e:k:s:ni:N:" opt; do
  case ${opt} in
    e ) ENDPOINT=$OPTARG ;;
    k ) API_KEY=$OPTARG ;;
    s ) API_SECRET=$OPTARG ;;
    n ) NO_VERIFY="true" ;;
    i ) IGNORED_CONTAINERS=$OPTARG ;;
    N ) IGNORED_NAMESPACES=$OPTARG ;;
    \? ) usage ;;
  esac
done

if [ -z "$ENDPOINT" ] || [ -z "$API_KEY" ] || [ -z "$API_SECRET" ]; then
  usage
fi

INSTALL_DIR="/opt/loggyto"
BIN_PATH="$INSTALL_DIR/log-agent"
SERVICE_FILE="/etc/systemd/system/loggyto-agent.service"

echo "[INFO] Criando diretório de instalação em $INSTALL_DIR..."
sudo mkdir -p "$INSTALL_DIR"

echo "[INFO] Baixando binário do log-agent de GitHub..."
sudo curl -sSL https://raw.githubusercontent.com/GabLeme/loggyto-agent/main/log-agent -o "$BIN_PATH"
sudo chmod +x "$BIN_PATH"

echo "[INFO] Escrevendo arquivo de serviço systemd..."
cat <<EOF | sudo tee "$SERVICE_FILE" > /dev/null
[Unit]
Description=Loggyto Agent
After=network.target

[Service]
ExecStart=$BIN_PATH
Restart=always
Environment=LOGGYTO_ENDPOINT=$ENDPOINT
Environment=LOGGYTO_API_KEY=$API_KEY
Environment=LOGGYTO_API_SECRET=$API_SECRET
Environment=LOGGYTO_NO_VERIFY=$NO_VERIFY
EOF

# Adiciona IGNORED_CONTAINERS se definido
if [ -n "$IGNORED_CONTAINERS" ]; then
  echo "Environment=LOGGYTO_IGNORED_CONTAINERS=$IGNORED_CONTAINERS" | sudo tee -a "$SERVICE_FILE" > /dev/null
fi

# Adiciona IGNORED_NAMESPACES se definido
if [ -n "$IGNORED_NAMESPACES" ]; then
  echo "Environment=LOGGYTO_IGNORED_NAMESPACES=$IGNORED_NAMESPACES" | sudo tee -a "$SERVICE_FILE" > /dev/null
fi

echo "[INFO] Recarregando systemd e iniciando agente..."
sudo systemctl daemon-reexec
sudo systemctl daemon-reload
sudo systemctl enable loggyto-agent
sudo systemctl start loggyto-agent

echo "[INFO] Loggyto Agent instalado e rodando!"
