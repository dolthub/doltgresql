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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"
)

// ErrDomainDoesNotAllowNullValues is returned when given value is NULL and a domain is non-nullable.
var ErrDomainDoesNotAllowNullValues = errors.NewKind(`domain %s does not allow null values`)

// ErrDomainValueViolatesCheckConstraint is returned when given value violates a domain check.
var ErrDomainValueViolatesCheckConstraint = errors.NewKind(`value for domain %s violates check constraint "%s"`)

// NewDomainType creates new instance of domain DoltgresType.
func NewDomainType(
	ctx *sql.Context,
	schema string,
	name string,
	asType *DoltgresType,
	defaultExpr string,
	notNull bool,
	checks []*sql.CheckDefinition,
	owner string, // TODO
) *DoltgresType {
	return &DoltgresType{
		OID:           asType.OID, // TODO: generate unique OID, using underlying type OID for now
		Name:          name,
		Schema:        schema,
		Owner:         owner,
		TypLength:     asType.TypLength,
		PassedByVal:   asType.PassedByVal,
		TypType:       TypeType_Domain,
		TypCategory:   asType.TypCategory,
		IsPreferred:   asType.IsPreferred,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         0,
		SubscriptFunc: toFuncID("-"),
		Elem:          0,
		Array:         0, // TODO: refers to array type of this type
		InputFunc:     toFuncID("domain_in", oid.T_cstring, oid.T_oid, oid.T_int4),
		OutputFunc:    asType.OutputFunc,
		ReceiveFunc:   toFuncID("domain_recv", oid.T_internal, oid.T_oid, oid.T_int4),
		SendFunc:      asType.SendFunc,
		ModInFunc:     asType.ModInFunc,
		ModOutFunc:    asType.ModOutFunc,
		AnalyzeFunc:   toFuncID("-"),
		Align:         asType.Align,
		Storage:       asType.Storage,
		NotNull:       notNull,
		BaseTypeOID:   asType.OID,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  0,
		DefaulBin:     "",
		Default:       defaultExpr,
		Acl:           nil,
		Checks:        checks,
		attTypMod:     -1,
		CompareFunc:   asType.CompareFunc,
	}
}
