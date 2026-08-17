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

package operators

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Collection contains a collection of operators.
type Collection struct {
	objinterface.RootObjectMap
}

// Operator represents a user-defined operator.
type Operator struct {
	ID         id.Operator
	Function   id.Function
	ReturnType id.Type
	Commutator string
	Negator    string
	Hashes     bool
	Merges     bool
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Operator{}

// NewCollection returns a new Collection.
func NewCollection(rom objinterface.RootObjectMap) *Collection {
	return &Collection{RootObjectMap: rom}
}

// GetOperator returns the operator with the given ID. Returns an Operator with an invalid ID if it cannot be found
// (Operator.ID.IsValid() == false).
func (pgo *Collection) GetOperator(ctx context.Context, operatorID id.Operator) (Operator, error) {
	h, err := pgo.Contents().Get(ctx, string(operatorID))
	if err != nil || h.IsEmpty() {
		return Operator{}, err
	}
	data, err := pgo.NodeStore().ReadBytes(ctx, h)
	if err != nil {
		return Operator{}, err
	}
	return DeserializeOperator(ctx, data)
}

// ResolveOperator returns the operator matching the given symbol and operand types, searching the given schemas in
// order and allowing unknown operand types to match any type. Returns false if no operator matches.
func (pgo *Collection) ResolveOperator(ctx context.Context, schemaNames []string, symbol string, leftType id.Type, rightType id.Type) (Operator, bool, error) {
	for _, schemaName := range schemaNames {
		o, err := pgo.GetOperator(ctx, id.NewOperator(schemaName, symbol, leftType, rightType))
		if err != nil {
			return Operator{}, false, err
		}
		if o.ID.IsValid() {
			return o, true, nil
		}
	}
	leftUnknown := leftType == pgtypes.Unknown.ID
	rightUnknown := rightType == pgtypes.Unknown.ID
	if !leftUnknown && !rightUnknown {
		return Operator{}, false, nil
	}
	for _, schemaName := range schemaNames {
		var resolved Operator
		err := pgo.IterateOperators(ctx, func(o Operator) (stop bool, err error) {
			if o.ID.SchemaName() != schemaName || o.ID.Symbol() != symbol {
				return false, nil
			}
			if (!leftUnknown && o.ID.LeftType() != leftType) || (!rightUnknown && o.ID.RightType() != rightType) {
				return false, nil
			}
			if resolved.ID.IsValid() {
				return true, errors.Errorf("operator is not unique: %s", symbol)
			}
			resolved = o
			return false, nil
		})
		if err != nil {
			return Operator{}, false, err
		}
		if resolved.ID.IsValid() {
			return resolved, true, nil
		}
	}
	return Operator{}, false, nil
}

// HasOperator returns whether the given operator exists.
func (pgo *Collection) HasOperator(ctx context.Context, operatorID id.Operator) bool {
	ok, err := pgo.Contents().Has(ctx, string(operatorID))
	return err == nil && ok
}

// AddOperator adds a new operator.
func (pgo *Collection) AddOperator(ctx context.Context, o Operator) error {
	if pgo.HasOperator(ctx, o.ID) {
		return errors.Errorf(`operator %s already exists`, o.ID.Symbol())
	}
	data, err := o.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgo.NodeStore().WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgo.Contents().Editor()
	if err = mapEditor.Add(ctx, string(o.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgo.SetContents(newMap)
	return nil
}

// DropOperator drops an existing operator.
func (pgo *Collection) DropOperator(ctx context.Context, operatorIDs ...id.Operator) error {
	if len(operatorIDs) == 0 {
		return nil
	}
	for _, operatorID := range operatorIDs {
		if ok, err := pgo.Contents().Has(ctx, string(operatorID)); err != nil {
			return err
		} else if !ok {
			return errors.Errorf(`operator does not exist: %s %s %s`,
				operatorID.LeftType().TypeName(), operatorID.Symbol(), operatorID.RightType().TypeName())
		}
	}

	mapEditor := pgo.Contents().Editor()
	for _, operatorID := range operatorIDs {
		if err := mapEditor.Delete(ctx, string(operatorID)); err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgo.SetContents(newMap)
	return nil
}

// resolveName returns the fully resolved name of the given operator. Returns an error if the name is ambiguous.
func (pgo *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Operator, error) {
	if len(formattedName) == 0 {
		return id.NullOperator, nil
	}

	// Check for an exact match
	fullID := pgo.tableNameToID(schemaName, formattedName)
	if pgo.HasOperator(ctx, fullID) {
		return fullID, nil
	}

	// Otherwise we'll iterate over all the names
	var resolvedID id.Operator
	err := pgo.IterateOperators(ctx, func(o Operator) (stop bool, err error) {
		if !strings.EqualFold(string(o.ID), string(fullID)) {
			return false, nil
		}
		// The above matches, so this counts as a match
		if resolvedID.IsValid() {
			operatorTableName := OperatorIDToTableName(o.ID)
			resolvedTableName := OperatorIDToTableName(resolvedID)
			return true, fmt.Errorf("`%s` is ambiguous, matches `%s` and `%s`",
				formattedName, operatorTableName.String(), resolvedTableName.String())
		}
		resolvedID = o.ID
		return false, nil
	})
	return resolvedID, err
}

// IterateOperators iterates over all operators in the collection.
func (pgo *Collection) IterateOperators(ctx context.Context, callback func(o Operator) (stop bool, err error)) error {
	return pgo.Contents().IterAll(ctx, func(_ string, v hash.Hash) error {
		data, err := pgo.NodeStore().ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		o, err := DeserializeOperator(ctx, data)
		if err != nil {
			return err
		}
		stop, err := callback(o)
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
}

// tableNameToID returns the ID that was encoded via the Name() call, as the returned TableName contains additional
// information (which this is able to process).
func (pgo *Collection) tableNameToID(schema string, formattedName string) id.Operator {
	sections := strings.Split(strings.TrimSuffix(strings.TrimPrefix(formattedName, "("), ")"), ")|(")
	if len(sections) != 5 {
		return id.NullOperator
	}
	return id.NewOperator(schema, sections[0],
		id.NewType(sections[1], sections[2]), id.NewType(sections[3], sections[4]))
}

// GetID implements the interface objinterface.RootObject.
func (operator Operator) GetID() id.Id {
	return operator.ID.AsId()
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (operator Operator) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Operators
}

// HashOf implements the interface objinterface.RootObject.
func (operator Operator) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := operator.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (operator Operator) Name() doltdb.TableName {
	return OperatorIDToTableName(operator.ID)
}

// OperatorIDToTableName returns the ID in a format that's better for user consumption.
func OperatorIDToTableName(operatorID id.Operator) doltdb.TableName {
	name := fmt.Sprintf(`(%s)|(%s)|(%s)|(%s)|(%s)`,
		operatorID.Symbol(),
		operatorID.LeftType().SchemaName(),
		operatorID.LeftType().TypeName(),
		operatorID.RightType().SchemaName(),
		operatorID.RightType().TypeName())
	return doltdb.TableName{
		Name:   name,
		Schema: operatorID.SchemaName(),
	}
}
