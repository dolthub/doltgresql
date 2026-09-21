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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initEncode registers the functions to the catalog.
func initEncode() {
	framework.RegisterFunction(encode)
}

// encode represents the PostgreSQL function of the same name, taking the same parameters.
var encode = framework.Function2{
	Name:       "encode",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		data, err := framework.UnwrapBytes(ctx, val1)
		if err != nil {
			return nil, err
		}
		format, err := framework.UnwrapString(ctx, val2)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(format) {
		case "hex":
			return hex.EncodeToString(data), nil
		case "base64":
			return base64.StdEncoding.EncodeToString(data), nil
		case "escape":
			return encodeEscape(data), nil
		default:
			return nil, fmt.Errorf(`unrecognized encoding: "%s"`, format)
		}
	},
}

// encodeEscape encodes data using PostgreSQL's "escape" format: bytes in the printable ASCII
// range are passed through as-is (with backslash doubled), while all other bytes are represented
// as a backslash followed by a three-digit octal value.
func encodeEscape(data []byte) string {
	var sb strings.Builder
	for _, b := range data {
		switch {
		case b == '\\':
			sb.WriteString(`\\`)
		case b < 0x20 || b > 0x7e:
			fmt.Fprintf(&sb, `\%03o`, b)
		default:
			sb.WriteByte(b)
		}
	}
	return sb.String()
}
