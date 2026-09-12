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

package main

import (
	"testing"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

// TestRecordedCopyStreams verifies that adjacent COPY inputs and COPY output retain independent data streams.
func TestRecordedCopyStreams(t *testing.T) {
	firstInput := &pgproto3.CopyData{Data: []byte("1\n")}
	secondInput := &pgproto3.CopyData{Data: []byte("2\n")}
	output := &pgproto3.CopyData{Data: []byte("3\n")}
	messages := []pgproto3.Message{
		&pgproto3.CopyInResponse{}, firstInput, &pgproto3.CopyDone{}, &pgproto3.CommandComplete{},
		&pgproto3.CopyInResponse{}, secondInput, &pgproto3.CopyDone{}, &pgproto3.CommandComplete{},
		&pgproto3.CopyOutResponse{}, output, &pgproto3.CopyDone{}, &pgproto3.CommandComplete{},
	}

	var recorded recordedCopyStreams
	for _, message := range messages {
		recorded.record(message)
	}

	// Each server request must receive exactly its corresponding input stream, independently of COPY output.
	require.Equal(t, [][]*pgproto3.CopyData{{firstInput}, {secondInput}}, recorded.copyFromInputs)
	require.Equal(t, []*pgproto3.CopyData{output}, recorded.copyToOutput)
}
