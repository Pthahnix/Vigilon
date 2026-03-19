# Vigilon

AI-powered GPU watchdog for shared lab servers. Monitors usage, evaluates task priority, and enforces fair resource allocation across users.

## Features

- Background daemon monitors GPU usage every 10 minutes
- 3-tier priority system: P0 (1 GPU), P1 (2 GPUs), P2 (3 GPUs)
- AI-powered resource request review via `vigil apply`
- Automatic violation detection with 30-minute grace period
- Idle detection: auto-reclaim after 3 consecutive idle cycles
- Timeout tolerance: expired but active tasks are not killed
- Duration buffer: AI-estimated time × 1.5 for safety margin
- Terminal warnings (`wall`) and audit logging

## Quick Start

```bash
# Build
go build -o vigil ./cmd/vigil/

# Install
sudo ./scripts/install.sh

# Edit config
sudo vim /etc/vigilon/config.yaml

# Apply for GPU priority upgrade
vigil apply my-request.md

# Check status
vigil status

# View your history
vigil log

# Release priority when done
vigil release
```

## Admin Commands

All admin commands require passphrase authentication.

```bash
vigil admin start          # Start daemon
vigil admin stop           # Stop daemon
vigil admin grant <user> <P0|P1|P2> [duration]  # Manual priority set
vigil admin reset <user>   # Reset user to P0
vigil admin purge          # Reset all users to P0
vigil admin check          # Run one detection cycle now
```

## Priority Levels

| Level | Name | Max GPUs | Use Case |
| ----- | ---- | -------- | -------- |
| P0 | Normal | 1 | Daily debugging, small experiments |
| P1 | Boost | 2 | Project sprints, model training |
| P2 | Urgent | 3 (all) | Paper deadlines, urgent experiments |

## License

See [LICENSE](LICENSE).
