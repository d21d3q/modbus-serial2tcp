# Modbus Serial to TCP Gateway

`modbus-serial2tcp` listens on a Modbus RTU serial line and forwards requests to Modbus TCP servers. Use it to expose RS232/RS485 devices to software that only speaks Modbus/TCP.

## Overview

- Acts as the inverse of [mbusd](https://github.com/3cky/mbusd): serial input is proxied to one or more TCP servers.
- Built in Go with minimal runtime dependencies.
- Works with common USB-to-serial adapters and physical serial ports.

## Features

- Forwards Modbus RTU frames to a TCP server and returns responses to the serial client.
- Configurable serial settings (baud rate, parity, timeout) and slave ID filtering.
- Optional verbose logging for troubleshooting.
- Pre-built binaries for multiple Linux architectures, or build from source.

## Installation

Grab a pre-built Linux binary from the [Releases page](https://github.com/d21d3q/modbus-serial2tcp/releases):

- `modbus-serial2tcp-amd64-linux` (x86_64)
- `modbus-serial2tcp-arm64-linux` (ARM64)
- `modbus-serial2tcp-arm-linux` (ARM)

Download the file that matches your platform, rename it if desired (e.g., `modbus-s2t`), and make it executable:

```bash
wget https://github.com/d21d3q/modbus-serial2tcp/releases/latest/download/modbus-serial2tcp-amd64-linux -O modbus-s2t
chmod +x modbus-s2t
```

## Usage

### Quick start

Forward requests from `/dev/ttyUSB0` at 9600 baud to a Modbus/TCP server running on `192.168.1.100:502` and filter for slave ID 1:

```bash
./modbus-s2t \
  --port /dev/ttyUSB0 \
  --speed 9600 \
  --tcp-server 192.168.1.100:502 \
  --slaveid 1
```

### Command-line options

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--port` | `-p` | Serial port path (e.g., `/dev/ttyUSB0`, `COM1`) | `/dev/ttyUSB0` | Yes |
| `--speed` | `-s` | Serial port baud rate | `9600` | Yes |
| `--tcp-server` | `-S` | TCP server address (`ip:port`) | – | Yes |
| `--slaveid` | `-i` | Modbus slave ID to filter (1–247) | `1` | Yes |
| `--parity` | `-P` | Serial parity (`N` none, `E` even, `O` odd) | `N` | No |
| `--timeout` | `-t` | Request timeout in seconds | `1` | No |
| `--verbose` | `-v` | Increase log verbosity (repeat for more detail) | `0` | No |

## Systemd service

For unattended use, install the included `modbus-s2t.service` unit:

1. Copy the binary to a system path and make it executable:

   ```bash
   sudo cp modbus-serial2tcp /usr/local/bin/
   sudo chmod +x /usr/local/bin/modbus-serial2tcp
   ```

2. Install the service file:

   ```bash
   sudo cp modbus-s2t.service /etc/systemd/system/
   ```

3. Edit `/etc/systemd/system/modbus-s2t.service` to match your serial device and TCP target. Example:

   ```ini
   ExecStart=/usr/local/bin/modbus-serial2tcp --port /dev/ttyUSB0 --speed 9600 --tcp-server 192.168.1.100:502 --slaveid 1
   ```

4. Enable and start the service:

   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable modbus-s2t.service
   sudo systemctl start modbus-s2t.service
   ```

5. Check status and logs:

   ```bash
   sudo systemctl status modbus-s2t.service
   sudo journalctl -u modbus-s2t.service -f
   ```

## Build from source

Requirements:

- Go 1.19 or later
- [just](https://github.com/casey/just) (optional, for using the provided recipes)

```bash
git clone https://github.com/d21d3q/modbus-serial2tcp.git
cd modbus-serial2tcp

# Build using just (produces binaries for common targets)
just build-all

# Or build manually
go build -o modbus-serial2tcp .
```

## Acknowledgments

- [mbserver](https://github.com/tbrandon/mbserver)
- [modbus](https://github.com/simonvetter/modbus)

## License

MIT License. See [LICENSE](LICENSE).
