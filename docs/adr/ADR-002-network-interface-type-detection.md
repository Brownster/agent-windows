# ADR-002: Network Interface Type Detection Strategy

## Status
**Accepted** - December 2024

## Context
For WebRTC troubleshooting, knowing the network interface type (ethernet, wifi, cellular, vpn) is crucial for correlating connection quality with network conditions. Windows provides interface type information through APIs, but the mapping is not always accurate or granular enough for WebRTC analysis.

## Decision
We decided to implement a **hybrid detection strategy** that combines Windows API interface types with intelligent friendly name pattern matching.

## Rationale

### WebRTC Correlation Requirements
WebRTC statistics often reference connection types, and correlating these with actual network interfaces helps identify:
- WiFi signal strength impacts on call quality
- Ethernet vs WiFi performance differences  
- VPN overhead effects on latency
- Cellular connection limitations

### Windows API Limitations
The Windows `IF_TYPE_*` constants provide basic categorization but:
- Many adapters report generic types (IF_TYPE_ETHERNET_CSMACD for everything)
- Virtual adapters often misclassified
- No distinction between physical and virtual interfaces
- Limited granularity for modern adapter types

### Pattern Matching Benefits
Friendly name analysis provides:
- More accurate virtual adapter detection
- Better WiFi adapter identification
- VPN and virtualization technology recognition
- Manufacturer-specific adapter classification

## Implementation Strategy

### Two-Stage Detection Process

#### Stage 1: Windows API Type Mapping
```go
interfaceType = map[uint32]string{
    windows.IF_TYPE_ETHERNET_CSMACD:   "ethernet",
    windows.IF_TYPE_IEEE80211:         "wifi", 
    windows.IF_TYPE_PPP:               "cellular",
    windows.IF_TYPE_TUNNEL:            "vpn",
    windows.IF_TYPE_SOFTWARE_LOOPBACK: "loopback",
}
```

#### Stage 2: Friendly Name Heuristics
Applied when API type is unknown or generic:

```go
// VPN patterns checked first to avoid misclassification
vpnPatterns := []string{"vpn", "tap", "tun", "virtual", "vmware", "virtualbox", "hyper-v"}

// WiFi patterns for wireless adapters
wifiPatterns := []string{"wi-fi", "wifi", "wireless", "802.11", "wlan"}

// Ethernet patterns for wired connections  
ethernetPatterns := []string{"ethernet", "gigabit", "fast ethernet", "realtek pcie"}

// Cellular patterns for mobile connections
cellularPatterns := []string{"cellular", "mobile", "3g", "4g", "lte", "5g", "modem"}
```

### Pattern Priority Order
1. **VPN/Virtual** (highest priority) - Prevents virtual adapters from being misclassified as ethernet
2. **WiFi** - Specific wireless technology indicators
3. **Ethernet** - Wired connection indicators  
4. **Cellular** - Mobile connection indicators
5. **Unknown** (default) - Fallback for unrecognized types

## Technical Design

### Function Signature
```go
func GetInterfaceType(ifType uint32, friendlyName string) string
```

### Error Handling
- Always returns a valid string (never empty)
- Defaults to "unknown" for unclassifiable interfaces
- Case-insensitive pattern matching
- Graceful handling of empty/null friendly names

### Performance Considerations
- Pattern matching uses simple `strings.Contains()` for performance
- Patterns ordered by frequency for early matching
- No regular expressions to avoid performance overhead
- Results could be cached per interface for repeated calls

## Use Cases Supported

### WebRTC Correlation Scenarios
1. **WiFi Quality Analysis**: Identify WiFi adapters for signal strength correlation
2. **VPN Impact Assessment**: Detect VPN interfaces to understand latency impact
3. **Ethernet vs WiFi Comparison**: Compare performance across interface types
4. **Mobile Connection Monitoring**: Track cellular connection quality
5. **Virtual Environment Detection**: Identify virtualized network adapters

### Grafana Dashboard Queries
```promql
# WiFi interface metrics
windows_net_bytes_received_total{interface_type="wifi"}

# Compare ethernet vs wifi performance
rate(windows_net_bytes_received_total{interface_type=~"ethernet|wifi"}[5m])

# VPN overhead analysis
windows_net_nic_info{interface_type="vpn"}
```

## Validation and Testing

### Test Coverage Strategy
- Unit tests for known adapter name patterns
- Edge cases (empty names, unknown types)
- Real-world adapter name validation
- Performance benchmarks for pattern matching

### Known Adapter Types Tested
- Intel WiFi adapters: "Intel(R) Wi-Fi 6 AX200 160MHz" → wifi
- Realtek Ethernet: "Realtek PCIe GbE Family Controller" → ethernet  
- VMware Virtual: "VMware Virtual Ethernet Adapter" → vpn
- VPN Clients: "TAP-Windows Adapter V9" → vpn
- Mobile Broadband: "Mobile Broadband Adapter" → cellular

## Consequences

### Positive
- **Accurate Classification**: Better than API-only approach
- **WebRTC Alignment**: Matches WebRTC connection type concepts
- **Extensible**: Easy to add new patterns for emerging technologies
- **Performance**: Lightweight pattern matching with good performance

### Negative  
- **Maintenance Overhead**: Patterns need updates for new adapter types
- **Localization Issues**: Friendly names may vary by Windows language
- **False Positives**: Pattern matching may occasionally misclassify

### Neutral
- **Heuristic Nature**: Results are best-effort, not guaranteed accurate
- **Cultural Dependency**: Adapter naming conventions vary by region/manufacturer

## Alternatives Considered

### 1. API-Only Detection (Rejected)
**Reason**: Insufficient granularity for WebRTC use cases
**Issues**: Too many false classifications, poor virtual adapter detection

### 2. WMI Deep Inspection (Rejected)  
**Reason**: Performance overhead and complexity
**Issues**: Requires extensive WMI queries, slower collection

### 3. Registry Analysis (Rejected)
**Reason**: Fragile and version-dependent
**Issues**: Registry structure changes between Windows versions

### 4. Hardware Detection (Rejected)
**Reason**: Requires elevated privileges and hardware knowledge
**Issues**: Complex implementation, privilege escalation needs

## Future Enhancements

### Short Term (v1.2.0)
- Add more adapter name patterns based on user feedback
- Implement adapter name pattern configuration
- Add confidence scoring for classification results

### Medium Term (v1.3.0)
- IPv6 interface support
- Network speed detection integration
- Signal strength metrics for WiFi interfaces

### Long Term (v1.4.0+)
- Machine learning-based classification
- Automatic pattern learning from environment
- Integration with Windows network location awareness

## Monitoring and Metrics

### Classification Accuracy Tracking
```go
// Track classification confidence
windows_agent_interface_classification_confidence{method="api|pattern|unknown"}

// Monitor unknown interface types for pattern expansion
windows_agent_unknown_interface_types{friendly_name="adapter_name"}
```

### Pattern Effectiveness
- Track which patterns match most frequently
- Identify new adapter types for pattern expansion
- Monitor false positive rates through user feedback

## References
- [Windows Network Interface Types](https://docs.microsoft.com/en-us/windows/win32/api/iptypes/ns-iptypes-ip_adapter_addresses_lh)
- [WebRTC Connection Types](https://developer.mozilla.org/en-US/docs/Web/API/RTCIceCandidate/type)
- [Prometheus Metric Labels Best Practices](https://prometheus.io/docs/practices/naming/)

---
*This ADR documents the network interface type detection strategy for Windows Agent Collector and guides future enhancements in this area.*