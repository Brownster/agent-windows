# Development Guide

This guide provides comprehensive information for developers working on the Windows Agent Collector.

## 📋 Table of Contents

- [Architecture Overview](#architecture-overview)
- [Development Environment](#development-environment)
- [Code Structure](#code-structure)
- [Testing Strategy](#testing-strategy)
- [Performance Guidelines](#performance-guidelines)
- [Debugging and Troubleshooting](#debugging-and-troubleshooting)
- [Release Process](#release-process)

## Architecture Overview

### High-Level Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│ Windows Agent   │    │ Push Gateway     │    │ Prometheus      │
│ Collector       │───▶│                  │───▶│                 │
│                 │    │                  │    │                 │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                                               │
         │                                               ▼
         │              ┌──────────────────┐    ┌─────────────────┐
         │              │ WebRTC Chrome    │    │ Grafana         │
         └──────────────│ Extension        │    │ Dashboard       │
       agent_id         │                  │    │                 │
       correlation      └──────────────────┘    └─────────────────┘
```

### Core Components

#### 1. Collector Framework
- **Location**: `pkg/collector/`
- **Purpose**: Common interfaces and utilities for metric collection
- **Key Files**:
  - `collect.go`: Core collection interfaces
  - `map.go`: Collector registry and management
  - `types.go`: Common data types

#### 2. Individual Collectors
- **Location**: `internal/collector/`
- **Purpose**: Specific metric collection implementations
- **Structure**:
  ```
  internal/collector/
  ├── cpu/        # CPU utilization and frequency
  ├── memory/     # Memory usage and availability  
  ├── net/        # Network interface metrics
  └── pagefile/   # Virtual memory/swap metrics
  ```

#### 3. Configuration System
- **Location**: `internal/config/`
- **Purpose**: Configuration parsing and validation
- **Features**:
  - YAML configuration support
  - Environment variable override
  - Command-line flag integration
  - Configuration validation

#### 4. Push Gateway Integration
- **Location**: `cmd/agent/main.go`
- **Purpose**: Metric submission to Prometheus Push Gateway
- **Features**:
  - Authentication support
  - Retry logic with exponential backoff
  - Agent ID labeling for correlation
  - Error handling and logging

## Development Environment

### Prerequisites

```bash
# Required Software
- Go 1.21+ 
- Git
- Windows 10/11 or Windows Server 2016+
- VS Code or GoLand (recommended)

# Development Tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install github.com/axw/gocov/gocov@latest
go install golang.org/x/tools/cmd/godoc@latest
```

### IDE Configuration

#### VS Code Settings (.vscode/settings.json)
```json
{
    "go.testFlags": ["-v", "-race"],
    "go.coverOnSave": true,
    "go.coverageDecorator": {
        "type": "gutter",
        "coveredHighlightColor": "rgba(64,128,64,0.5)",
        "uncoveredHighlightColor": "rgba(128,64,64,0.5)"
    },
    "go.lintOnSave": "package",
    "go.lintTool": "golangci-lint",
    "go.buildOnSave": "workspace"
}
```

#### Recommended Extensions
- Go (Google)
- Go Test Explorer
- GitLens
- Error Lens
- Test Explorer UI

### Environment Setup

```bash
# Clone and setup
git clone https://github.com/Brownster/agent-windows.git
cd agent-windows

# Install dependencies
go mod download

# Verify setup
go test ./...
go build ./cmd/agent
```

## Code Structure

### Package Hierarchy

```
agent-windows/
├── cmd/
│   └── agent/              # Main application
│       ├── main.go         # Entry point and CLI
│       ├── main_test.go    # Integration tests
│       └── 0_service.go    # Windows service integration
├── internal/               # Private packages
│   ├── collector/          # Metric collectors
│   │   ├── cpu/           # CPU metrics
│   │   ├── memory/        # Memory metrics
│   │   ├── net/           # Network metrics
│   │   └── pagefile/      # Pagefile metrics
│   ├── config/            # Configuration handling
│   ├── log/               # Logging utilities
│   ├── types/             # Common types
│   └── utils/             # Shared utilities
├── pkg/                   # Public packages
│   └── collector/         # Public collector interfaces
├── docs/                  # Documentation
│   ├── adr/              # Architecture Decision Records
│   └── examples/         # Configuration examples
└── .github/              # GitHub Actions workflows
    └── workflows/
```

### Naming Conventions

#### Files and Packages
- **Packages**: Single word, lowercase (`cpu`, `memory`, `net`)
- **Files**: Lowercase with underscores (`cpu_info.go`, `net_test.go`)
- **Test files**: `*_test.go` suffix
- **Benchmark files**: Include `Benchmark` functions

#### Functions and Variables
```go
// ✅ Good naming
func CollectCPUMetrics() ([]Metric, error)
func (c *CPUCollector) getCPUUtilization() float64
var defaultConfigPath = "config.yaml"

// ❌ Avoid abbreviations
func CollectCPU() ([]Metric, error)     // Too abbreviated
func GetUtil() float64                   // Unclear purpose
var cfgPath = "config.yaml"            // Abbreviated
```

#### Constants and Types
```go
// Constants: UPPER_CASE with underscores
const (
    DEFAULT_PUSH_INTERVAL = 30 * time.Second
    MAX_RETRY_ATTEMPTS   = 3
)

// Types: PascalCase
type MetricCollector interface {
    Collect() ([]Metric, error)
}

type PushConfig struct {
    URL      string
    Username string  
    Password string
}
```

## Testing Strategy

### Test Coverage Requirements

We maintain **80% test coverage** across the codebase. Coverage is tracked per package and enforced in CI/CD.

```bash
# Check coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
go tool cover -func=coverage.out | grep total
```

### Testing Patterns

#### Table-Driven Tests
```go
func TestExpandEnabledCollectors(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected []string
        wantErr  bool
    }{
        {
            name:     "defaults expansion",
            input:    "[defaults]",
            expected: []string{"cpu", "memory", "net", "pagefile"},
            wantErr:  false,
        },
        {
            name:     "empty input",
            input:    "",
            expected: []string{},
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := expandEnabledCollectors(tt.input)
            require.Equal(t, tt.expected, result)
        })
    }
}
```

#### Mock Interfaces
```go
// Define testable interfaces
type MetricGatherer interface {
    Gather() ([]*dto.MetricFamily, error)
}

// Use dependency injection for testing
func NewPushGateway(gatherer MetricGatherer) *PushGateway {
    return &PushGateway{gatherer: gatherer}
}

// Test with mocks
func TestPushMetrics(t *testing.T) {
    mockGatherer := &MockGatherer{
        metrics: []*dto.MetricFamily{...},
    }
    
    gateway := NewPushGateway(mockGatherer)
    err := gateway.Push()
    require.NoError(t, err)
}
```

#### Windows API Testing
```go
// Test Windows-specific functionality
//go:build windows

func TestCPUCollector_Windows(t *testing.T) {
    if runtime.GOOS != "windows" {
        t.Skip("Windows-only test")
    }
    
    collector := cpu.New()
    metrics, err := collector.Collect()
    
    require.NoError(t, err)
    require.NotEmpty(t, metrics)
    
    // Verify Windows-specific metrics
    cpuTimeMetric := findMetric(metrics, "windows_cpu_time_total")
    require.NotNil(t, cpuTimeMetric)
}
```

### Performance Testing

#### Benchmark Tests
```go
func BenchmarkCPUCollector_Collect(b *testing.B) {
    collector := cpu.New()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        _, err := collector.Collect()
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

#### Memory Profiling
```bash
# Profile memory usage
go test -memprofile=mem.prof -bench=BenchmarkCPUCollector
go tool pprof mem.prof

# Profile CPU usage  
go test -cpuprofile=cpu.prof -bench=BenchmarkCPUCollector
go tool pprof cpu.prof
```

## Performance Guidelines

### Resource Usage Targets

| Metric | Target | Measurement |
|--------|--------|-------------|
| Memory Usage | <50MB | Baseline runtime memory |
| CPU Overhead | <2% | Average CPU usage during collection |
| Collection Time | <500ms | Time to collect all metrics |
| Startup Time | <2s | Time from start to first metric collection |

### Optimization Techniques

#### Efficient Data Collection
```go
// ✅ Batch API calls
func (c *NetworkCollector) collectAll() error {
    adapters, err := c.getAllAdapters()  // Single API call
    if err != nil {
        return err
    }
    
    for _, adapter := range adapters {
        c.processAdapter(adapter)
    }
    return nil
}

// ❌ Avoid repeated API calls
func (c *NetworkCollector) collectSeparately() error {
    for _, name := range c.getAdapterNames() {
        adapter, err := c.getAdapter(name)  // Multiple API calls
        if err != nil {
            continue
        }
        c.processAdapter(adapter)
    }
    return nil
}
```

#### Memory Management
```go
// ✅ Reuse slices and maps
type CPUCollector struct {
    metrics []Metric  // Reused across collections
    buffer  []byte    // Reused buffer for API calls
}

func (c *CPUCollector) Collect() ([]Metric, error) {
    c.metrics = c.metrics[:0]  // Reset slice, keep capacity
    
    // Collect into existing slice
    return c.metrics, nil
}
```

#### Caching Strategies
```go
// Cache static information
type NetworkCollector struct {
    interfaceTypes map[string]string  // Cache interface types
    lastUpdate     time.Time
}

func (c *NetworkCollector) getInterfaceType(name string) string {
    if c.interfaceTypes == nil || time.Since(c.lastUpdate) > 5*time.Minute {
        c.refreshInterfaceTypes()
    }
    return c.interfaceTypes[name]
}
```

## Debugging and Troubleshooting

### Logging Configuration

#### Log Levels
```go
// Use structured logging with appropriate levels
logger.LogAttrs(ctx, slog.LevelDebug, "Collecting metrics",
    slog.String("collector", "cpu"),
    slog.Duration("interval", 30*time.Second),
)

logger.LogAttrs(ctx, slog.LevelError, "Failed to collect metrics",
    slog.Any("error", err),
    slog.String("collector", "memory"),
)
```

#### Debug Mode
```bash
# Enable debug logging
windows-agent-collector.exe --log.level=debug

# Trace specific collectors
windows-agent-collector.exe --log.level=debug --collectors.enabled=cpu,memory
```

### Common Issues and Solutions

#### Permission Issues
```
Error: Access denied when reading performance counters
```
**Solution**: Run as Administrator or ensure the service account has Performance Monitor Users rights.

#### Network Connectivity
```
Error: Failed to push metrics: connection refused
```
**Solution**: 
1. Verify Push Gateway URL and port
2. Check firewall rules
3. Validate authentication credentials

#### Memory Leaks
```
Error: Memory usage continuously increasing
```
**Debugging**:
```bash
# Profile memory usage
go tool pprof http://localhost:6060/debug/pprof/heap

# Monitor with built-in metrics
curl http://localhost:8080/debug/metrics
```

### Development Tools

#### Live Reloading
```bash
# Install air for live reloading
go install github.com/cosmtrek/air@latest

# Create .air.toml configuration
# Run with auto-reload
air
```

#### Debugging in VS Code
```json
// .vscode/launch.json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Debug Agent",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "./cmd/agent",
            "args": [
                "--agent-id=debug_agent",
                "--push.gateway-url=http://localhost:9091",
                "--log.level=debug"
            ],
            "env": {}
        }
    ]
}
```

## Release Process

### Version Management

We follow [Semantic Versioning](https://semver.org/):
- **MAJOR.MINOR.PATCH** (e.g., 1.2.3)
- **Pre-release**: 1.2.3-alpha.1, 1.2.3-beta.1, 1.2.3-rc.1

### Release Checklist

#### Pre-Release (Development)
- [ ] All tests passing (80% coverage maintained)
- [ ] Performance benchmarks within acceptable range
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated in relevant files

#### Release Creation
```bash
# Create release branch
git checkout -b release/v1.2.0

# Update version references
# Update CHANGELOG.md
# Commit changes

# Create and push tag
git tag v1.2.0
git push origin v1.2.0
```

#### Post-Release
- [ ] GitHub release created automatically
- [ ] Windows executable tested
- [ ] Documentation published
- [ ] Community notification (if significant release)

### Automated Release Pipeline

The GitHub Actions workflow automatically:
1. Builds Windows executable
2. Runs full test suite
3. Generates coverage reports
4. Creates GitHub release with assets
5. Updates documentation

### Hotfix Process

For critical issues requiring immediate release:

```bash
# Create hotfix branch from latest release tag
git checkout v1.2.0
git checkout -b hotfix/v1.2.1

# Make minimal fix
# Update version to 1.2.1
# Commit and tag

git tag v1.2.1
git push origin v1.2.1

# Merge back to master
git checkout master
git merge hotfix/v1.2.1
```

---

This development guide provides the foundation for high-quality contributions to the Windows Agent Collector. For additional questions, please refer to the [Contributing Guide](../CONTRIBUTING.md) or open a GitHub Discussion.