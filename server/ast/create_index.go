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
	"math"
	"slices"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// nodeCreateIndex handles *tree.CreateIndex nodes.
func nodeCreateIndex(ctx *Context, node *tree.CreateIndex) (*vitess.AlterTable, error) {
	if node == nil {
		return nil, nil
	}
	if node.Concurrently {
		return nil, errors.Errorf("concurrent index creation is not yet supported")
	}
	var vectorOptions []*vitess.IndexOption
	columns := node.Columns
	indexParams := node.IndexParams
	method := strings.ToLower(node.Using)
	if method != "" && method != "btree" {
		if _, ok := extensions.GetAccessMethod(method); !ok {
			return nil, errors.Errorf("index method %s is not yet supported", node.Using)
		}
		var err error
		if vectorOptions, err = vectorIndexOptions(node, method); err != nil {
			return nil, err
		}
		columns = make(tree.IndexElemList, len(node.Columns))
		copy(columns, node.Columns)
		columns[0].OpClass = nil
		indexParams.StorageParams = nil
	}
	indexDef, err := nodeIndexTableDef(ctx, &tree.IndexTableDef{
		Name:        node.Name,
		Columns:     columns,
		IndexParams: indexParams,
	})
	if err != nil {
		return nil, err
	}
	tableName, err := nodeTableName(ctx, &node.Table)
	if err != nil {
		return nil, err
	}
	var indexType string
	if node.Unique {
		indexType = vitess.UniqueStr
	}
	if vectorOptions != nil {
		indexType = vitess.VectorStr
	}
	var predicate vitess.Expr
	if node.Predicate != nil {
		predicate, err = nodeExpr(ctx, node.Predicate)
		if err != nil {
			return nil, err
		}
	}
	return &vitess.AlterTable{
		Table: tableName,
		Statements: []*vitess.DDL{
			{
				Action:      vitess.AlterStr,
				Table:       tableName,
				IfNotExists: node.IfNotExists,
				IndexSpec: &vitess.IndexSpec{
					Action:    vitess.CreateStr,
					FromName:  indexDef.Info.Name,
					ToName:    indexDef.Info.Name,
					Type:      indexType,
					Fields:    indexDef.Fields,
					Options:   append(indexDef.Options, vectorOptions...),
					Predicate: predicate,
				},
			},
		},
	}, nil
}

// vectorIndexOptions validates a CREATE INDEX statement that uses a vector access method
func vectorIndexOptions(node *tree.CreateIndex, method string) ([]*vitess.IndexOption, error) {
	if node.Unique {
		return nil, errors.Errorf(`access method "%s" does not support unique indexes`, method)
	}
	if node.IndexParams.IncludeColumns != nil {
		return nil, errors.Errorf(`access method "%s" does not support included columns`, method)
	}
	if len(node.Columns) != 1 {
		return nil, errors.Errorf(`access method "%s" does not support multicolumn indexes`, method)
	}
	if node.Predicate != nil {
		return nil, errors.Errorf("partial vector indexes are not yet supported")
	}
	column := node.Columns[0]
	if column.Expr != nil {
		return nil, errors.Errorf("expression columns in vector indexes are not yet supported")
	}
	options := []*vitess.IndexOption{
		{Name: sql.VectorAccessMethodOptionName, Value: vitess.NewStrVal([]byte(method))},
	}
	if column.OpClass != nil {
		if len(column.OpClass.Options) > 0 {
			return nil, errors.Errorf(`operator class "%s" has no options`, column.OpClass.Name)
		}
		opclass, ok := extensions.GetOperatorClass(column.OpClass.Name)
		if !ok || !slices.Contains(opclass.AccessMethods, method) {
			return nil, errors.Errorf(`operator class "%s" does not exist for access method "%s"`, column.OpClass.Name, method)
		}
		if opclass.DistanceType == nil {
			return nil, errors.Errorf(`operator class "%s" is not yet supported`, column.OpClass.Name)
		}
		options = append(options,
			&vitess.IndexOption{Name: sql.VectorOpClassOptionName, Value: vitess.NewStrVal([]byte(column.OpClass.Name))},
			&vitess.IndexOption{Name: sql.VectorDistanceTypeOptionName, Value: vitess.NewStrVal([]byte(opclass.DistanceType.String()))},
		)
	}
	if err := validateVectorStorageParams(method, node.IndexParams.StorageParams); err != nil {
		return nil, err
	}
	return options, nil
}

// validateVectorStorageParams checks the given storage parameters against the access method's declared parameters.
// The values are otherwise unused, as the underlying index has no equivalent tuning knobs.
func validateVectorStorageParams(method string, storageParams tree.StorageParams) error {
	am, _ := extensions.GetAccessMethod(method)
	values := make(map[string]int64, len(am.Params))
	for _, param := range am.Params {
		values[param.Name] = param.Default
	}
	for _, storageParam := range storageParams {
		name := strings.ToLower(string(storageParam.Key))
		var declared extdef.AccessMethodParam
		for _, param := range am.Params {
			if param.Name == name {
				declared = param
				break
			}
		}
		if declared.Name == "" {
			return errors.Errorf(`unrecognized parameter "%s"`, name)
		}
		str := tree.AsString(storageParam.Value)
		if strVal, ok := storageParam.Value.(*tree.StrVal); ok {
			str = strVal.RawString()
		}
		value, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			floatValue, err := strconv.ParseFloat(str, 64)
			if err != nil {
				return errors.Errorf(`invalid value for integer option "%s": %s`, name, str)
			}
			value = int64(math.RoundToEven(floatValue))
		}
		if err != nil {
			return err
		}
		if value < declared.Min || value > declared.Max {
			return errors.Errorf(`value %d out of bounds for option "%s"`, value, name)
		}
		values[name] = value
	}
	if am.CheckParams != nil {
		if err := am.CheckParams(values); err != nil {
			return err
		}
	}
	if len(storageParams) > 0 {
		logrus.Warnf("storage parameters are not used by %s indexes, ignoring", method)
	}
	return nil
}
