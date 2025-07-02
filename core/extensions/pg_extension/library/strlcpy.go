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

//go:build !darwin

package extension_cgo

/*
#include "exports.h"
*/
import "C"
import "unsafe"

//export strlcpy
func strlcpy(dst *C.char, src *C.pgext_const_char, size C.size_t) C.size_t {
	var srcLen C.size_t
	for {
		if *(*C.char)(unsafe.Pointer(uintptr(unsafe.Pointer(src)) + uintptr(srcLen))) == 0 {
			break
		}
		srcLen++
	}
	if size != 0 {
		n := srcLen
		if n >= size {
			n = size - 1
		}
		dstSlice := unsafe.Slice((*byte)(unsafe.Pointer(dst)), int(n+1))
		srcSlice := unsafe.Slice((*byte)(unsafe.Pointer(src)), int(n))
		copy(dstSlice, srcSlice)
		dstSlice[n] = 0
	}
	return srcLen
}
