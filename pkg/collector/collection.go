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

package collector

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/Brownster/agent-windows/internal/collector/cpu"
	"github.com/Brownster/agent-windows/internal/collector/memory"
	"github.com/Brownster/agent-windows/internal/collector/net"
	"github.com/Brownster/agent-windows/internal/collector/pagefile"
	"github.com/Brownster/agent-windows/internal/mi"
	"github.com/Brownster/agent-windows/internal/types"
	"github.com/prometheus/client_golang/prometheus"
)

// NewWithFlags returns a new windows agent collector collection with kingpin flags registration
func NewWithFlags(app *kingpin.Application) Collection {
	collectors := Map{}

	// Only initialize our essential collectors
	if BuildersWithFlags["cpu"] != nil {
		collectors["cpu"] = BuildersWithFlags["cpu"](app)
	}
	if BuildersWithFlags["memory"] != nil {
		collectors["memory"] = BuildersWithFlags["memory"](app)
	}
	if BuildersWithFlags["net"] != nil {
		collectors["net"] = BuildersWithFlags["net"](app)
	}
	if BuildersWithFlags["pagefile"] != nil {
		collectors["pagefile"] = BuildersWithFlags["pagefile"](app)
	}

	return NewCollection(collectors)
}

// NewWithConfig returns a new windows agent collector collection with config
func NewWithConfig(config Config) Collection {
	collectors := Map{
		"cpu":      cpu.New(&config.CPU),
		"memory":   memory.New(&config.Memory),
		"net":      net.New(&config.Net),
		"pagefile": pagefile.New(&config.Pagefile),
	}

	return NewCollection(collectors)
}

// NewCollection returns a new windows agent collector collection
func NewCollection(collectors Map) Collection {
	return Collection{
		collectors: collectors,
		startTime:  time.Now(),
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "collector", "scrape_duration_seconds"),
			"windows_exporter: Time spent on collector scrape.",
			nil,
			nil,
		),
		collectorScrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "collector", "collector_duration_seconds"),
			"windows_exporter: Time spent on each collector.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "collector", "collector_success"),
			"windows_exporter: Whether the collector was successful.",
			[]string{"collector"},
			nil,
		),
		collectorScrapeTimeoutDesc: prometheus.NewDesc(
			prometheus.BuildFQName(types.Namespace, "collector", "collector_timeout"),
			"windows_exporter: Whether the collector timed out.",
			[]string{"collector"},
			nil,
		),
	}
}

// Enable enables collectors by name.
func (c *Collection) Enable(collectors []string) error {
	enabled := Map{}

	for _, name := range collectors {
		if collector, exists := c.collectors[name]; exists {
			enabled[name] = collector
		} else {
			return fmt.Errorf("collector %s not available", name)
		}
	}

	c.collectors = enabled

	return nil
}

// Build initializes all collectors in the collection.
func (c *Collection) Build(ctx context.Context, logger *slog.Logger) error {
	app, err := mi.ApplicationInitialize()
	if err != nil {
		return fmt.Errorf("failed to initialize MI application: %w", err)
	}

	session, err := app.NewSession(nil)
	if err != nil {
		return fmt.Errorf("failed to create MI session: %w", err)
	}

	c.miSession = session

	maxConcurrency := len(c.collectors)
	if maxConcurrency == 0 {
		maxConcurrency = 1
	}

	c.concurrencyCh = make(chan struct{}, maxConcurrency)

	for name, collector := range c.collectors {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := collector.Build(logger, c.miSession); err != nil {
			return fmt.Errorf("failed to build %s collector: %w", name, err)
		}
	}

	return nil
}

// Close closes all collectors in the collection.
func (c *Collection) Close() {
	for _, collector := range c.collectors {
		collector.Close()
	}

	if c.miSession != nil {
		c.miSession.Close()
	}
}

// Collectors returns a slice of collector names.
func (c *Collection) Collectors() []string {
	names := make([]string, 0, len(c.collectors))

	for name := range c.collectors {
		names = append(names, name)
	}

	slices.Sort(names)

	return names
}

// Describe implements prometheus.Collector interface.
func (c *Collection) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.scrapeDurationDesc
	ch <- c.collectorScrapeDurationDesc
	ch <- c.collectorScrapeSuccessDesc
	ch <- c.collectorScrapeTimeoutDesc
}

// Collect implements prometheus.Collector interface.
func (c *Collection) Collect(ch chan<- prometheus.Metric, logger *slog.Logger, maxScrapeDuration time.Duration) {
	c.collectAll(ch, logger, maxScrapeDuration)
}