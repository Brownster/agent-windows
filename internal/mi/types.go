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

package mi

import (
	"time"
	"unsafe"

	"github.com/Brownster/agent-windows/internal/utils"
	"golang.org/x/sys/windows"
)

type Boolean uint8

const (
	False Boolean = 0
	True  Boolean = 1
)

type QueryDialect *uint16

func NewQueryDialect(queryDialect string) (QueryDialect, error) {
	return windows.UTF16PtrFromString(queryDialect)
}

//nolint:gochecknoglobals
var QueryDialectWQL = utils.Must(NewQueryDialect("WQL"))

type Namespace *uint16

func NewNamespace(namespace string) (Namespace, error) {
	return windows.UTF16PtrFromString(namespace)
}

//nolint:gochecknoglobals
var (
	NamespaceRootCIMv2             = utils.Must(NewNamespace("root/CIMv2"))
	NamespaceRootWindowsFSRM       = utils.Must(NewNamespace("root/microsoft/windows/fsrm"))
	NamespaceRootWebAdministration = utils.Must(NewNamespace("root/WebAdministration"))
	NamespaceRootMSCluster         = utils.Must(NewNamespace("root/MSCluster"))
	NamespaceRootMicrosoftDNS      = utils.Must(NewNamespace("root/MicrosoftDNS"))
)

type Query *uint16

func NewQuery(query string) (Query, error) {
	return windows.UTF16PtrFromString(query)
}

// UTF16PtrFromString converts a string to a UTF-16 pointer at initialization time.
//
//nolint:ireturn
func UTF16PtrFromString[T *uint16](s string) T {
	val, err := windows.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}

	return val
}

type Timestamp struct {
	Year         uint32
	Month        uint32
	Day          uint32
	Hour         uint32
	Minute       uint32
	Second       uint32
	Microseconds uint32
	UTC          int32
}

type Interval struct {
	Days         uint32
	Hours        uint32
	Minutes      uint32
	Seconds      uint32
	Microseconds uint32
	Padding1     uint32
	Padding2     uint32
	Padding3     uint32
}

func NewInterval(interval time.Duration) *Interval {
	// Convert the duration to a number of microseconds
	microseconds := interval.Microseconds()

	// Create a new interval with the microseconds
	return &Interval{
		Days:         uint32(microseconds / (24 * 60 * 60 * 1000000)),
		Hours:        uint32(microseconds / (60 * 60 * 1000000)),
		Minutes:      uint32(microseconds / (60 * 1000000)),
		Seconds:      uint32(microseconds / 1000000),
		Microseconds: uint32(microseconds % 1000000),
	}
}

type Datetime struct {
	IsTimestamp bool
	Timestamp   *Timestamp // Used when IsTimestamp is true
	Interval    *Interval  // Used when IsTimestamp is false
}

type PropertyDecl struct {
	Flags         uint32
	Code          uint32
	Name          *uint16
	Mqualifiers   uintptr
	NumQualifiers uint32
	PropertyType  ValueType
	ClassName     *uint16
	Subscript     uint32
	Offset        uint32
	Origin        *uint16
	Propagator    *uint16
	Value         uintptr
}

func (c *ClassDecl) Properties() []*PropertyDecl {
	// Create a slice to hold the properties
	properties := make([]*PropertyDecl, c.NumProperties)

	// Mproperties is a pointer to an array of pointers to PropertyDecl
	propertiesArray := (**PropertyDecl)(unsafe.Pointer(c.Mproperties))

	// Iterate over the number of properties and fetch each property
	for i := range c.NumProperties {
		// Get the property pointer at index i
		propertyPtr := *(**PropertyDecl)(unsafe.Pointer(uintptr(unsafe.Pointer(propertiesArray)) + uintptr(i)*unsafe.Sizeof(uintptr(0))))

		// Append the property to the slice
		properties[i] = propertyPtr
	}

	// Return the slice of properties
	return properties
}
