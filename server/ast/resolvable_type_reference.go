// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"fmt"
	"strconv"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
)

type typeParams struct {
	name   string
	length *vitess.SQLVal
	scale  *vitess.SQLVal
}

func nodeResolvableTypeReference(typ tree.ResolvableTypeReference) (*typeParams, error) {
	if typ == nil {
		return nil, nil
	}

	var columnTypeName string
	var columnTypeLength *vitess.SQLVal
	var columnTypeScale *vitess.SQLVal
	switch columnType := typ.(type) {
	case *tree.ArrayTypeReference:
		return nil, fmt.Errorf("array types are not yet supported")
	case *tree.OIDTypeReference:
		return nil, fmt.Errorf("referencing types by their OID is not yet supported")
	case *tree.UnresolvedObjectName:
		return nil, fmt.Errorf("type declaration format is not yet supported")
	case *types.GeoMetadata:
		return nil, fmt.Errorf("geometry types are not yet supported")
	case *types.T:
		columnTypeName = columnType.SQLStandardName()
		switch columnType.Family() {
		case types.DecimalFamily:
			columnTypeLength = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Precision()))))
			columnTypeScale = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Scale()))))
		case types.JsonFamily:
			columnTypeName = "JSON"
		case types.StringFamily:
			columnTypeLength = vitess.NewIntVal([]byte(strconv.Itoa(int(columnType.Width()))))
		case types.TimestampFamily:
			columnTypeName = columnType.Name()
		}
	}

	return &typeParams{
		name:   columnTypeName,
		length: columnTypeLength,
		scale:  columnTypeScale,
	}, nil
}
