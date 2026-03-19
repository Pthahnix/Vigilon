#!/bin/bash
set -e
echo "Building Vigilon..."
go build -o vigil ./cmd/vigil/
echo "Installing binary..."
sudo cp vigil /usr/local/bin/
echo "Creating directories..."
sudo mkdir -p /etc/vigilon /var/log/vigilon
sudo cp configs/config.example.yaml /etc/vigilon/config.yaml
echo "Installing systemd service..."
sudo cp vigilon.service /etc/systemd/system/
sudo systemctl daemon-reload
echo "Done. Edit /etc/vigilon/config.yaml then run:"
echo "  sudo systemctl enable --now vigilon"
