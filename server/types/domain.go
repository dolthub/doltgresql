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
)

// NewDomainType creates new instance of domain DoltgresType.
func NewDomainType(
	schema string,
	name string,
	asType DoltgresType,
	defaultExpr string,
	notNull bool,
	checks []*sql.CheckDefinition,
	owner string, // TODO
) (DoltgresType, error) {
	return DoltgresType{
		OID:           asType.OID, // TODO: generate unique OID, using underlying type OID for now
		Name:          name,
		Schema:        schema,
		Owner:         owner,
		Length:        asType.Length,
		PassedByVal:   asType.PassedByVal,
		TypType:       TypeType_Domain,
		TypCategory:   asType.TypCategory,
		IsPreferred:   asType.IsPreferred,
		IsDefined:     true,
		Delimiter:     ",",
		RelID:         0,
		SubscriptFunc: "",
		Elem:          0,
		Array:         0, // TODO: refers to array type of this type
		InputFunc:     "domain_in",
		OutputFunc:    asType.OutputFunc,
		ReceiveFunc:   "domain_recv",
		SendFunc:      asType.SendFunc,
		ModInFunc:     asType.ModInFunc,
		ModOutFunc:    asType.ModOutFunc,
		AnalyzeFunc:   "-",
		Align:         asType.Align,
		Storage:       asType.Storage,
		NotNull:       notNull,
		BaseTypeOID:   asType.OID,
		TypMod:        -1,
		NDims:         0,
		Collation:     0,
		DefaulBin:     "",
		Default:       defaultExpr,
		Acl:           "",
		Checks:        checks,
	}, nil
}
