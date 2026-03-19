# Vigilon

AI-powered GPU watchdog for shared lab servers. Monitors usage, evaluates task priority, and enforces fair resource allocation across users.

## Features

- Background daemon monitors GPU usage every 10 minutes
- 3-tier priority system: P0 (1 GPU), P1 (2 GPUs), P2 (3 GPUs)
- AI-powered resource request review via `vigil apply`
- Automatic violation detection with 30-minute grace period
- Terminal warnings (`wall`) and audit logging

## Quick Start

```bash
# Build
go build -o vigil ./cmd/vigil/

# Install
sudo ./scripts/install.sh

# Edit config
sudo vim /etc/vigilon/config.yaml

# Start daemon
sudo systemctl enable --now vigilon

# Apply for GPU priority upgrade
vigil apply my-request.md

# Check status
vigil status

# View your history
vigil log
```

## Priority Levels

| Level | Name | Max GPUs | Use Case |
| ----- | ---- | -------- | -------- |
| P0 | Normal | 1 | Daily debugging, small experiments |
| P1 | Boost | 2 | Project sprints, model training |
| P2 | Urgent | 3 (all) | Paper deadlines, urgent experiments |

## License

See [LICENSE](LICENSE).
