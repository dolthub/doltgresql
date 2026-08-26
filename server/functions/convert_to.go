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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

const encodingSupportIssuesURL = "https://github.com/dolthub/doltgresql/issues"

// initConvertTo registers the functions to the catalog.
func initConvertTo() {
	framework.RegisterFunction(convert_to_text_name)
}

// convert_to_text_name represents the PostgreSQL function of the same name, taking the same parameters.
var convert_to_text_name = framework.Function2{
	Name:       "convert_to",
	Return:     pgtypes.Bytea,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Name},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		input, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		encodingName, err := framework.UnwrapString(ctx, val2)
		if err != nil {
			return nil, err
		}

		destination := lookupPostgresEncoding(encodingName)
		if destination == nil {
			return nil, pgerror.WithCandidateCode(
				fmt.Errorf(`invalid destination encoding name "%s"`, encodingName), pgcode.InvalidParameterValue)
		}
		if destination.passThrough {
			return []byte(input), nil
		}
		if destination.encoder == nil {
			return nil, pgerror.WithCandidateCode(fmt.Errorf(
				`destination encoding "%s" is recognized but not yet supported; request support at %s`,
				destination.name, encodingSupportIssuesURL), pgcode.FeatureNotSupported)
		}

		converted, err := destination.encoder.NewEncoder().Bytes([]byte(input))
		if err != nil {
			return nil, pgerror.WithCandidateCode(untranslatableCharacterError(input, *destination),
				pgcode.UntranslatableCharacter)
		}
		return converted, nil
	},
}

// untranslatableCharacterError describes the first input character unsupported by destination.
func untranslatableCharacterError(input string, destination postgresEncoding) error {
	for _, r := range input {
		encodedRune := string(r)
		if _, err := destination.encoder.NewEncoder().String(encodedRune); err != nil {
			bytes := []byte(encodedRune)
			parts := make([]string, len(bytes))
			for i, b := range bytes {
				parts[i] = fmt.Sprintf("0x%02x", b)
			}
			return fmt.Errorf(`character with byte sequence %s in encoding "UTF8" has no equivalent in encoding "%s"`,
				strings.Join(parts, " "), destination.name)
		}
	}
	return fmt.Errorf(`character has no equivalent in encoding "%s"`, destination.name)
}
