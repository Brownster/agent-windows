# Windows Agent Collector Documentation

This directory contains documentation for the Windows Agent Collector, a lightweight metrics collector designed for WebRTC voice quality troubleshooting.

## Available Collectors

The Windows Agent Collector includes only essential collectors needed for voice quality analysis:

### Core Collectors

- **[CPU Collector](collector.cpu.md)** - CPU utilization, frequency, and per-core metrics
- **[Memory Collector](collector.memory.md)** - Memory usage, availability, and utilization 
- **[Network Collector](collector.net.md)** - Network interface metrics with enhanced type detection
- **[Pagefile Collector](collector.pagefile.md)** - Pagefile/swap usage and availability

## Key Features

### WebRTC Correlation
All metrics include an `agent_id` label that enables correlation with WebRTC statistics collected from Chrome extensions or browser APIs.

### Enhanced Network Detection
The network collector provides detailed interface type detection (ethernet, wifi, cellular, vpn) which is crucial for diagnosing voice quality issues based on connection type.

### Push Gateway Architecture
Unlike traditional pull-based exporters, this agent pushes metrics to a Prometheus Push Gateway, making it ideal for environments where direct scraping isn't feasible.

## Configuration

See the main [configuration example](../config-example.yaml) for detailed setup instructions.

## Differences from windows_exporter

This agent is purpose-built for specific use cases:

- **Focused Scope**: Only 4 collectors vs 50+ in windows_exporter
- **Push vs Pull**: Pushes metrics instead of exposing HTTP endpoints
- **WebRTC Integration**: Built-in correlation capabilities
- **Lightweight**: Significantly reduced resource usage
- **Network Awareness**: Enhanced network interface type detection

For general Windows monitoring, use the full [windows_exporter](https://github.com/prometheus-community/windows_exporter). For WebRTC voice quality troubleshooting, this agent provides focused, efficient monitoring.