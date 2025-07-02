// Copyright 2025 Dolthub, Inc.
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

//go:build linux

package pg_extension

/*
#cgo LDFLAGS: -ldl
#cgo LDFLAGS: -Wl,-E
#include <dlfcn.h>
#include <stdlib.h>
*/
import "C"

import (
	"fmt"
	"unsafe"

	_ "github.com/dolthub/doltgresql/core/extensions/pg_extension/library"
)

// PLATFORM specifies which platform applies to the current library loader. This will always be a three-letter string.
const PLATFORM = "LIN"

// unixLib is the Linux-specific implementation of InternalLoadedLibrary.
type unixLib struct {
	path   string
	handle unsafe.Pointer
}

var _ InternalLoadedLibrary = (*unixLib)(nil)

// loadLibraryInternal handles the loading of an extension's SO.
func loadLibraryInternal(path string) (InternalLoadedLibrary, error) {
	pathC := C.CString(path)
	defer C.free(unsafe.Pointer(pathC))

	handle := C.dlopen(pathC, C.RTLD_LAZY|C.RTLD_GLOBAL)
	if handle == nil {
		return nil, fmt.Errorf("error while loading extension `%s`\n%s", path, C.GoString(C.dlerror()))
	}
	return &unixLib{
		path:   path,
		handle: handle,
	}, nil
}

// Lookup implements the interface InternalLoadedLibrary.
func (u *unixLib) Lookup(sym string) (uintptr, error) {
	symC := C.CString(sym)
	defer C.free(unsafe.Pointer(symC))

	ptr := C.dlsym(u.handle, symC)
	if ptr == nil {
		return 0, fmt.Errorf("symbol %s not found", sym)
	}
	return uintptr(ptr), nil
}

// Close implements the interface InternalLoadedLibrary.
func (u *unixLib) Close() error {
	if C.dlclose(u.handle) != 0 {
		return fmt.Errorf("error while closing extension `%s`\n%s", u.path, C.GoString(C.dlerror()))
	}
	return nil
}
