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

typedef enum
{
    PG_MD5 = 0,
    PG_SHA1,
    PG_SHA224,
    PG_SHA256,
    PG_SHA384,
    PG_SHA512,
} pg_cryptohash_type;

typedef struct pg_cryptohash_ctx {
	pg_cryptohash_type hashType;
} pg_cryptohash_ctx;
*/
import "C"
import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
	"sync"
	"unsafe"
)

var pg_cryptohash_store sync.Map

//export pg_cryptohash_create
func pg_cryptohash_create(typ C.pg_cryptohash_type) *C.pg_cryptohash_ctx {
	ctx := (*C.pg_cryptohash_ctx)(C.malloc(C.size_t(unsafe.Sizeof(C.pg_cryptohash_ctx{}))))
	ctx.hashType = typ
	ctxPtr := uintptr(unsafe.Pointer(ctx))
	switch typ {
	case 1:
		pg_cryptohash_store.Store(ctxPtr, sha1.New())
	case C.PG_SHA224:
		pg_cryptohash_store.Store(ctxPtr, sha512.New512_224())
	case C.PG_SHA256:
		pg_cryptohash_store.Store(ctxPtr, sha256.New())
	case C.PG_SHA384:
		pg_cryptohash_store.Store(ctxPtr, sha512.New384())
	case C.PG_SHA512:
		pg_cryptohash_store.Store(ctxPtr, sha512.New())
	default:
		// Default to MD5
		pg_cryptohash_store.Store(ctxPtr, md5.New())
	}
	return ctx
}

//export pg_cryptohash_init
func pg_cryptohash_init(ctx *C.pg_cryptohash_ctx) C.int {
	if ctx == nil {
		return -1
	}
	return 0
}

//export pg_cryptohash_update
func pg_cryptohash_update(ctx *C.pg_cryptohash_ctx, data *C.pgext_const_uint8, len C.size_t) C.int {
	if ctx == nil {
		return -1
	}
	if len == 0 {
		return 0
	}
	ctxPtr := uintptr(unsafe.Pointer(ctx))
	storedHashAny, ok := pg_cryptohash_store.Load(ctxPtr)
	if !ok {
		return -1
	}
	storedHash := storedHashAny.(hash.Hash)
	dataSlice := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(len))
	if _, err := storedHash.Write(dataSlice); err != nil {
		return 1
	}
	return 0
}

//export pg_cryptohash_final
func pg_cryptohash_final(ctx *C.pg_cryptohash_ctx, dest *C.uint8_t, destLen C.size_t) C.int {
	if ctx == nil {
		return -1
	}
	ctxPtr := uintptr(unsafe.Pointer(ctx))
	storedHashAny, ok := pg_cryptohash_store.Load(ctxPtr)
	if !ok {
		return -1
	}
	storedHash := storedHashAny.(hash.Hash)
	sum := storedHash.Sum(nil)
	destSlice := unsafe.Slice((*byte)(unsafe.Pointer(dest)), int(destLen))
	// If the destination slice is too small, then it's invalid
	if len(sum) > len(destSlice) {
		return -1
	}
	copy(destSlice, sum)
	return 0
}

//export pg_cryptohash_free
func pg_cryptohash_free(ctx *C.pg_cryptohash_ctx) {
	if ctx != nil {
		ctxPtr := uintptr(unsafe.Pointer(ctx))
		pg_cryptohash_store.Delete(ctxPtr)
		C.free(unsafe.Pointer(ctx))
	}
}

//export pg_cryptohash_error
func pg_cryptohash_error(ctx *C.pg_cryptohash_ctx) *C.pgext_const_char {
	return C.CString("")
}
