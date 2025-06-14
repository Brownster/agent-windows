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

package testutils

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/alecthomas/kingpin/v2"
	"github.com/Brownster/agent-windows/internal/mi"
	"github.com/Brownster/agent-windows/internal/pdh"
	"github.com/Brownster/agent-windows/pkg/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

func FuncBenchmarkCollector[C collector.Collector](b *testing.B, name string, collectFunc collector.BuilderWithFlags[C], fn ...func(app *kingpin.Application)) {
	b.Helper()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	app := kingpin.New("windows_exporter", "Windows metrics exporter.")
	c := collectFunc(app)

	for _, f := range fn {
		f(app)
	}

	collectors := collector.NewCollection(map[string]collector.Collector{name: c})
	require.NoError(b, collectors.Build(b.Context(), logger))

	metrics := make(chan prometheus.Metric)

	go func() {
		for {
			<-metrics
		}
	}()

	for b.Loop() {
		require.NoError(b, c.Collect(metrics))
	}
}

func TestCollector[C collector.Collector, V interface{}](t *testing.T, fn func(*V) C, conf *V) {
	t.Helper()

	var (
		metrics []prometheus.Metric
		err     error
	)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := fn(conf)
	ch := make(chan prometheus.Metric, 10000)

	miApp, err := mi.ApplicationInitialize()
	require.NoError(t, err)

	miSession, err := miApp.NewSession(nil)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, c.Close())
		require.NoError(t, miSession.Close())
		require.NoError(t, miApp.Close())
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()

		for metric := range ch {
			metrics = append(metrics, metric)
		}
	}()

	err = c.Build(logger, miSession)

	switch {
	case err == nil:
	case errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE),
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoCounter)),
		errors.Is(err, pdh.NewPdhError(pdh.CstatusNoObject)),
		errors.Is(err, os.ErrNotExist):
	default:
		require.NoError(t, err)
	}

	time.Sleep(1 * time.Second)

	err = c.Collect(ch)

	switch {
	// container collector
	case errors.Is(err, windows.Errno(2151088411)),
		errors.Is(err, pdh.ErrPerformanceCounterNotInitialized),
		errors.Is(err, pdh.ErrNoData),
		errors.Is(err, mi.MI_RESULT_INVALID_NAMESPACE),
		errors.Is(err, mi.MI_RESULT_INVALID_QUERY):
		t.Skip("collector not supported on this system")
	default:
		require.NoError(t, err)
	}

	close(ch)

	wg.Wait()
}
