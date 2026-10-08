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

package aggregates

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
)

// Collection contains a collection of aggregate functions.
type Collection struct {
	objinterface.RootObjectMap
}

// Aggregate represents a created aggregate function.
type Aggregate struct {
	ID          id.Function
	ReturnType  id.Type
	SFunc       id.Function // State transition function
	SType       id.Type     // Internal state type
	FinalFunc   id.Function // Final calculation function
	CombineFunc id.Function
	InitCond    string
	HasInitCond bool
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Aggregate{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, rom objinterface.RootObjectMap) *Collection {
	return &Collection{RootObjectMap: rom}
}

// GetAggregate returns the aggregate with the given ID. Returns an Aggregate with an invalid ID if it cannot be found
// (Aggregate.ID.IsValid() == false).
func (pga *Collection) GetAggregate(ctx context.Context, aggregateID id.Function) (Aggregate, error) {
	h, err := pga.Contents().Get(ctx, string(aggregateID))
	if err != nil || h.IsEmpty() {
		return Aggregate{}, err
	}
	data, err := pga.NodeStore().ReadBytes(ctx, h)
	if err != nil {
		return Aggregate{}, err
	}
	return DeserializeAggregate(ctx, data)
}

// GetAggregateOverloads returns every aggregate that shares the given aggregate's schema and name.
func (pga *Collection) GetAggregateOverloads(ctx context.Context, aggregateID id.Function) ([]Aggregate, error) {
	var overloads []Aggregate
	err := pga.IterateAggregates(ctx, func(a Aggregate) (stop bool, err error) {
		if a.ID.SchemaName() == aggregateID.SchemaName() && a.ID.FunctionName() == aggregateID.FunctionName() {
			overloads = append(overloads, a)
		}
		return false, nil
	})
	return overloads, err
}

// HasAggregate returns whether the given aggregate exists.
func (pga *Collection) HasAggregate(ctx context.Context, aggregateID id.Function) bool {
	ok, err := pga.Contents().Has(ctx, string(aggregateID))
	return err == nil && ok
}

// HasAggregateName returns whether an aggregate with the given name exists in any schema.
func (pga *Collection) HasAggregateName(ctx context.Context, name string) (bool, error) {
	found := false
	err := pga.Contents().IterAll(ctx, func(k string, _ hash.Hash) error {
		if id.Function(k).FunctionName() == name {
			found = true
			return io.EOF
		}
		return nil
	})
	if err != nil && err != io.EOF {
		return false, err
	}
	return found, nil
}

// AddAggregate adds a new aggregate.
func (pga *Collection) AddAggregate(ctx context.Context, a Aggregate) error {
	if pga.HasAggregate(ctx, a.ID) {
		return errors.Errorf(`aggregate "%s" already exists with same argument types`, a.ID.FunctionName())
	}
	data, err := a.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pga.NodeStore().WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pga.Contents().Editor()
	if err = mapEditor.Add(ctx, string(a.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pga.SetContents(newMap)
	return nil
}

// DropAggregate drops an existing aggregate.
func (pga *Collection) DropAggregate(ctx context.Context, aggregateIDs ...id.Function) error {
	if len(aggregateIDs) == 0 {
		return nil
	}
	for _, aggregateID := range aggregateIDs {
		if ok, err := pga.Contents().Has(ctx, string(aggregateID)); err != nil {
			return err
		} else if !ok {
			return errors.Errorf(`aggregate %s does not exist`, aggregateID.DisplayString())
		}
	}

	mapEditor := pga.Contents().Editor()
	for _, aggregateID := range aggregateIDs {
		if err := mapEditor.Delete(ctx, string(aggregateID)); err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pga.SetContents(newMap)
	return nil
}

// resolveName returns the fully resolved name of the given aggregate. Returns an error if the name is ambiguous.
func (pga *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Function, error) {
	partialID := functions.TableNameToFunctionID(schemaName, formattedName)
	if !partialID.IsValid() {
		return id.NullFunction, nil
	}

	// Check for an exact match
	if pga.HasAggregate(ctx, partialID) {
		return partialID, nil
	}

	// Otherwise we'll iterate over all the names
	var resolvedID id.Function
	partialParams := partialID.Parameters()
	err := pga.IterateAggregates(ctx, func(a Aggregate) (stop bool, err error) {
		if !strings.EqualFold(a.ID.FunctionName(), partialID.FunctionName()) {
			return false, nil
		}
		if len(partialID.SchemaName()) > 0 && !strings.EqualFold(a.ID.SchemaName(), partialID.SchemaName()) {
			return false, nil
		}
		if len(partialParams) > 0 {
			if a.ID.ParameterCount() != len(partialParams) {
				return false, nil
			}
			for i, param := range a.ID.Parameters() {
				if len(partialParams[i].TypeName()) > 0 && !strings.EqualFold(partialParams[i].TypeName(), param.TypeName()) {
					return false, nil
				}
				if len(partialParams[i].SchemaName()) > 0 && !strings.EqualFold(partialParams[i].SchemaName(), param.SchemaName()) {
					return false, nil
				}
			}
		}
		// Everything must have matched to have made it here
		if resolvedID.IsValid() {
			aggregateTableName := functions.FunctionIDToTableName(a.ID)
			resolvedTableName := functions.FunctionIDToTableName(resolvedID)
			return true, fmt.Errorf("`%s` is ambiguous, matches `%s` and `%s`",
				formattedName, aggregateTableName.String(), resolvedTableName.String())
		}
		resolvedID = a.ID
		return false, nil
	})
	return resolvedID, err
}

// IterateAggregates iterates over all aggregates in the collection.
func (pga *Collection) IterateAggregates(ctx context.Context, callback func(a Aggregate) (stop bool, err error)) error {
	return pga.Contents().IterAll(ctx, func(_ string, v hash.Hash) error {
		data, err := pga.NodeStore().ReadBytes(ctx, v)
		if err != nil {
			return err
		}
		a, err := DeserializeAggregate(ctx, data)
		if err != nil {
			return err
		}
		stop, err := callback(a)
		if err != nil {
			return err
		} else if stop {
			return io.EOF
		} else {
			return nil
		}
	})
}

// GetID implements the interface objinterface.RootObject.
func (aggregate Aggregate) GetID() id.Id {
	return aggregate.ID.AsId()
}

// GetRootObjectID implements the interface objinterface.RootObject.
func (aggregate Aggregate) GetRootObjectID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Aggregates
}

// HashOf implements the interface objinterface.RootObject.
func (aggregate Aggregate) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := aggregate.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (aggregate Aggregate) Name() doltdb.TableName {
	return functions.FunctionIDToTableName(aggregate.ID)
}
