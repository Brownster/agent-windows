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
	"maps"
	"slices"

	"github.com/alecthomas/kingpin/v2"
	"github.com/Brownster/agent-windows/internal/collector/cpu"
	"github.com/Brownster/agent-windows/internal/collector/memory"
	"github.com/Brownster/agent-windows/internal/collector/net"
	"github.com/Brownster/agent-windows/internal/collector/pagefile"
)

func NewBuilderWithFlags[C Collector](fn BuilderWithFlags[C]) BuilderWithFlags[Collector] {
	return func(app *kingpin.Application) Collector {
		return fn(app)
	}
}

//nolint:gochecknoglobals
var BuildersWithFlags = map[string]BuilderWithFlags[Collector]{
	cpu.Name:      NewBuilderWithFlags(cpu.NewWithFlags),
	memory.Name:   NewBuilderWithFlags(memory.NewWithFlags),
	net.Name:      NewBuilderWithFlags(net.NewWithFlags),
	pagefile.Name: NewBuilderWithFlags(pagefile.NewWithFlags),
}

// Available returns a sorted list of available collectors.
//
//goland:noinspection GoUnusedExportedFunction
func Available() []string {
	return slices.Sorted(maps.Keys(BuildersWithFlags))
}