# ADR-001: Push Gateway Architecture Over HTTP Server

## Status
**Accepted** - December 2024

## Context
The original windows_exporter uses an HTTP server model where Prometheus scrapes metrics from endpoints (pull model). For the Windows Agent Collector designed for WebRTC troubleshooting, we needed to decide between:

1. **Pull Model (HTTP Server)**: Traditional Prometheus approach
2. **Push Model (Push Gateway)**: Active metric submission

## Decision
We decided to implement a **Push Gateway architecture** where the agent actively pushes metrics to a Prometheus Push Gateway.

## Rationale

### Advantages of Push Model
1. **Simplified Deployment**: No need to configure firewall rules or network access to agent machines
2. **Better for Ephemeral Workloads**: Ideal for troubleshooting scenarios where agents may come and go
3. **Centralized Collection**: All metrics flow to a central gateway, simplifying monitoring infrastructure
4. **NAT/Firewall Friendly**: Outbound connections are typically easier than inbound
5. **WebRTC Use Case Alignment**: Matches the temporary, diagnostic nature of WebRTC troubleshooting

### Technical Benefits
- **Reduced Network Complexity**: No need for service discovery or endpoint management
- **Authentication Support**: Push Gateway provides centralized authentication
- **Buffering Capability**: Gateway can buffer metrics during network issues
- **Horizontal Scaling**: Multiple agents can push to the same gateway

### Trade-offs Accepted
- **Additional Component**: Requires Push Gateway deployment
- **Different Operational Model**: DevOps teams need to adapt from pull to push
- **Potential Data Loss**: If push fails and no local buffering
- **Less Real-time**: Depends on push frequency rather than scrape interval

## Implementation Details

### Push Configuration
```go
type PushConfig struct {
    URL      string        // Push Gateway URL
    Username string        // Basic auth username
    Password string        // Basic auth password
    Interval time.Duration // Push frequency
    AgentID  string        // Unique agent identifier
    JobName  string        // Prometheus job name
}
```

### Error Handling Strategy
- Immediate retry on transient failures
- Exponential backoff for persistent failures
- Comprehensive error logging with context
- Graceful degradation when push gateway unavailable

### Security Considerations
- Support for HTTPS endpoints
- Basic authentication support
- Future TLS client certificate support
- No sensitive data in metric labels

## Consequences

### Positive
- **Simplified Agent Deployment**: No firewall configuration needed
- **Better User Experience**: Agents work out-of-the-box in most network environments
- **Aligned with Use Case**: Perfect for temporary WebRTC troubleshooting scenarios
- **Centralized Management**: All metrics flow through controlled gateway

### Negative
- **Infrastructure Dependency**: Requires Push Gateway deployment and maintenance
- **Operational Change**: Teams familiar with pull model need to adapt
- **Potential Bottleneck**: Gateway becomes single point of failure
- **Different Monitoring Patterns**: Traditional Prometheus alerting may need adjustment

### Neutral
- **Learning Curve**: Teams need to understand push gateway operations
- **Monitoring Approach**: Different from traditional Prometheus setups but not inherently better/worse

## Alternatives Considered

### 1. Traditional HTTP Server (Rejected)
**Pros**: 
- Familiar to Prometheus users
- No additional infrastructure components
- Direct scraping model

**Cons**: 
- Complex firewall/network configuration
- Poor fit for WebRTC troubleshooting use case
- Requires service discovery for dynamic agents

### 2. Hybrid Approach (Deferred)
**Description**: Support both push and pull models
**Decision**: Deferred to future versions to avoid complexity

### 3. Direct Prometheus Remote Write (Rejected)
**Pros**: 
- No intermediate gateway needed
- Direct to Prometheus

**Cons**: 
- More complex authentication
- Requires Prometheus remote write configuration
- Less flexible for different backends

## Compliance and Standards
- Follows Prometheus Push Gateway conventions
- Compatible with standard Prometheus deployments
- Maintains metric naming conventions
- Supports standard Prometheus data types

## Future Considerations
- **Local Buffering**: Implement local metric buffering for offline scenarios
- **Multiple Gateways**: Support for multiple gateway targets for redundancy
- **Gateway Health Monitoring**: Monitor gateway availability and switch targets
- **Compression**: Add metric compression for bandwidth optimization

## References
- [Prometheus Push Gateway Documentation](https://prometheus.io/docs/instrumenting/pushing/)
- [When to Use Push Gateway](https://prometheus.io/docs/practices/pushing/)
- [WebRTC Monitoring Best Practices](https://webrtc.org/getting-started/testing)

---
*This ADR documents the architectural decision for Windows Agent Collector v1.0.0 and serves as reference for future development decisions.*