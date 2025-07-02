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

package pg_extension

/*
#cgo CFLAGS: "-I${SRCDIR}/library"
#include "exports.h"

static inline Datum CallFmgrFunctionC(FunctionCallInfo fcinfo) {
    return ((PGFunction)fcinfo->flinfo->fn_addr)(fcinfo);
}
*/
import "C"
import "unsafe"

// Datum is a C pointer to some data. Depending on the function being called, it may not be a pointer that should be
// freed, as some functions return pointers to static memory.
type Datum uintptr

// NullableDatum is used for arguments to Fmgr function calls.
type NullableDatum struct {
	Value  Datum
	IsNull bool
}

// CallFmgrFunction calls the given function and forwards the arguments.
func CallFmgrFunction(fn uintptr, args ...NullableDatum) (result Datum, isNotNull bool) {
	fi := Malloc[C.FmgrInfo]()
	defer Free(fi)
	ZeroMemory(fi)
	fc := Malloc[C.FunctionCallInfoBaseData]()
	defer Free(fc)
	ZeroMemory(fc)
	fi.fn_addr = unsafe.Pointer(fn)
	fc.flinfo = fi
	fc.nargs = C.int16_t(len(args))

	for i, arg := range args {
		fc.args[i].value = C.Datum(arg.Value)
		fc.args[i].isnull = C.bool(arg.IsNull)
	}
	result = Datum(C.CallFmgrFunctionC(fc))
	return result, !bool(fc.isnull) && result != 0
}
