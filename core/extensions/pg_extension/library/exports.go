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

package extension_cgo

/*
#include "exports.h"

static inline Datum FunctionPassthrough(PGFunction f, FunctionCallInfoBaseData *fcinfo) {
	return (*f)(fcinfo);
}
*/
import "C"
import (
	"encoding/hex"
	"fmt"
	"os"
	"unsafe"
)

func main() {}

//export errcode
func errcode(code C.int) C.int {
	return code
}

//export palloc
func palloc(sz C.size_t) unsafe.Pointer {
	// TODO: should track this pointer so we know to free it later
	return C.malloc(sz)
}

//export palloc0
func palloc0(sz C.size_t) unsafe.Pointer {
	// TODO: should track this pointer so we know to free it later
	ptr := C.malloc(sz)
	if ptr != nil {
		C.memset(ptr, 0, sz)
	}
	return ptr
}

//export MemoryContextAlloc
func MemoryContextAlloc(c unsafe.Pointer, sz C.size_t) unsafe.Pointer {
	// TODO: should track this pointer so we know to free it later, could use the memory context
	return C.malloc(sz)
}

//export MemoryContextAllocExtended
func MemoryContextAllocExtended(c unsafe.Pointer, sz C.size_t, f C.int) unsafe.Pointer {
	// TODO: should track this pointer so we know to free it later, could use the memory context
	return C.malloc(sz)
}

//export pg_detoast_datum_packed
func pg_detoast_datum_packed(d unsafe.Pointer) unsafe.Pointer {
	return d
}

//export text_to_cstring
func text_to_cstring(t unsafe.Pointer) *C.char {
	return C.CString(C.GoString((*C.char)(t)))
}

//export uuid_in
func uuid_in(fc C.FunctionCallInfo) C.Datum {
	uuidInputStr := (*C.pgext_const_char)(unsafe.Pointer(uintptr(fc.args[0].value)))
	uuidInputBytes := decodeUuidStr([]byte(C.GoString(uuidInputStr)))
	outputBytes := (*C.pgext_unsigned_char)(C.malloc(C.size_t(len(uuidInputBytes))))
	C.memcpy(unsafe.Pointer(outputBytes), unsafe.Pointer(&uuidInputBytes[0]), C.size_t(len(uuidInputBytes)))
	return C.Datum(uintptr(unsafe.Pointer(outputBytes)))
}

// decodeUuidStr is a helper function for uuid_in, which converts the given byte slice to the static array representation.
func decodeUuidStr(strBytes []byte) [16]byte {
	if strBytes[8] != '-' || strBytes[13] != '-' || strBytes[18] != '-' || strBytes[23] != '-' {
		return [16]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
	}
	u := [16]byte{}
	src := strBytes
	dst := u[:]
	for i, byteGroup := range []int{8, 4, 4, 4, 12} {
		if i > 0 {
			src = src[1:]
		}
		_, err := hex.Decode(dst[:byteGroup/2], src[:byteGroup])
		if err != nil {
			return [16]byte{255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255}
		}
		src = src[byteGroup:]
		dst = dst[byteGroup/2:]
	}
	return u
}

//export uuid_out
func uuid_out(ptr unsafe.Pointer) C.Datum {
	uuidInputBytes := C.GoBytes(ptr, 16)
	textBuffer := make([]byte, 36)
	_ = hex.Encode(textBuffer[0:8], uuidInputBytes[0:4])
	textBuffer[8] = '-'
	_ = hex.Encode(textBuffer[9:13], uuidInputBytes[4:6])
	textBuffer[13] = '-'
	_ = hex.Encode(textBuffer[14:18], uuidInputBytes[6:8])
	textBuffer[18] = '-'
	_ = hex.Encode(textBuffer[19:23], uuidInputBytes[8:10])
	textBuffer[23] = '-'
	_ = hex.Encode(textBuffer[24:], uuidInputBytes[10:])
	return C.Datum(uintptr(unsafe.Pointer(C.CString(string(textBuffer)))))
}

//export DirectFunctionCall1Coll
func DirectFunctionCall1Coll(fn unsafe.Pointer, collation C.uint32_t, arg1 C.Datum) C.Datum {
	fc := (*C.FunctionCallInfoBaseData)(C.malloc(C.SZ_FCINFO))
	if fc == nil {
		_, _ = fmt.Fprintln(os.Stderr, "DirectFunctionCall1Coll: out of memory")
		return 0
	}
	defer C.free(unsafe.Pointer(fc))
	C.memset(unsafe.Pointer(fc), 0, C.SZ_FCINFO)

	fc.isnull = false
	fc.fncollation = collation
	fc.nargs = 1
	fc.args[0].value = arg1
	fc.args[0].isnull = false

	result := C.FunctionPassthrough(C.PGFunction(fn), fc)
	if fc.isnull {
		_, _ = fmt.Fprintf(os.Stderr, "function %p returned NULL\n", fn)
	}
	return result
}
