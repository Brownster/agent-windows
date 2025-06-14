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

package net_test

import (
	"testing"

	"github.com/alecthomas/kingpin/v2"
	"github.com/Brownster/agent-windows/internal/collector/net"
	"github.com/Brownster/agent-windows/internal/utils/testutils"
)

func BenchmarkCollector(b *testing.B) {
	// PrinterInclude is not set in testing context (kingpin flags not parsed), causing the collector to skip all interfaces.
	localNicInclude := ".+"

	testutils.FuncBenchmarkCollector(b, net.Name, net.NewWithFlags, func(app *kingpin.Application) {
		app.GetFlag("collector.net.nic-include").StringVar(&localNicInclude)
	})
}
