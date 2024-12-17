// Copyright 2024 Dolthub, Inc.
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

package id

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

const longString1 = `
0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789
0123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789`

func TestInternal(t *testing.T) {
	tests := []struct {
		section Section
		data    []string
	}{
		{Section_Table, []string{"exampleschema", "exampletable"}},
		{Section_Table, []string{"random$string", longString1, "bogus_data"}},
		{Section_Type, []string{`best "type" ever`, "worst type ever?"}},
	}
	for testIdx, test := range tests {
		t.Run(fmt.Sprintf("%d", testIdx), func(t *testing.T) {
			id := NewInternal(test.section, test.data...)
			for {
				require.True(t, id.IsValid())
				require.Equal(t, test.section, id.Section())
				data := id.Data()
				require.Len(t, data, len(test.data))
				for i := range data {
					require.Equal(t, test.data[i], data[i])
					require.Equal(t, test.data[i], id.Segment(i))
				}
				// If this is using the first format, then we'll rerun the test using a variant forced to the second format
				if !id.usesSecondFormat() {
					id = newInternalSecondFormat(test.section, test.data)
					continue
				}
				break
			}
		})
	}
}
