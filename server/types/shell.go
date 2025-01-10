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
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/core/id"
)

// ErrTypeIsOnlyAShell is returned when given shell type is attempted to be used.
var ErrTypeIsOnlyAShell = errors.NewKind(`type "%s" is only a shell`)

// NewShellType creates new instance of shell DoltgresType.
func NewShellType(ctx *sql.Context, internalID id.InternalType) *DoltgresType {
	return &DoltgresType{
		ID:            internalID,
		TypLength:     4,
		PassedByVal:   true,
		TypType:       TypeType_Pseudo,
		TypCategory:   TypeCategory_PseudoTypes,
		IsPreferred:   false,
		IsDefined:     false,
		Delimiter:     ",",
		RelID:         id.Null,
		SubscriptFunc: toFuncID("-"),
		Elem:          id.NullType,
		Array:         id.NullType,
		InputFunc:     toFuncID("shell_in", toInternal("cstring")),
		OutputFunc:    toFuncID("shell_out", toInternal("void")),
		ReceiveFunc:   toFuncID("-"),
		SendFunc:      toFuncID("-"),
		ModInFunc:     toFuncID("-"),
		ModOutFunc:    toFuncID("-"),
		AnalyzeFunc:   toFuncID("-"),
		Align:         TypeAlignment_Int,
		Storage:       TypeStorage_Plain,
		NotNull:       false,
		BaseTypeID:    id.NullType,
		TypMod:        -1,
		NDims:         0,
		TypCollation:  id.NullCollation,
		DefaulBin:     "",
		Default:       "",
		Acl:           nil,
		Checks:        nil,
		attTypMod:     -1,
		CompareFunc:   toFuncID("-"),
	}
}
