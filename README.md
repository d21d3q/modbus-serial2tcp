# Modbus Serial to TCP Gateway

A lightweight, efficient Modbus RTU (Serial) to TCP gateway written in Go. This tool bridges the gap between Modbus RTU devices and TCP/IP networks, enabling integration with contemporary industrial automation systems.

## Overview

This gateway performs the **opposite function** of the popular [mbusd](https://github.com/3cky/mbusd) tool:
- **mbusd**: Modbus TCP → Serial (acts as a TCP server, forwards to serial devices)
- **modbus-serial2tcp**: Modbus Serial → TCP (acts as a serial listener, forwards to TCP servers)

## Features

- 🔄 **Bidirectional Communication**: Seamlessly forwards Modbus RTU frames to TCP servers and responses back to serial devices
- 📡 **Serial Port Support**: Compatible with various serial interfaces (USB-to-Serial, RS232, RS485)
- ⚙️ **Configurable Parameters**: Flexible baud rate, parity, timeout, and slave ID filtering
- 🛡️ **Robust Error Handling**: Automatic reconnection and graceful error recovery (eg on USB removal)
- 🚀 **Cross-Platform**: Pre-built binaries for multiple Linux architectures

## Use Cases

- **Industrial IoT Integration**: Connect legacy Modbus RTU devices to cloud-based SCADA systems
- **Protocol Translation**: Bridge between serial-based PLCs and TCP-based HMI systems
- **Remote Monitoring**: Enable remote access to serial Modbus devices through network infrastructure
- **System Modernization**: Upgrade legacy industrial networks without replacing existing hardware

## Installation

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/username/modbus-serial2tcp/releases):

```bash
# For ARM Linux (Raspberry Pi, etc.)
wget https://github.com/d21d3q/modbus-serial2tcp/releases/latest/download/modbus-serial2tcp-arm-linux -O modbus-s2t
chmod +x modbus-s2t

# For ARM64 Linux
wget https://github.com/d21d3q/modbus-serial2tcp/releases/latest/download/modbus-serial2tcp-arm64-linux -O modbus-s2t
chmod +x modbus-s2t

# For x86_64 Linux
wget https://github.com/d21d3q/modbus-serial2tcp/releases/latest/download/modbus-serial2tcp-amd64-linux -O modbus-s2t
chmod +x modbus-s2t
```

### Build from Source

Requirements:
- Go 1.19 or later
- [just](https://github.com/casey/just) (optional, for using build recipes)

```bash
# Clone the repository
git clone https://github.com/username/modbus-serial2tcp.git
cd modbus-serial2tcp

# Build using just (recommended)
just build-all

# Or build manually with Go
go build -o modbus-serial2tcp .
```

## Usage

### Basic Usage

```bash
./modbus-s2t \
  --port /dev/ttyUSB0 \
  --speed 9600 \
  --tcp-server 192.168.1.100:502 \
  --slaveid 1
```

### Command Line Options

| Flag | Short | Description | Default | Required |
|------|-------|-------------|---------|----------|
| `--port` | `-p` | Serial port path (e.g., `/dev/ttyUSB0`, `COM1`) | `/dev/ttyUSB0` | ✅ |
| `--speed` | `-s` | Serial port baud rate | `9600` | ✅ |
| `--tcp-server` | `-S` | TCP server address (format: `ip:port`) | - | ✅ |
| `--slaveid` | `-i` | Modbus Slave ID to filter (1-247) | `1` | ✅ |
| `--parity` | `-P` | Serial parity (`N`=none, `E`=even, `O`=odd) | `N` | ❌ |
| `--timeout` | `-t` | Request timeout in seconds | `1` | ❌ |
| `--verbose` | `-v` | Verbosity level (use multiple times: `-vvv`) | `0` | ❌ |

### Examples

#### Basic RS485 to TCP Gateway
```bash
# Forward Modbus RTU on RS485 to a TCP server
./modbus-s2t \
  --port /dev/ttyUSB0 \
  --speed 9600 \
  --parity E \
  --tcp-server 10.0.0.50:502 \
  --slaveid 5 \
  --verbose
```


## Systemd Service Installation

For running as daemon, use the included systemd service:

1. **Copy the binary to system location:**
   ```bash
   sudo cp modbus-serial2tcp /usr/local/bin/
   sudo chmod +x /usr/local/bin/modbus-serial2tcp
   ```

2. **Install the service file:**
   ```bash
   sudo cp modbus-s2t.service /etc/systemd/system/
   ```

3. **Configure the service:**
   Edit `/etc/systemd/system/modbus-s2t.service` to match your setup:
   ```ini
   ExecStart=/usr/local/bin/modbus-serial2tcp --port /dev/ttyUSB0 --speed 9600 --tcp-server 192.168.1.100:502 --slaveid 1
   ```

4. **Enable and start the service:**
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable modbus-s2t.service
   sudo systemctl start modbus-s2t.service
   ```

5. **Check service status:**
   ```bash
   sudo systemctl status modbus-s2t.service
   sudo journalctl -u modbus-s2t.service -f  # Follow logs
   ```

## Acknowledgments

This project is built upon the excellent work of:

- **[mbserver](https://github.com/tbrandon/mbserver)** 
- **[modbus](https://github.com/simonvetter/modbus)** 

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

**Made with ❤️ for the Industrial IoT community**
