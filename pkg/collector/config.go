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
	"github.com/Brownster/agent-windows/internal/collector/cpu"
	"github.com/Brownster/agent-windows/internal/collector/memory"
	"github.com/Brownster/agent-windows/internal/collector/net"
	"github.com/Brownster/agent-windows/internal/collector/pagefile"
)

// Config provides configuration for the windows_agent_collector
type Config struct {
	CPU      cpu.Config      `yaml:"cpu"`
	Memory   memory.Config   `yaml:"memory"`
	Net      net.Config      `yaml:"net"`
	Pagefile pagefile.Config `yaml:"pagefile"`
}

//nolint:gochecknoglobals
var ConfigDefaults = Config{
	CPU:      cpu.ConfigDefaults,
	Memory:   memory.ConfigDefaults,
	Net:      net.ConfigDefaults,
	Pagefile: pagefile.ConfigDefaults,
}