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

package functions

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncateString(t *testing.T) {
	tests := []struct {
		inStr  string
		inLen  int32
		expStr string
		expLen int32
	}{
		{
			// ascii string is not truncated
			inStr:  "abc",
			inLen:  100,
			expStr: "abc",
			expLen: 3,
		},
		{
			// ascii string is truncated
			inStr:  "abcdefgh",
			inLen:  5,
			expStr: "abcde",
			expLen: 5,
		},
		{
			// non ascii string is not truncated
			inStr:  "こんにちは", // 5 characters 15 bytes
			inLen:  100,
			expStr: "こんにちは",
			expLen: 5,
		},
		{
			// non ascii string is truncated
			inStr:  "こんにちは", // 5 characters 15 bytes
			inLen:  3,
			expStr: "こんに",
			expLen: 3,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("truncateString(%s, %d)", test.inStr, test.inLen), func(t *testing.T) {
			outStr, outLen := truncateString(test.inStr, test.inLen)
			assert.Equal(t, test.expStr, outStr)
			assert.Equal(t, test.expLen, outLen)
		})
	}

}
