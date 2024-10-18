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

package types

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/types"
)

// NewDomainType creates new instance of domain Type.
func NewDomainType(
	ctx *sql.Context,
	domain types.DomainType,
	owner string, // TODO
) (*Type, error) {
	passedByVal := false
	l := domain.AsType.MaxTextResponseByteLength(ctx)
	if l&1 == 0 && l < 9 {
		passedByVal = true
	}
	return &Type{
		Name:        domain.Name,
		Owner:       owner,
		Length:      int16(l),
		PassedByVal: passedByVal,
		Typ:         types.TypeType_Domain,
		Category:    domain.AsType.Category(),
		IsPreferred: domain.AsType.IsPreferredType(),
		IsDefined:   true,
		Delimiter:   ",",
		RelID:       0, // composite type only
		Subscript:   "",
		Elem:        0,
		Array:       0, // TODO: refers to array type of this type
		Input:       "domain_in",
		Output:      "",
		Receive:     "domain_recv",
		Send:        "",
		ModIn:       "-",
		ModOut:      "-",
		Analyze:     "-",
		Align:       domain.AsType.Alignment(),
		Storage:     types.TypeStorage_Plain, // TODO
		NotNull:     domain.NotNull,
		BaseTypeOID: domain.AsType.OID(),
		TypMod:      -1,
		NDims:       0,
		Collation:   0,
		DefaulBin:   "",
		Default:     domain.DefaultExpr,
		Acl:         "",
		Checks:      domain.Checks,
	}, nil
}
