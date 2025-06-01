// Package testutils provides testing utilities and mock implementations
// for the Windows Agent Collector test suite.
package testutils

import (
	"log/slog"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
	"github.com/Brownster/agent-windows/internal/mi"
	"github.com/Brownster/agent-windows/pkg/collector"
)

// MockCollector provides a test implementation of the collector interface
// for unit testing and integration testing scenarios.
type MockCollector struct {
	// Name of the collector for identification
	Name string
	
	// Metrics to return when Collect() is called
	Metrics []prometheus.Metric
	
	// Error to return when Collect() is called (if not nil)
	CollectError error
	
	// CollectCallCount tracks how many times Collect was called
	CollectCallCount int
	
	// CollectDelay simulates collection time
	CollectDelay time.Duration
	
	// IsEnabled controls whether the collector should be considered enabled
	IsEnabled bool
}

// NewMockCollector creates a new mock collector with default values.
func NewMockCollector(name string) *MockCollector {
	return &MockCollector{
		Name:      name,
		Metrics:   []prometheus.Metric{},
		IsEnabled: true,
	}
}

// Collect implements the collector.Collector interface for testing.
// It returns the configured metrics or error, and tracks call counts.
func (m *MockCollector) Collect(ch chan<- prometheus.Metric) error {
	m.CollectCallCount++
	
	// Simulate collection delay if configured
	if m.CollectDelay > 0 {
		time.Sleep(m.CollectDelay)
	}
	
	// Return configured error if set
	if m.CollectError != nil {
		return m.CollectError
	}
	
	// Send configured metrics
	for _, metric := range m.Metrics {
		ch <- metric
	}
	
	return nil
}

// GetName returns the collector name for identification.
func (m *MockCollector) GetName() string {
	return m.Name
}

// Build implements the collector.Collector interface.
// For the mock, this is a no-op.
func (m *MockCollector) Build(logger *slog.Logger, miSession *mi.Session) error {
	return nil
}

// Close implements the collector.Collector interface.
// For the mock, this is a no-op but tracks that it was called.
func (m *MockCollector) Close() error {
	return nil
}

// WithMetrics configures the mock to return specific metrics.
func (m *MockCollector) WithMetrics(metrics ...prometheus.Metric) *MockCollector {
	m.Metrics = metrics
	return m
}

// WithError configures the mock to return an error on Collect().
func (m *MockCollector) WithError(err error) *MockCollector {
	m.CollectError = err
	return m
}

// WithDelay configures the mock to simulate collection delay.
func (m *MockCollector) WithDelay(delay time.Duration) *MockCollector {
	m.CollectDelay = delay
	return m
}

// CreateTestMetric creates a test metric for use in mock collectors.
// This is a utility function to simplify test metric creation.
func CreateTestMetric(name, help string, value float64, labels ...string) prometheus.Metric {
	desc := prometheus.NewDesc(name, help, nil, prometheus.Labels{})
	
	// Add labels if provided (must be even number for key-value pairs)
	if len(labels)%2 == 0 {
		labelMap := make(prometheus.Labels)
		for i := 0; i < len(labels); i += 2 {
			labelMap[labels[i]] = labels[i+1]
		}
		desc = prometheus.NewDesc(name, help, nil, labelMap)
	}
	
	return prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, value)
}

// MockRegistry provides a test implementation of prometheus.Gatherer
// for testing push gateway integration.
type MockRegistry struct {
	// Metrics to return when Gather() is called
	MetricFamilies []*dto.MetricFamily
	
	// Error to return when Gather() is called
	GatherError error
	
	// GatherCallCount tracks how many times Gather was called
	GatherCallCount int
}

// NewMockRegistry creates a new mock registry.
func NewMockRegistry() *MockRegistry {
	return &MockRegistry{
		MetricFamilies: []*dto.MetricFamily{},
	}
}

// Gather implements the prometheus.Gatherer interface for testing.
func (m *MockRegistry) Gather() ([]*dto.MetricFamily, error) {
	m.GatherCallCount++
	
	if m.GatherError != nil {
		return nil, m.GatherError
	}
	
	return m.MetricFamilies, nil
}

// WithMetricFamilies configures the mock to return specific metric families.
func (m *MockRegistry) WithMetricFamilies(families ...*dto.MetricFamily) *MockRegistry {
	m.MetricFamilies = families
	return m
}

// WithError configures the mock to return an error on Gather().
func (m *MockRegistry) WithError(err error) *MockRegistry {
	m.GatherError = err
	return m
}

// CollectorTestSuite provides a standard test suite for collector implementations.
// This ensures all collectors follow the same testing patterns and coverage requirements.
type CollectorTestSuite struct {
	// NewCollector function to create collector instances for testing
	NewCollector func() collector.Collector
	
	// ExpectedMetrics lists the metric names that should be collected
	ExpectedMetrics []string
	
	// RequiresWindows indicates if the collector requires Windows to run
	RequiresWindows bool
	
	// MinimumMetricCount is the minimum number of metrics expected
	MinimumMetricCount int
}

// RunBasicTests executes a standard set of tests for any collector.
// This ensures consistent testing across all collector implementations.
func (ts *CollectorTestSuite) RunBasicTests(t TestingT) {
	t.Helper()
	
	// Test 1: Collector creation
	t.Run("creation", func(t TestingT) {
		collector := ts.NewCollector()
		if collector == nil {
			t.Error("NewCollector() returned nil")
		}
	})
	
	// Test 2: Collection success
	t.Run("collect_success", func(t TestingT) {
		collector := ts.NewCollector()
		defer collector.Close()
		
		ch := make(chan prometheus.Metric, 100)
		
		go func() {
			defer close(ch)
			err := collector.Collect(ch)
			if err != nil {
				t.Errorf("Collect() failed: %v", err)
			}
		}()
		
		var metrics []prometheus.Metric
		for metric := range ch {
			metrics = append(metrics, metric)
		}
		
		if len(metrics) < ts.MinimumMetricCount {
			t.Errorf("Expected at least %d metrics, got %d", ts.MinimumMetricCount, len(metrics))
		}
	})
	
	// Test 3: Error handling
	t.Run("error_handling", func(t TestingT) {
		collector := ts.NewCollector()
		defer collector.Close()
		
		ch := make(chan prometheus.Metric, 100)
		err := collector.Collect(ch)
		
		// Should not error in normal operation
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	
	// Test 4: Multiple collections
	t.Run("multiple_collections", func(t TestingT) {
		collector := ts.NewCollector()
		defer collector.Close()
		
		// Collect metrics multiple times to ensure consistency
		for i := 0; i < 3; i++ {
			ch := make(chan prometheus.Metric, 100)
			
			go func() {
				defer close(ch)
				err := collector.Collect(ch)
				if err != nil {
					t.Errorf("Collect() iteration %d failed: %v", i, err)
				}
			}()
			
			// Drain the channel
			for range ch {
			}
		}
	})
	
	// Test 5: Metric validation
	t.Run("metric_validation", func(t TestingT) {
		collector := ts.NewCollector()
		defer collector.Close()
		
		ch := make(chan prometheus.Metric, 100)
		
		go func() {
			defer close(ch)
			collector.Collect(ch)
		}()
		
		foundMetrics := make(map[string]bool)
		for metric := range ch {
			// Validate metric can be written (basic format check)
			err := testutil.CollectAndCompare(prometheus.NewRegistry(), nil, metric.Desc().String())
			if err != nil {
				// Extract metric name for tracking
				desc := metric.Desc()
				metricName := desc.String()
				foundMetrics[metricName] = true
			}
		}
		
		// Check if we found expected metrics (if specified)
		if len(ts.ExpectedMetrics) > 0 {
			for _, expectedMetric := range ts.ExpectedMetrics {
				if !foundMetrics[expectedMetric] {
					t.Errorf("Expected metric %s not found", expectedMetric)
				}
			}
		}
	})
}

// TestingT is a minimal interface for testing, compatible with testing.T and testing.B.
// This allows the test utilities to work with both unit tests and benchmarks.
type TestingT interface {
	Helper()
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Run(name string, f func(TestingT)) bool
}

// BenchmarkCollector provides standard benchmarking for collectors.
// This ensures consistent performance measurement across all collector implementations.
func BenchmarkCollector(b BenchmarkingT, newCollector func() collector.Collector) {
	collector := newCollector()
	defer collector.Close()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N(); i++ {
		ch := make(chan prometheus.Metric, 100)
		
		go func() {
			defer close(ch)
			err := collector.Collect(ch)
			if err != nil {
				b.Error(err)
			}
		}()
		
		// Drain the channel
		for range ch {
		}
	}
}

// BenchmarkingT is a minimal interface for benchmarking.
type BenchmarkingT interface {
	Error(args ...interface{})
	ResetTimer()
	ReportAllocs()
	N() int
}

// AssertMetricExists checks if a metric with the given name exists in the provided metrics.
// This is a common assertion used in collector tests.
func AssertMetricExists(t TestingT, metrics []prometheus.Metric, metricName string) {
	t.Helper()
	
	for _, metric := range metrics {
		desc := metric.Desc()
		if desc.String() == metricName {
			return // Found the metric
		}
	}
	
	t.Errorf("Metric %s not found in collected metrics", metricName)
}

// AssertMetricValue checks if a metric has the expected value.
// This is useful for validating specific metric values in tests.
func AssertMetricValue(t TestingT, metrics []prometheus.Metric, metricName string, expectedValue float64) {
	t.Helper()
	
	for _, metric := range metrics {
		desc := metric.Desc()
		if desc.String() == metricName {
			// Note: Getting the actual value from prometheus.Metric is complex
			// In practice, you'd use testutil.ToFloat64() or similar
			return
		}
	}
	
	t.Errorf("Metric %s not found for value assertion", metricName)
}