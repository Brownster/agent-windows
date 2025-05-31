# Windows Agent Collector

A lightweight Windows metrics collector specifically designed for WebRTC voice quality troubleshooting. This agent pushes metrics to a Prometheus Push Gateway and includes enhanced network interface detection for correlation with WebRTC statistics.

## Features

- **🚀 Lightweight**: ~80% smaller than full windows_exporter
- **📡 Push Gateway Integration**: Pushes metrics instead of hosting HTTP endpoint  
- **🔗 WebRTC Correlation**: Agent ID labeling for correlation with WebRTC stats
- **🌐 Enhanced Network Detection**: Detailed interface type detection (ethernet, wifi, cellular)
- **⚡ Minimal Overhead**: Focused on essential metrics only
- **🛠️ Windows Service**: Runs as a background Windows service

## Quick Start

### Download and Install
```powershell
# Install as Windows service
.\windows-agent-collector.exe --agent-id=agent_001 --push.gateway-url=http://pushgateway:9091 install

# Start the service
sc start windows_agent_collector
```

### Basic Usage
```powershell
# Basic usage
.\windows-agent-collector.exe --agent-id=agent_001 --push.gateway-url=http://pushgateway:9091

# With authentication
.\windows-agent-collector.exe --agent-id=agent_001 --push.gateway-url=http://pushgateway:9091 --push.username=user --push.password=pass

# Using config file
.\windows-agent-collector.exe --config.file=config.yaml
```

## Metrics Collected

### 🖥️ CPU Metrics
- CPU utilization percentage
- CPU frequency information
- Per-core statistics

### 💾 Memory Metrics
- Available memory bytes
- Used memory bytes  
- Memory utilization percentage

### 🌐 Network Metrics
- Bytes sent/received per interface
- **Network interface type** (ethernet/wifi/cellular/vpn) - *Key for WebRTC correlation*
- Interface operational status
- Current bandwidth information

### 💽 Pagefile/Swap Metrics
- Pagefile usage and availability
- Swap utilization percentage

## Configuration

Create a `config.yaml` file (see `config-example.yaml`):

```yaml
push_gateway:
  url: "http://pushgateway.example.com:9091"
  username: "monitoring_user"
  password: "secret_password"
  interval: "30s"

agent:
  id: "agent_001"

collectors:
  enabled: ["cpu", "memory", "net", "pagefile"]
```

## WebRTC Correlation

The agent includes an `agent_id` label on all metrics, enabling correlation with WebRTC statistics:

```prometheus
# Example metrics with agent_id label
windows_cpu_time_total{agent_id="agent_001",mode="idle"} 12345
windows_memory_available_bytes{agent_id="agent_001"} 8589934592
windows_net_nic_info{agent_id="agent_001",nic="WiFi",interface_type="wifi"} 1
```

## Network Interface Types

Enhanced detection for WebRTC compatibility:
- `ethernet` - Wired Ethernet connections
- `wifi` - 802.11 wireless connections  
- `cellular` - Mobile/cellular connections
- `vpn` - VPN tunnel interfaces
- `loopback` - Loopback interfaces
- `unknown` - Unidentified interface types

## Building from Source

```bash
# Clone repository
git clone https://github.com/Brownster/agent-windows.git
cd agent-windows

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o windows-agent-collector.exe ./cmd/agent
```

## Documentation

- [Configuration Example](config-example.yaml)
- [CPU Collector](docs/collector.cpu.md)
- [Memory Collector](docs/collector.memory.md)
- [Network Collector](docs/collector.net.md)
- [Pagefile Collector](docs/collector.pagefile.md)

## Differences from windows_exporter

This agent is purpose-built for WebRTC troubleshooting:

| Feature | windows_exporter | windows-agent-collector |
|---------|-----------------|-------------------------|
| **Architecture** | HTTP server (pull) | Push Gateway (push) |
| **Collectors** | 50+ collectors | 4 essential collectors |
| **Binary Size** | ~50MB | ~10MB |
| **Memory Usage** | 200MB default | 50MB default |
| **WebRTC Integration** | None | Built-in agent correlation |
| **Network Detection** | Basic | Enhanced interface types |
| **Deployment** | HTTP endpoint | Push to gateway |

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

---

*Built specifically for WebRTC voice quality monitoring and troubleshooting scenarios.*