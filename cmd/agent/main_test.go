// SPDX-License-Identifier: Apache-2.0
//
// Copyright The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build windows

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
)

func TestExpandEnabledCollectors(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "defaults",
			input:    "[defaults]",
			expected: []string{"cpu", "memory", "net", "pagefile"},
		},
		{
			name:     "explicit collectors",
			input:    "cpu,memory",
			expected: []string{"cpu", "memory"},
		},
		{
			name:     "unsupported collectors filtered",
			input:    "cpu,memory,iis,exchange",
			expected: []string{"cpu", "memory"},
		},
		{
			name:     "empty input",
			input:    "",
			expected: []string{},
		},
		{
			name:     "single collector",
			input:    "cpu",
			expected: []string{"cpu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandEnabledCollectors(tt.input)
			require.Equal(t, tt.expected, result)
		})
	}
}


func TestPushMetrics(t *testing.T) {
	// Create a mock push gateway server
	var receivedRequests []string
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()

		// Capture the request details
		receivedRequests = append(receivedRequests, fmt.Sprintf("%s %s", r.Method, r.URL.Path))

		// Check for basic auth
		username, password, hasAuth := r.BasicAuth()
		if hasAuth {
			receivedRequests = append(receivedRequests, fmt.Sprintf("auth: %s:%s", username, password))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Create a test registry with a simple metric
	registry := prometheus.NewRegistry()
	testGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_metric",
		Help: "A test metric",
	})
	testGauge.Set(42)
	registry.MustRegister(testGauge)

	tests := []struct {
		name   string
		config PushConfig
	}{
		{
			name: "basic push",
			config: PushConfig{
				URL:     server.URL,
				JobName: "test_job",
				AgentID: "test_agent",
			},
		},
		{
			name: "push with auth",
			config: PushConfig{
				URL:      server.URL,
				Username: "testuser",
				Password: "testpass",
				JobName:  "test_job",
				AgentID:  "test_agent",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu.Lock()
			receivedRequests = []string{} // Reset requests
			mu.Unlock()

			ctx := context.Background()
			err := pushMetrics(ctx, nil, tt.config, registry)
			require.NoError(t, err)

			mu.Lock()
			defer mu.Unlock()

			// Verify request was made
			require.NotEmpty(t, receivedRequests)
			require.Contains(t, receivedRequests[0], "PUT")
			require.Contains(t, receivedRequests[0], "/metrics/job/test_job/agent_id/test_agent")

			// Verify auth if provided
			if tt.config.Username != "" {
				require.Contains(t, receivedRequests, fmt.Sprintf("auth: %s:%s", tt.config.Username, tt.config.Password))
			}
		})
	}
}

func TestRunBasicValidation(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		exitCode int
	}{
		{
			name:     "missing required agent-id",
			args:     []string{"--push.gateway-url=http://localhost:9091"},
			exitCode: 1,
		},
		{
			name:     "missing required push gateway url",
			args:     []string{"--agent-id=test123"},
			exitCode: 1,
		},
		{
			name:     "invalid push interval",
			args:     []string{"--agent-id=test123", "--push.gateway-url=http://localhost:9091", "--push.interval=invalid"},
			exitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()


			exitCode := run(ctx, tt.args)
			require.Equal(t, tt.exitCode, exitCode)
		})
	}
}

func TestAgentCollectorWrapper(t *testing.T) {
	// Create a mock collector
	mockRegistry := prometheus.NewRegistry()
	testGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "test_metric",
		Help: "A test metric",
	})
	testGauge.Set(42)
	mockRegistry.MustRegister(testGauge)

	// Create agent wrapper
	wrapper := &AgentCollectorWrapper{
		agentID: "test_agent_123",
	}

	// Test Describe
	descs := make(chan *prometheus.Desc, 10)
	go func() {
		defer close(descs)
		wrapper.Describe(descs)
	}()

	// Collect descriptions
	var descriptions []*prometheus.Desc
	for desc := range descs {
		descriptions = append(descriptions, desc)
	}

	// Should have at least some descriptions
	require.NotEmpty(t, descriptions)
}

func TestPushConfig(t *testing.T) {
	config := PushConfig{
		URL:      "http://example.com:9091",
		Username: "user",
		Password: "pass",
		Interval: 30 * time.Second,
		AgentID:  "agent_001",
		JobName:  "windows_agent",
	}

	require.Equal(t, "http://example.com:9091", config.URL)
	require.Equal(t, "user", config.Username)
	require.Equal(t, "pass", config.Password)
	require.Equal(t, 30*time.Second, config.Interval)
	require.Equal(t, "agent_001", config.AgentID)
	require.Equal(t, "windows_agent", config.JobName)
}
