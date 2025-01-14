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
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/core/id"

	"github.com/dolthub/go-mysql-server/sql"
)

// ErrDomainDoesNotAllowNullValues is returned when given value is NULL and a domain is non-nullable.
var ErrDomainDoesNotAllowNullValues = errors.NewKind(`domain %s does not allow null values`)

// ErrDomainValueViolatesCheckConstraint is returned when given value violates a domain check.
var ErrDomainValueViolatesCheckConstraint = errors.NewKind(`value for domain %s violates check constraint "%s"`)

// NewDomainType creates new instance of domain DoltgresType.
func NewDomainType(
	ctx *sql.Context,
	asType *DoltgresType,
	defaultExpr string,
	notNull bool,
	checks []*sql.CheckDefinition,
	arrayID, internalID id.Type,
) *DoltgresType {
	return &DoltgresType{
		ID:            internalID,
		TypLength:     asType.TypLength,
		PassedByVal:   asType.PassedByVal,
		TypType:       TypeType_Domain,
		TypCategory:   asType.TypCategory,
		IsPreferred:   asType.IsPreferred,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         id.Null,
		SubscriptFunc: toFuncID("-"),
		Elem:          id.NullType,
		Array:         arrayID,
		InputFunc:     toFuncID("domain_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
		OutputFunc:    asType.OutputFunc,
		ReceiveFunc:   toFuncID("domain_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
		SendFunc:      asType.SendFunc,
		ModInFunc:     asType.ModInFunc,
		ModOutFunc:    asType.ModOutFunc,
		AnalyzeFunc:   toFuncID("-"),
		Align:         asType.Align,
		Storage:       asType.Storage,
		NotNull:       notNull,
		BaseTypeID:    asType.ID,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  id.NullCollation,
		DefaulBin:     "",
		Default:       defaultExpr,
		Acl:           nil,
		Checks:        checks,
		attTypMod:     -1,
		CompareFunc:   asType.CompareFunc,
	}
}
