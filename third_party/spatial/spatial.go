// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package spatial

// #cgo CPPFLAGS: -I${SRCDIR}/include
// #cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/darwin-arm64
// #cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/darwin-amd64
// #cgo linux,arm64,!musl LDFLAGS: -L${SRCDIR}/lib/linux-arm64
// #cgo linux,amd64,!musl LDFLAGS: -L${SRCDIR}/lib/linux-amd64
// #cgo linux,arm64,musl LDFLAGS: -L${SRCDIR}/lib/linux-arm64-musl
// #cgo linux,amd64,musl LDFLAGS: -L${SRCDIR}/lib/linux-amd64-musl
// #cgo windows,amd64,!msvcrt LDFLAGS: -L${SRCDIR}/lib/windows-amd64
// #cgo windows,amd64,msvcrt LDFLAGS: -L${SRCDIR}/lib/windows-amd64-msvcrt
// #cgo LDFLAGS: -lgeos_c -lgeos -lproj -lsqlite3
// #cgo linux windows LDFLAGS: -lm
// #cgo linux LDFLAGS: -lpthread
// #include <geos_c.h>
// #include <proj.h>
import "C"

import "sync"

// projVersion holds the version of the PROJ library, read once since proj_info fills a shared buffer on every call.
var projVersion = sync.OnceValue(func() string {
	return C.GoString(C.proj_info().version)
})

// GEOSVersion returns the version of the GEOS library.
func GEOSVersion() string {
	return C.GoString(C.GEOSversion())
}

// PROJVersion returns the version of the PROJ library.
func PROJVersion() string {
	return projVersion()
}
