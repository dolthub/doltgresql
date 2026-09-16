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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDecode registers the functions to the catalog.
func initDecode() {
	framework.RegisterFunction(decode)
}

// errInvalidByteaSyntax is returned when the "escape" format input cannot be decoded.
var errInvalidByteaSyntax = pgerror.WithCandidateCode(
	errors.New("invalid input syntax for type bytea"), pgcode.InvalidTextRepresentation)

// decode represents the PostgreSQL function of the same name, taking the same parameters.
var decode = framework.Function2{
	Name:       "decode",
	Return:     pgtypes.Bytea,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		data, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		format, err := framework.UnwrapString(ctx, val2)
		if err != nil {
			return nil, err
		}
		switch strings.ToLower(format) {
		case "hex":
			decoded, err := hex.DecodeString(strings.Join(strings.Fields(data), ""))
			var invalidByte hex.InvalidByteError
			if errors.As(err, &invalidByte) {
				return nil, pgerror.WithCandidateCode(
					fmt.Errorf(`invalid hexadecimal digit: "%c"`, rune(invalidByte)), pgcode.InvalidParameterValue)
			} else if err != nil {
				return nil, pgerror.WithCandidateCode(
					errors.New("invalid hexadecimal data: odd number of digits"), pgcode.InvalidParameterValue)
			}
			return decoded, nil
		case "base64":
			decoded, err := base64.StdEncoding.DecodeString(strings.Join(strings.Fields(data), ""))
			if err != nil {
				return nil, pgerror.WithCandidateCode(
					errors.New("invalid symbol found while decoding base64 sequence"), pgcode.InvalidParameterValue)
			}
			return decoded, nil
		case "escape":
			decoded := make([]byte, 0, len(data))
			for i := 0; i < len(data); i++ {
				if data[i] != '\\' {
					decoded = append(decoded, data[i])
				} else if i+1 < len(data) && data[i+1] == '\\' {
					decoded = append(decoded, '\\')
					i++
				} else if i+3 >= len(data) {
					return nil, errInvalidByteaSyntax
				} else if b, err := strconv.ParseUint(data[i+1:i+4], 8, 8); err != nil {
					return nil, errInvalidByteaSyntax
				} else {
					decoded = append(decoded, byte(b))
					i += 3
				}
			}
			return decoded, nil
		default:
			return nil, fmt.Errorf(`unrecognized encoding: "%s"`, format)
		}
	},
}
