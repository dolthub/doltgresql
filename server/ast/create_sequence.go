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
	"math"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateSequence handles *tree.CreateSequence nodes.
func nodeCreateSequence(ctx *Context, node *tree.CreateSequence) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Persistence.IsTemporary() {
		return nil, fmt.Errorf("temporary sequences are not yet supported")
	}
	if node.Persistence.IsUnlogged() {
		return nil, fmt.Errorf("unlogged sequences are not yet supported")
	}
	name, err := nodeTableName(ctx, &node.Name)
	if err != nil {
		return nil, err
	}
	if len(name.DbQualifier.String()) > 0 {
		return nil, fmt.Errorf("CREATE SEQUENCE is currently only supported for the current database")
	}
	// Read all options and check whether they've been set (if not, we'll use the defaults)
	minValueLimit := int64(math.MinInt64)
	maxValueLimit := int64(math.MaxInt64)
	increment := int64(1)
	var minValue int64
	var maxValue int64
	var start int64
	var dataType *pgtypes.DoltgresType
	var ownerTableName string
	var ownerColumnName string
	minValueSet := false
	maxValueSet := false
	incrementSet := false
	startSet := false
	cycle := false
	for _, option := range node.Options {
		switch option.Name {
		case tree.SeqOptAs:
			if !dataType.IsEmptyType() {
				return nil, fmt.Errorf("conflicting or redundant options")
			}
			_, dataType, err = nodeResolvableTypeReference(ctx, option.AsType)
			if err != nil {
				return nil, err
			}
			switch dataType.ID {
			case pgtypes.Int16.ID:
				minValueLimit = int64(math.MinInt16)
				maxValueLimit = int64(math.MaxInt16)
			case pgtypes.Int32.ID:
				minValueLimit = int64(math.MinInt32)
				maxValueLimit = int64(math.MaxInt32)
			case pgtypes.Int64.ID:
				minValueLimit = int64(math.MinInt64)
				maxValueLimit = int64(math.MaxInt64)
			default:
				return nil, fmt.Errorf("sequence type must be smallint, integer, or bigint")
			}
		case tree.SeqOptCycle:
			cycle = true
		case tree.SeqOptNoCycle:
			cycle = false
		case tree.SeqOptOwnedBy:
			expr, err := nodeExpr(ctx, option.ColumnItemVal)
			if err != nil {
				return nil, err
			}
			colName, ok := expr.(*vitess.ColName)
			if !ok {
				return nil, fmt.Errorf("expected sequence owner to be a table and column name")
			}
			if len(colName.Qualifier.SchemaQualifier.String()) > 0 || len(colName.Qualifier.DbQualifier.String()) > 0 {
				return nil, fmt.Errorf("sequence owner must be in the same schema as the sequence")
			}
			ownerTableName = colName.Qualifier.Name.String()
			ownerColumnName = colName.Name.String()
		case tree.SeqOptCache:
			// TODO: implement caching
			if *option.IntVal != 1 {
				return nil, fmt.Errorf("sequence caching for values other than 1 are not yet supported")
			}
		case tree.SeqOptIncrement:
			increment = *option.IntVal
			if incrementSet {
				return nil, fmt.Errorf("conflicting or redundant options")
			}
			if increment == 0 {
				return nil, fmt.Errorf("INCREMENT must not be zero")
			}
			incrementSet = true
		case tree.SeqOptMinValue:
			if option.IntVal != nil {
				minValue = *option.IntVal
				if minValueSet {
					return nil, fmt.Errorf("conflicting or redundant options")
				}
				minValueSet = true
			}
		case tree.SeqOptMaxValue:
			if option.IntVal != nil {
				maxValue = *option.IntVal
				if maxValueSet {
					return nil, fmt.Errorf("conflicting or redundant options")
				}
				maxValueSet = true
			}
		case tree.SeqOptStart:
			start = *option.IntVal
			if startSet {
				return nil, fmt.Errorf("conflicting or redundant options")
			}
			startSet = true
		default:
			return nil, fmt.Errorf("unknown CREATE SEQUENCE option")
		}
	}
	// Determine what all values should be based on what was set and what is inferred, as well as perform
	// validation for options that make sense
	if minValueSet {
		if minValue < minValueLimit || minValue > maxValueLimit {
			return nil, fmt.Errorf("MINVALUE (%d) is out of range for sequence data type %s", minValue, dataType.String())
		}
	} else if increment > 0 {
		minValue = 1
	} else {
		minValue = minValueLimit
	}
	if maxValueSet {
		if maxValue < minValueLimit || maxValue > maxValueLimit {
			return nil, fmt.Errorf("MAXVALUE (%d) is out of range for sequence data type %s", maxValue, dataType.String())
		}
	} else if increment > 0 {
		maxValue = maxValueLimit
	} else {
		maxValue = -1
	}
	if startSet {
		if start < minValue {
			return nil, fmt.Errorf("START value (%d) cannot be less than MINVALUE (%d))", start, minValue)
		}
		if start > maxValue {
			return nil, fmt.Errorf("START value (%d) cannot be greater than MAXVALUE (%d)", start, maxValue)
		}
	} else if increment > 0 {
		start = minValue
	} else {
		start = maxValue
	}
	if dataType.IsEmptyType() {
		dataType = pgtypes.Int64
	}
	// Returns the stored procedure call with all options
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateSequence(node.IfNotExists, name.SchemaQualifier.String(), &sequences.Sequence{
			Id:          id.NewSequence("", name.Name.String()),
			DataTypeID:  dataType.ID,
			Persistence: sequences.Persistence_Permanent,
			Start:       start,
			Current:     start,
			Increment:   increment,
			Minimum:     minValue,
			Maximum:     maxValue,
			Cache:       1,
			Cycle:       cycle,
			IsAtEnd:     false,
			OwnerTable:  id.NewTable("", ownerTableName),
			OwnerColumn: ownerColumnName,
		}),
		Children: nil,
	}, nil
}
