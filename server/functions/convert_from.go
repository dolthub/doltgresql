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
	"unicode/utf8"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initConvertFrom registers the functions to the catalog.
func initConvertFrom() {
	framework.RegisterFunction(convert_from_bytea_name)
}

// convert_from_bytea_name represents the PostgreSQL function of the same name, taking the same parameters.
var convert_from_bytea_name = framework.Function2{
	Name:       "convert_from",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Bytea, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		input, err := framework.UnwrapBytes(ctx, val1)
		if err != nil {
			return nil, err
		}
		encodingName, err := framework.UnwrapString(ctx, val2)
		if err != nil {
			return nil, err
		}

		source := lookupPostgresEncoding(encodingName)
		if source == nil {
			return nil, pgerror.WithCandidateCode(
				fmt.Errorf(`invalid source encoding name "%s"`, encodingName), pgcode.InvalidParameterValue)
		}
		if source.passThrough {
			if err = validUTF8(input); err != nil {
				return nil, err
			}
			return string(input), nil
		}
		if source.encoder == nil {
			return nil, pgerror.WithCandidateCode(fmt.Errorf(
				`source encoding "%s" is recognized but not yet supported; request support at %s`,
				source.name, encodingSupportIssuesURL), pgcode.FeatureNotSupported)
		}

		converted, err := source.encoder.NewDecoder().Bytes(input)
		if err != nil {
			return nil, pgerror.WithCandidateCode(
				fmt.Errorf(`invalid byte sequence for encoding "%s"`, source.name), pgcode.CharacterNotInRepertoire)
		}
		return string(converted), nil
	},
}

// validUTF8 returns an error naming the first byte that is not part of a valid UTF-8 sequence.
func validUTF8(input []byte) error {
	for i := 0; i < len(input); {
		r, size := utf8.DecodeRune(input[i:])
		if r == utf8.RuneError && size <= 1 {
			return pgerror.WithCandidateCode(
				fmt.Errorf(`invalid byte sequence for encoding "UTF8": 0x%02x`, input[i]), pgcode.CharacterNotInRepertoire)
		}
		i += size
	}
	return nil
}
