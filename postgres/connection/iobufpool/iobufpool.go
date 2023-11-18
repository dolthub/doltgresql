// Copyright 2023 Dolthub, Inc.
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

package iobufpool

import "sync"

const minPoolExpOf2 = 8

var pools [18]*sync.Pool

func init() {
	for i := range pools {
		bufLen := 1 << (minPoolExpOf2 + i)
		pools[i] = &sync.Pool{
			New: func() any {
				buf := make([]byte, bufLen)
				return &buf
			},
		}
	}
}

// Get gets a []byte of len size with cap <= size*2.
func Get(size int) *[]byte {
	i := getPoolIdx(size)
	if i >= len(pools) {
		buf := make([]byte, size)
		return &buf
	}

	ptrBuf := (pools[i].Get().(*[]byte))
	*ptrBuf = (*ptrBuf)[:size]

	return ptrBuf
}

func getPoolIdx(size int) int {
	size--
	size >>= minPoolExpOf2
	i := 0
	for size > 0 {
		size >>= 1
		i++
	}

	return i
}

// Put returns buf to the pool.
func Put(buf *[]byte) {
	i := putPoolIdx(cap(*buf))
	if i < 0 {
		return
	}

	pools[i].Put(buf)
}

func putPoolIdx(size int) int {
	minPoolSize := 1 << minPoolExpOf2
	for i := range pools {
		if size == minPoolSize<<i {
			return i
		}
	}

	return -1
}
