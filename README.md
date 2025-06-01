# Windows Agent Collector

A lightweight Windows metrics collector specifically designed for WebRTC voice quality troubleshooting. This agent pushes metrics to a Prometheus Push Gateway and includes enhanced network interface detection for correlation with WebRTC statistics.

> **🔗 Sister Application**: Pairs with the [WebRTC Chrome Extension](https://github.com/Brownster/agent-webrtc) for comprehensive voice quality monitoring. Both applications use the same `agent_id` for seamless correlation in Grafana dashboards.

## Features

- **🚀 Lightweight**: ~80% smaller than full windows_exporter
- **📡 Push Gateway Integration**: Pushes metrics instead of hosting HTTP endpoint  
- **🔗 WebRTC Correlation**: Agent ID labeling for correlation with WebRTC stats
- **🌐 Enhanced Network Detection**: Detailed interface type detection (ethernet, wifi, cellular)
- **⚡ Minimal Overhead**: Focused on essential metrics only
- **🛠️ Windows Service**: Runs as a background Windows service

## Quick Start

### Download and Install
Download the latest release from [GitHub Releases](https://github.com/Brownster/agent-windows/releases).

#### Basic Installation
```powershell
# Install as Windows service (no authentication)
.\windows-agent-collector.exe --agent-id=agent_001 --push.gateway-url=http://pushgateway:9091 install

# Start the service
sc start windows_agent_collector
```

#### Installation with Authentication
```powershell
# Install with Push Gateway authentication
.\windows-agent-collector.exe `
  --agent-id=agent_001 `
  --push.gateway-url=http://pushgateway:9091 `
  --push.username=monitoring_user `
  --push.password=secure_password `
  --push.interval=30s `
  install

# Start the service
sc start windows_agent_collector
```

#### Secure Installation (Recommended)
1. **Create configuration file** (`config.yaml`):
```yaml
push_gateway:
  url: "http://pushgateway.example.com:9091"
  username: "monitoring_user"
  password: "secure_password"
  interval: "30s"

agent:
  id: "agent_001"

collectors:
  enabled: ["cpu", "memory", "net", "pagefile"]

log:
  level: "info"
```

2. **Set file permissions** (Administrator only):
```powershell
icacls config.yaml /inheritance:d
icacls config.yaml /grant:r "Administrators:(R)"
icacls config.yaml /remove "Users"
```

3. **Install and start service**:
```powershell
.\windows-agent-collector.exe --config.file=config.yaml install
sc start windows_agent_collector
```

### Service Management
```powershell
# Check service status
sc query windows_agent_collector

# Stop service
sc stop windows_agent_collector

# Restart service
sc stop windows_agent_collector && sc start windows_agent_collector

# Uninstall service
sc stop windows_agent_collector
.\windows-agent-collector.exe uninstall
```

### Basic Usage (Non-Service)
```powershell
# Run directly (for testing)
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

### Command Line Options

| Flag | Description | Required | Default |
|------|-------------|----------|---------|
| `--agent-id` | Unique agent identifier for correlation | ✅ Yes | - |
| `--push.gateway-url` | Prometheus Push Gateway URL | ✅ Yes | - |
| `--push.username` | Basic auth username for push gateway | No | - |
| `--push.password` | Basic auth password for push gateway | No | - |
| `--push.interval` | Push frequency (e.g., 30s, 1m) | No | 30s |
| `--push.job-name` | Prometheus job name | No | windows_agent |
| `--collectors.enabled` | Comma-separated list of collectors | No | cpu,memory,net,pagefile |
| `--config.file` | Path to YAML configuration file | No | - |
| `--log.level` | Log level (debug, info, warn, error) | No | info |
| `--log.format` | Log format (text, json) | No | text |

### Configuration File

Create a `config.yaml` file for more complex configurations:

```yaml
# Basic configuration
push_gateway:
  url: "http://pushgateway.example.com:9091"
  username: "monitoring_user"
  password: "secret_password"
  interval: "30s"
  job_name: "windows_agent"
  timeout: "10s"

agent:
  id: "agent_001"  # Must match Chrome extension

collectors:
  enabled: ["cpu", "memory", "net", "pagefile"]

log:
  level: "info"
  format: "text"
  file: "C:\\logs\\agent-collector.log"  # Optional log file

# Advanced service configuration  
service:
  name: "windows_agent_collector"
  display_name: "Windows Agent Collector"
  description: "Lightweight Windows metrics collector for WebRTC troubleshooting"
```

### Environment Variables

You can override configuration using environment variables:

```powershell
$env:AGENT_ID = "agent_001"
$env:PUSH_GATEWAY_URL = "http://pushgateway:9091"
$env:PUSH_GATEWAY_USERNAME = "monitoring_user"
$env:PUSH_GATEWAY_PASSWORD = "secure_password"
```

### Production Configuration Example

```yaml
# Production-ready configuration
push_gateway:
  url: "https://pushgateway.prod.company.com:9091"  # Use HTTPS
  username: "${PUSH_GATEWAY_USERNAME}"              # Environment variable
  password: "${PUSH_GATEWAY_PASSWORD}"              # Environment variable
  interval: "30s"
  job_name: "windows_agent_prod"
  timeout: "15s"

agent:
  id: "${AGENT_ID}"  # Environment variable

collectors:
  enabled: ["cpu", "memory", "net", "pagefile"]

log:
  level: "info"
  format: "json"  # Structured logging for production
  file: "C:\\ProgramData\\WindowsAgentCollector\\logs\\agent.log"

service:
  name: "windows_agent_collector"
  display_name: "Windows Agent Collector (Production)"
  description: "Production Windows metrics collector for WebRTC voice quality monitoring"
```

## WebRTC Integration & Correlation

### 🌐 Chrome Extension Integration

This Windows agent works seamlessly with the **[WebRTC Chrome Extension](https://github.com/Brownster/agent-webrtc)** to provide comprehensive voice quality monitoring:

- **Windows Agent**: Collects system metrics (CPU, memory, network interface types)
- **Chrome Extension**: Captures WebRTC statistics (audio quality, packet loss, jitter, bitrates)
- **Shared Agent ID**: Both use the same `agent_id` for correlation

### 📊 Grafana Dashboard Correlation

In Grafana, you can create unified dashboards combining metrics from both sources:

```prometheus
# Windows system metrics
windows_cpu_time_total{agent_id="agent_001",mode="idle"}
windows_memory_available_bytes{agent_id="agent_001"}
windows_net_nic_info{agent_id="agent_001",nic="WiFi",interface_type="wifi"}

# WebRTC metrics (from Chrome extension)
webrtc_audio_packets_lost{agent_id="agent_001"}
webrtc_audio_jitter{agent_id="agent_001"}
webrtc_connection_state{agent_id="agent_001",interface_type="wifi"}
```

### 🔍 Troubleshooting Example

Correlate network issues with call quality:
- High CPU usage + packet loss = resource contention
- WiFi interface + high jitter = wireless connectivity issues
- Memory pressure + audio dropouts = system performance impact

## Network Interface Types

Enhanced detection for WebRTC compatibility:
- `ethernet` - Wired Ethernet connections
- `wifi` - 802.11 wireless connections  
- `cellular` - Mobile/cellular connections
- `vpn` - VPN tunnel interfaces
- `loopback` - Loopback interfaces
- `unknown` - Unidentified interface types

## Complete Monitoring Setup

### 1. Install Windows Agent
```powershell
# Download from releases or build from source
.\windows-agent-collector.exe --agent-id=agent_001 --push.gateway-url=http://pushgateway:9091 install
```

### 2. Install Chrome Extension
Install the [WebRTC Chrome Extension](https://github.com/Brownster/agent-webrtc) and configure it with the same `agent_id=agent_001`.

### 3. Configure Grafana
Create dashboards that query both metric sources using the shared `agent_id` label for unified troubleshooting views.

## Troubleshooting

### Service Installation Issues

#### Access Denied
```powershell
# Run PowerShell as Administrator
# Right-click PowerShell -> "Run as Administrator"
```

#### Service Won't Start
```powershell
# Check Windows Event Log for errors
Get-EventLog -LogName Application -Source "windows_agent_collector" -EntryType Error -Newest 5

# Test configuration manually first
.\windows-agent-collector.exe --config.file=config.yaml --log.level=debug
```

#### Authentication Failures
```powershell
# Test Push Gateway connectivity
curl -u "monitoring_user:secure_password" http://pushgateway:9091/metrics

# Verify credentials in configuration
.\windows-agent-collector.exe --config.file=config.yaml --log.level=debug
```

### Service Management
```powershell
# Check service status and details
Get-Service -Name "windows_agent_collector"
sc qc windows_agent_collector

# View service logs
Get-EventLog -LogName Application -Source "windows_agent_collector" -Newest 10

# Check service executable path
(Get-WmiObject win32_service | Where-Object {$_.name -eq "windows_agent_collector"}).PathName
```

### Configuration Validation
```powershell
# Test configuration file syntax
.\windows-agent-collector.exe --config.file=config.yaml --log.level=debug --help

# Validate Push Gateway connectivity
Test-NetConnection -ComputerName pushgateway.example.com -Port 9091
```

### Common Configuration Issues

| Issue | Solution |
|-------|----------|
| **Push Gateway unreachable** | Check URL, firewall rules, and network connectivity |
| **Authentication failed** | Verify username/password, test with curl |
| **Service crashes on startup** | Check Event Log, validate configuration file |
| **No metrics appearing** | Verify agent_id matches Chrome extension |
| **Permission denied** | Run as Administrator, check file permissions |

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

## Related Projects

- **[WebRTC Chrome Extension](https://github.com/Brownster/agent-webrtc)** - Sister application for capturing WebRTC statistics
- **[Original windows_exporter](https://github.com/prometheus-community/windows_exporter)** - Full-featured Windows metrics exporter

---

*Built specifically for WebRTC voice quality monitoring and troubleshooting scenarios. Use together with the WebRTC Chrome Extension for comprehensive insights.*