# Installation Guide

## Requirements

- Go 1.24+ (for building from source)
- Docker 20.10+ (optional, for containerized deployment)
- GCC/Clang (on worker nodes for native compilation)

## Installation Methods

### 1. Pre-built Binaries

Download the latest release for your platform:

**Linux (amd64)**
```bash
curl -sSL https://github.com/h3nr1-d14z/hybridgrid/releases/latest/download/hybridgrid_linux_amd64.tar.gz | tar xz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/
```

**Linux (arm64)**
```bash
curl -sSL https://github.com/h3nr1-d14z/hybridgrid/releases/latest/download/hybridgrid_linux_arm64.tar.gz | tar xz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/
```

**macOS (Intel)**
```bash
curl -sSL https://github.com/h3nr1-d14z/hybridgrid/releases/latest/download/hybridgrid_darwin_amd64.tar.gz | tar xz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/
```

**macOS (Apple Silicon)**
```bash
curl -sSL https://github.com/h3nr1-d14z/hybridgrid/releases/latest/download/hybridgrid_darwin_arm64.tar.gz | tar xz
sudo mv hg-coord hg-worker hgbuild /usr/local/bin/
```

**Windows**
Download `hybridgrid_windows_amd64.zip` from releases and extract to your PATH.

### 2. Docker Images

```bash
# Pull images
docker pull ghcr.io/h3nr1-d14z/hybridgrid/hg-coord:latest
docker pull ghcr.io/h3nr1-d14z/hybridgrid/hg-worker:latest
docker pull ghcr.io/h3nr1-d14z/hybridgrid/hgbuild:latest

# Run coordinator
docker run -d --name hg-coord \
  -p 9000:9000 -p 8080:8080 \
  ghcr.io/h3nr1-d14z/hybridgrid/hg-coord:latest

# Run worker
docker run -d --name hg-worker \
  -e HG_COORDINATOR_ADDR=host.docker.internal:9000 \
  ghcr.io/h3nr1-d14z/hybridgrid/hg-worker:latest
```

### 3. Docker Compose

```bash
git clone https://github.com/h3nr1-d14z/hybridgrid.git
cd hybridgrid
docker compose up -d
```

### 4. From Source

```bash
# Clone
git clone https://github.com/h3nr1-d14z/hybridgrid.git
cd hybridgrid

# Build all binaries
go build -o bin/hg-coord ./cmd/hg-coord
go build -o bin/hg-worker ./cmd/hg-worker
go build -o bin/hgbuild ./cmd/hgbuild

# Install
sudo cp bin/* /usr/local/bin/
```

## Post-Installation

### Verify Installation

```bash
hg-coord --version
hg-worker --version
hgbuild --version
```

### Create Config Directory

```bash
mkdir -p ~/.hybridgrid
```

### Start Services

**Option A: Systemd (Linux)**

Create `/etc/systemd/system/hg-coord.service`:
```ini
[Unit]
Description=Hybrid-Grid Coordinator
After=network.target

[Service]
Type=simple
User=hybridgrid
ExecStart=/usr/local/bin/hg-coord serve
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl enable --now hg-coord
```

**Option B: Manual**

```bash
# Terminal 1: Start coordinator
hg-coord serve

# Terminal 2: Start worker
hg-worker serve --coordinator=localhost:9000
```

## Network Configuration

### Firewall Rules

Open these ports:
- **9000/tcp** - Coordinator gRPC (required)
- **8080/tcp** - Coordinator HTTP/Dashboard (optional)
- **50052/tcp** - Worker gRPC (internal)
- **9090/tcp** - Worker metrics (optional)
- **5353/udp** - mDNS discovery (LAN only)

### mDNS Discovery

For automatic LAN discovery, ensure:
1. Workers and coordinator are on the same subnet
2. Multicast traffic is allowed (UDP 5353)
3. No IGMP snooping blocking mDNS

## Next Steps

- [Configuration Reference](configuration.md) - Customize settings
- [Architecture Overview](architecture.md) - Understand the system
