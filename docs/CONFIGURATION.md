# Configuration Guide

This guide provides detailed information about configuring the Windows Agent Collector.

## Configuration Methods

The Windows Agent Collector supports three configuration methods (in order of precedence):

1. **Command Line Flags** (highest priority)
2. **Environment Variables** 
3. **Configuration File** (lowest priority)

## Configuration File Structure

The configuration file uses YAML format and must match the command line flag structure:

### Basic Configuration

```yaml
# Push Gateway Configuration
push:
  gateway-url: "http://pushgateway.example.com:9091"
  username: "monitoring_user"
  password: "secret_password"
  interval: "30s"
  job-name: "windows_agent"

# Agent Configuration
agent-id: "agent_001"

# Collectors Configuration
collectors:
  enabled: "cpu,memory,net,pagefile"

# Logging Configuration
log:
  level: "info"
  format: "text"

# Process Configuration
process:
  priority: "normal"
  memory-limit: "0"
```

### Advanced Configuration

```yaml
# Complete configuration example
push:
  gateway-url: "https://pushgateway.prod.company.com:9091"
  username: "${PUSH_USERNAME}"
  password: "${PUSH_PASSWORD}"
  interval: "30s"
  job-name: "windows_agent_prod"

agent-id: "${AGENT_ID}"

collectors:
  enabled: "cpu,memory,net,pagefile"

log:
  level: "info"
  format: "json"

process:
  priority: "normal"
  memory-limit: "0"
```

## Environment Variables

You can use environment variables in the configuration file or set them directly:

| Environment Variable | Configuration Key | Description |
|---------------------|------------------|-------------|
| `AGENT_ID` | `agent-id` | Agent identifier |
| `PUSH_GATEWAY_URL` | `push.gateway-url` | Push Gateway URL |
| `PUSH_USERNAME` | `push.username` | Basic auth username |
| `PUSH_PASSWORD` | `push.password` | Basic auth password |
| `PUSH_INTERVAL` | `push.interval` | Push interval |
| `PUSH_JOB_NAME` | `push.job-name` | Prometheus job name |
| `COLLECTORS_ENABLED` | `collectors.enabled` | Enabled collectors |
| `LOG_LEVEL` | `log.level` | Log level |
| `LOG_FORMAT` | `log.format` | Log format |

### Setting Environment Variables

```powershell
# Windows PowerShell
$env:AGENT_ID = "agent_001"
$env:PUSH_GATEWAY_URL = "http://pushgateway:9091"
$env:PUSH_USERNAME = "monitoring_user"
$env:PUSH_PASSWORD = "secure_password"

# Windows Command Prompt
set AGENT_ID=agent_001
set PUSH_GATEWAY_URL=http://pushgateway:9091
set PUSH_USERNAME=monitoring_user
set PUSH_PASSWORD=secure_password
```

## Configuration Validation

### Validate Configuration File

```powershell
# Test configuration file syntax
.\windows-agent-collector.exe --config.file=config.yaml --help

# Dry run with debug logging
.\windows-agent-collector.exe --config.file=config.yaml --log.level=debug
```

### Common Configuration Errors

#### 1. Incorrect YAML Structure

❌ **Wrong:**
```yaml
push_gateway:  # Underscore not supported
  url: "http://pushgateway:9091"
```

✅ **Correct:**
```yaml
push:  # Use hyphen structure matching CLI flags
  gateway-url: "http://pushgateway:9091"
```

#### 2. Array vs String for Collectors

❌ **Wrong:**
```yaml
collectors:
  enabled: ["cpu", "memory", "net", "pagefile"]  # Array not supported
```

✅ **Correct:**
```yaml
collectors:
  enabled: "cpu,memory,net,pagefile"  # Comma-separated string
```

#### 3. Missing Required Fields

❌ **Wrong:**
```yaml
push:
  username: "user"
  password: "pass"
# Missing gateway-url and agent-id
```

✅ **Correct:**
```yaml
push:
  gateway-url: "http://pushgateway:9091"  # Required
  username: "user"
  password: "pass"

agent-id: "agent_001"  # Required
```

## Security Best Practices

### 1. File Permissions

Set restrictive permissions on configuration files:

```powershell
# Restrict config file access
icacls config.yaml /inheritance:d
icacls config.yaml /grant:r "Administrators:(R)"
icacls config.yaml /remove "Users"
```

### 2. Environment Variables for Secrets

Use environment variables for sensitive information:

```yaml
push:
  gateway-url: "${PUSH_GATEWAY_URL}"
  username: "${PUSH_USERNAME}"
  password: "${PUSH_PASSWORD}"
  
agent-id: "${AGENT_ID}"
```

### 3. HTTPS for Production

Always use HTTPS in production:

```yaml
push:
  gateway-url: "https://pushgateway.prod.company.com:9091"  # HTTPS
```

## Configuration Examples

### Development Environment

```yaml
# dev-config.yaml
push:
  gateway-url: "http://localhost:9091"
  interval: "10s"  # Faster updates for development
  job-name: "windows_agent_dev"

agent-id: "dev_agent_001"

collectors:
  enabled: "cpu,memory,net,pagefile"

log:
  level: "debug"  # Verbose logging for development
  format: "text"
```

### Production Environment

```yaml
# prod-config.yaml
push:
  gateway-url: "https://pushgateway.prod.company.com:9091"
  username: "${PUSH_USERNAME}"
  password: "${PUSH_PASSWORD}"
  interval: "30s"
  job-name: "windows_agent_prod"

agent-id: "${AGENT_ID}"

collectors:
  enabled: "cpu,memory,net,pagefile"

log:
  level: "info"
  format: "json"  # Structured logging for production

process:
  priority: "normal"
  memory-limit: "0"
```

### High-Performance Environment

```yaml
# high-perf-config.yaml
push:
  gateway-url: "https://pushgateway.company.com:9091"
  username: "${PUSH_USERNAME}"
  password: "${PUSH_PASSWORD}"
  interval: "60s"  # Less frequent updates
  job-name: "windows_agent_highperf"

agent-id: "${AGENT_ID}"

collectors:
  enabled: "cpu,memory,net,pagefile"

log:
  level: "warn"  # Minimal logging
  format: "json"

process:
  priority: "belownormal"  # Lower priority
  memory-limit: "52428800"  # 50MB limit
```

## Troubleshooting Configuration

### Configuration File Not Found

```
Error: failed to load configuration file: failed to open configuration file: open config.yaml: The system cannot find the file specified.
```

**Solution:**
- Verify the file path is correct
- Use absolute path if needed
- Check file exists: `Test-Path config.yaml`

### YAML Syntax Errors

```
Error: configuration file validation error: yaml: line 7: could not find expected ':'
```

**Solution:**
- Validate YAML syntax online or with tools
- Check indentation (use spaces, not tabs)
- Ensure proper key-value structure

### Unknown Configuration Keys

```
Error: configuration file validation error: yaml: unmarshal errors: line 1: field pushgateway not found in type config.configFile
```

**Solution:**
- Use correct key names matching CLI flags
- Check [Configuration Reference](#configuration-reference) for valid keys
- Remove unsupported configuration options

### Environment Variable Substitution

Environment variables are substituted at runtime:

```yaml
# This will be replaced with actual environment variable value
agent-id: "${AGENT_ID}"
```

If environment variable is not set, the literal string (including `${}`) will be used.

## Configuration Reference

### Complete Flag-to-YAML Mapping

| CLI Flag | YAML Path | Type | Default | Description |
|----------|-----------|------|---------|-------------|
| `--agent-id` | `agent-id` | string | *required* | Agent identifier |
| `--push.gateway-url` | `push.gateway-url` | string | *required* | Push Gateway URL |
| `--push.username` | `push.username` | string | "" | Basic auth username |
| `--push.password` | `push.password` | string | "" | Basic auth password |
| `--push.interval` | `push.interval` | duration | "30s" | Push interval |
| `--push.job-name` | `push.job-name` | string | "windows_agent" | Job name |
| `--collectors.enabled` | `collectors.enabled` | string | "cpu,memory,net,pagefile" | Enabled collectors |
| `--log.level` | `log.level` | string | "info" | Log level |
| `--log.format` | `log.format` | string | "text" | Log format |
| `--process.priority` | `process.priority` | string | "normal" | Process priority |
| `--process.memory-limit` | `process.memory-limit` | string | "0" | Memory limit in bytes |

### Valid Values

#### Log Levels
- `debug` - Very verbose output
- `info` - General information
- `warn` - Warning messages only
- `error` - Error messages only

#### Log Formats
- `text` - Human-readable text format
- `json` - Structured JSON format

#### Process Priorities
- `realtime` - Real-time priority (use with caution)
- `high` - High priority
- `abovenormal` - Above normal priority
- `normal` - Normal priority (default)
- `belownormal` - Below normal priority
- `low` - Low priority

#### Collectors
- `cpu` - CPU utilization metrics
- `memory` - Memory usage metrics
- `net` - Network interface metrics
- `pagefile` - Virtual memory metrics

---

For more information, see the main [README](../README.md) or [Development Guide](DEVELOPMENT.md).