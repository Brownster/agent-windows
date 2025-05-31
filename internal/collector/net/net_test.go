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

	"github.com/Brownster/agent-windows/internal/collector/net"
	"github.com/Brownster/agent-windows/internal/utils/testutils"
	"github.com/stretchr/testify/require"
)

func TestCollector(t *testing.T) {
	testutils.TestCollector(t, net.New, nil)
}

func TestGetInterfaceType(t *testing.T) {
	tests := []struct {
		name         string
		ifType       uint32
		friendlyName string
		expected     string
	}{
		{
			name:         "ethernet by type",
			ifType:       6, // IF_TYPE_ETHERNET_CSMACD
			friendlyName: "Ethernet",
			expected:     "ethernet",
		},
		{
			name:         "wifi by friendly name",
			ifType:       0, // unknown type
			friendlyName: "Intel(R) Wi-Fi 6 AX200 160MHz",
			expected:     "wifi",
		},
		{
			name:         "ethernet by friendly name",
			ifType:       0,
			friendlyName: "Realtek PCIe GbE Family Controller",
			expected:     "ethernet",
		},
		{
			name:         "vpn by friendly name",
			ifType:       0,
			friendlyName: "TAP-Windows Adapter V9",
			expected:     "vpn",
		},
		{
			name:         "cellular by friendly name",
			ifType:       0,
			friendlyName: "Mobile Broadband Adapter",
			expected:     "cellular",
		},
		{
			name:         "wifi by wlan pattern",
			ifType:       0,
			friendlyName: "Qualcomm Atheros QCA9377 Wireless Network Adapter",
			expected:     "wifi",
		},
		{
			name:         "gigabit ethernet",
			ifType:       0,
			friendlyName: "Intel(R) Ethernet Connection I217-LM",
			expected:     "ethernet",
		},
		{
			name:         "virtual adapter",
			ifType:       0,
			friendlyName: "VMware Virtual Ethernet Adapter",
			expected:     "vpn",
		},
		{
			name:         "unknown interface",
			ifType:       999,
			friendlyName: "Unknown Adapter",
			expected:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := net.GetInterfaceType(tt.ifType, tt.friendlyName)
			require.Equal(t, tt.expected, result)
		})
	}
}
