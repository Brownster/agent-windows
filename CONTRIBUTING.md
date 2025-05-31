# Contributing to Windows Agent Collector

Thank you for your interest in contributing to the Windows Agent Collector! This document provides guidelines and information for contributors.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Guidelines](#development-guidelines)
- [Testing Guidelines](#testing-guidelines)
- [Documentation Standards](#documentation-standards)
- [Pull Request Process](#pull-request-process)
- [Release Process](#release-process)

## Code of Conduct

This project adheres to a code of conduct that promotes a welcoming and inclusive environment. By participating, you agree to uphold these standards.

## Getting Started

### Prerequisites

- **Go 1.21+**: Required for building and testing
- **Windows 10/11 or Windows Server 2016+**: For testing Windows-specific functionality
- **Git**: For version control
- **GitHub CLI** (optional): For easier GitHub integration

### Development Environment Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/Brownster/agent-windows.git
   cd agent-windows
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Verify setup**
   ```bash
   # Run tests
   go test ./...
   
   # Build for Windows
   GOOS=windows GOARCH=amd64 go build -o windows-agent-collector.exe ./cmd/agent
   ```

4. **Install development tools**
   ```bash
   # Linting
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   
   # Coverage reporting
   go install github.com/axw/gocov/gocov@latest
   
   # Documentation generation
   go install golang.org/x/tools/cmd/godoc@latest
   ```

## Development Guidelines

### 🎯 Core Principles

#### 1. DRY (Don't Repeat Yourself)
- Extract common functionality into shared utilities
- Use interfaces to reduce code duplication
- Create reusable configuration patterns

**Example:**
```go
// ❌ Avoid repetition
func (c *CPUCollector) logError(err error) {
    c.logger.Error("CPU collector error", "error", err)
}

func (m *MemoryCollector) logError(err error) {
    m.logger.Error("Memory collector error", "error", err)
}

// ✅ Use shared utilities
func (c *BaseCollector) logError(name string, err error) {
    c.logger.Error(fmt.Sprintf("%s collector error", name), "error", err)
}
```

#### 2. Clear and Descriptive Code
- Use meaningful variable and function names
- Write self-documenting code
- Add comments for complex logic

**Example:**
```go
// ❌ Unclear naming
func (c *Collector) proc() error {
    d := c.getData()
    if d == nil {
        return errors.New("no data")
    }
    return nil
}

// ✅ Descriptive naming
func (c *NetworkCollector) collectInterfaceMetrics() error {
    interfaceData := c.getNetworkInterfaceData()
    if interfaceData == nil {
        return errors.New("failed to retrieve network interface data")
    }
    return c.processInterfaceData(interfaceData)
}
```

### 🏗️ Code Structure

#### Package Organization
```
internal/
├── collector/          # Metric collectors
│   ├── cpu/           # CPU metrics
│   ├── memory/        # Memory metrics
│   ├── net/           # Network metrics
│   └── pagefile/      # Pagefile metrics
├── config/            # Configuration handling
├── utils/             # Shared utilities
└── types/             # Common types and interfaces

pkg/
└── collector/         # Public collector interfaces

cmd/
└── agent/             # Main application entry point
```

#### Naming Conventions
- **Packages**: lowercase, single word when possible
- **Files**: lowercase with underscores for separation
- **Functions**: camelCase, descriptive verbs
- **Variables**: camelCase, descriptive nouns
- **Constants**: UPPER_CASE with underscores

### 🔧 Code Quality Standards

#### Error Handling
```go
// ✅ Proper error handling with context
func (c *Collector) collectMetrics() error {
    data, err := c.getData()
    if err != nil {
        return fmt.Errorf("failed to collect metrics: %w", err)
    }
    
    if err := c.processData(data); err != nil {
        return fmt.Errorf("failed to process metric data: %w", err)
    }
    
    return nil
}
```

#### Logging Standards
```go
// ✅ Structured logging with context
logger.LogAttrs(ctx, slog.LevelInfo, "Collecting metrics",
    slog.String("collector", "cpu"),
    slog.Duration("interval", 30*time.Second),
    slog.String("agent_id", agentID),
)
```

#### Resource Management
```go
// ✅ Proper resource cleanup
func (c *Collector) collectData() error {
    handle, err := openHandle()
    if err != nil {
        return err
    }
    defer handle.Close()
    
    // Use handle...
    return nil
}
```

## Testing Guidelines

### 🧪 Testing Standards

#### Test Coverage Target: 80%
We aim for 80% test coverage across the codebase. Use the following command to check coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Test Structure
```go
func TestFunctionName(t *testing.T) {
    tests := []struct {
        name     string
        input    inputType
        expected expectedType
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    validInput,
            expected: expectedOutput,
            wantErr:  false,
        },
        {
            name:     "invalid input",
            input:    invalidInput,
            expected: zeroValue,
            wantErr:  true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := functionUnderTest(tt.input)
            
            if tt.wantErr {
                require.Error(t, err)
                return
            }
            
            require.NoError(t, err)
            require.Equal(t, tt.expected, result)
        })
    }
}
```

#### Testing Categories

1. **Unit Tests**: Test individual functions and methods
2. **Integration Tests**: Test component interactions
3. **Performance Tests**: Benchmark critical paths
4. **Windows-specific Tests**: Test Windows API integrations

#### Mock Usage
```go
// Use interfaces for testability
type MetricCollector interface {
    Collect() ([]Metric, error)
}

// Test with mocks
func TestPushMetrics(t *testing.T) {
    mockCollector := &MockCollector{
        metrics: []Metric{{Name: "test", Value: 42}},
    }
    
    err := pushMetrics(mockCollector)
    require.NoError(t, err)
}
```

### 🚀 Performance Testing
```go
func BenchmarkCollectCPUMetrics(b *testing.B) {
    collector := cpu.New()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := collector.Collect()
        if err != nil {
            b.Fatal(err)
        }
    }
    b.ReportAllocs()
}
```

## Documentation Standards

### 📝 Code Documentation

#### Function Documentation
```go
// CollectNetworkMetrics gathers network interface statistics including
// bytes sent/received, interface types, and operational status.
// It returns a slice of network metrics with agent_id labels for
// correlation with WebRTC statistics.
//
// The function performs the following steps:
// 1. Enumerates network interfaces using Windows API
// 2. Determines interface types (ethernet, wifi, vpn, etc.)
// 3. Collects performance counters for each interface
// 4. Formats metrics for Prometheus consumption
//
// Returns an error if network interface enumeration fails or if
// performance counter access is denied.
func (c *NetworkCollector) CollectNetworkMetrics() ([]Metric, error) {
    // Implementation...
}
```

#### Package Documentation
```go
// Package net provides network interface metric collection for the Windows
// Agent Collector. It focuses on gathering metrics essential for WebRTC
// voice quality troubleshooting, including interface types, throughput
// statistics, and operational status.
//
// The package enhances the basic Windows network metrics with intelligent
// interface type detection using both Windows API interface types and
// friendly name heuristics. This enables correlation with WebRTC statistics
// collected by the companion Chrome extension.
//
// Key features:
//   - Network interface type detection (ethernet, wifi, cellular, vpn)
//   - Per-interface throughput metrics
//   - Interface operational status monitoring
//   - WebRTC correlation through consistent agent_id labeling
//
// Example usage:
//   collector := net.New()
//   metrics, err := collector.Collect()
//   if err != nil {
//       log.Fatal(err)
//   }
package net
```

### 📖 User Documentation

#### README Updates
- Keep README.md current with new features
- Include practical examples
- Update configuration documentation
- Maintain troubleshooting guides

#### Configuration Documentation
```yaml
# Example configuration with detailed comments
push_gateway:
  # Prometheus Push Gateway URL (required)
  url: "http://pushgateway.example.com:9091"
  
  # Basic authentication credentials (optional)
  username: "monitoring_user"
  password: "secret_password"
  
  # Push interval - how often to send metrics (default: 30s)
  interval: "30s"

agent:
  # Unique identifier for correlating with WebRTC stats (required)
  # Must match the agent_id configured in the Chrome extension
  id: "agent_001"

collectors:
  # List of enabled metric collectors
  # Available: cpu, memory, net, pagefile
  enabled: ["cpu", "memory", "net", "pagefile"]
```

## Pull Request Process

### 🔄 Before Submitting

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Write tests first** (TDD approach recommended)
   ```bash
   # Write failing tests
   go test ./...
   
   # Implement feature
   # ...
   
   # Ensure tests pass
   go test ./...
   ```

3. **Run quality checks**
   ```bash
   # Linting
   golangci-lint run
   
   # Coverage check
   go test -coverprofile=coverage.out ./...
   go tool cover -func=coverage.out
   
   # Build verification
   GOOS=windows GOARCH=amd64 go build ./cmd/agent
   ```

4. **Update documentation**
   - Add/update function comments
   - Update README if needed
   - Update CHANGELOG.md

### 📝 Pull Request Template

```markdown
## Description
Brief description of changes and motivation.

## Type of Change
- [ ] Bug fix (non-breaking change that fixes an issue)
- [ ] New feature (non-breaking change that adds functionality)
- [ ] Breaking change (fix or feature that changes existing functionality)
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed
- [ ] Performance impact assessed

## Documentation
- [ ] Code comments updated
- [ ] README updated (if needed)
- [ ] Configuration documentation updated (if needed)

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Tests pass locally
- [ ] Coverage remains above 80%
- [ ] No breaking changes to public API
```

### 🔍 Review Process

1. **Automated checks** must pass:
   - Build verification
   - Test suite execution
   - Linting checks
   - Coverage requirements

2. **Manual review** focuses on:
   - Code quality and style
   - Test coverage and quality
   - Documentation completeness
   - Architecture consistency

3. **Approval requirements**:
   - At least one maintainer approval
   - All automated checks passing
   - No unresolved review comments

## Release Process

### 🏷️ Versioning
We follow [Semantic Versioning](https://semver.org/):
- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

### 📦 Release Steps
1. Update version in relevant files
2. Update CHANGELOG.md
3. Create release tag: `git tag v1.x.y`
4. Push tag: `git push origin v1.x.y`
5. GitHub Actions automatically builds and releases

### 📋 Release Checklist
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated
- [ ] Release notes prepared
- [ ] Windows executable tested

## Getting Help

- **Issues**: Report bugs and request features via GitHub Issues
- **Discussions**: Use GitHub Discussions for questions and ideas
- **Documentation**: Check README.md and docs/ directory
- **Code Review**: Tag maintainers for urgent reviews

## Recognition

Contributors will be recognized in:
- CHANGELOG.md for significant contributions
- README.md contributors section
- Release notes for major features

Thank you for contributing to Windows Agent Collector! 🚀