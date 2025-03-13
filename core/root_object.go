// Copyright 2025 Dolthub, Inc.
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

package core

import (
	"context"
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/types"

	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/sequences"
	"github.com/dolthub/doltgresql/core/typecollection"
)

// RootObjectMap is a map that all root object collections make use of.
type RootObjectMap struct {
	m   types.Map
	vrw types.ValueReadWriter
}

// RootObjectID is an ID that distinguishes names and root objects from one another.
type RootObjectID int64

const (
	RootObjectID_None      RootObjectID = 0
	RootObjectID_Sequences RootObjectID = 1
	RootObjectID_Types     RootObjectID = 2
	RootObjectID_Functions RootObjectID = 3
)

// NewRootObjectMap returns a new RootObjectMap.
func NewRootObjectMap(ctx context.Context, vrw types.ValueReadWriter) (RootObjectMap, error) {
	m, err := types.NewMap(ctx, vrw)
	return ToRootObjectMap(ctx, m, vrw), err
}

// ToRootObjectMap creates a RootObjectMap from the given map and vrw.
func ToRootObjectMap(ctx context.Context, m types.Map, vrw types.ValueReadWriter) RootObjectMap {
	return RootObjectMap{
		m:   m,
		vrw: vrw,
	}
}

// ResolveName returns the fully resolved name of the given item (if the item exists). Also returns the type of the item.
func (rom RootObjectMap) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, RootObjectID, error) {
	var resolvedName doltdb.TableName
	resolvedRawID := id.Null
	resolvedObjID := RootObjectID_None

	// Check sequence names
	val, ok, err := rom.m.MaybeGet(ctx, types.Int(RootObjectID_Sequences))
	if err != nil {
		return doltdb.TableName{}, id.Null, RootObjectID_None, err
	}
	if ok {
		seqs := sequences.CollectionFromMap(val.(types.Map))
		resolvedID, err := seqs.ResolveName(ctx, name.Schema, name.Name)
		if err != nil {
			return doltdb.TableName{}, id.Null, RootObjectID_None, err
		}
		if resolvedID.IsValid() {
			resolvedName = doltdb.TableName{
				Name:   resolvedID.SequenceName(),
				Schema: resolvedID.SchemaName(),
			}
			resolvedRawID = resolvedID.AsId()
			resolvedObjID = RootObjectID_Sequences
		}
	}
	// Check function names
	val, ok, err = rom.m.MaybeGet(ctx, types.Int(RootObjectID_Functions))
	if err != nil {
		return doltdb.TableName{}, id.Null, RootObjectID_None, err
	}
	if ok {
		funcs := functions.CollectionFromMap(val.(types.Map), rom.vrw)
		resolvedID, err := funcs.ResolveName(ctx, name.Schema, name.Name)
		if err != nil {
			return doltdb.TableName{}, id.Null, RootObjectID_None, err
		}
		if resolvedID.IsValid() {
			if resolvedObjID != RootObjectID_None {
				return doltdb.TableName{}, id.Null, RootObjectID_None, fmt.Errorf(`"%s" is ambiguous`, name.String())
			}
			resolvedName = functions.FunctionIDToTableName(resolvedID)
			resolvedRawID = resolvedID.AsId()
			resolvedObjID = RootObjectID_Functions
		}
	}
	// Check type names
	val, ok, err = rom.m.MaybeGet(ctx, types.Int(RootObjectID_Types))
	if err != nil {
		return doltdb.TableName{}, id.Null, RootObjectID_None, err
	}
	{
		var typeColl *typecollection.TypeCollection
		if ok {
			typeColl = typecollection.CollectionFromMap(val.(types.Map))
		} else {
			emptyMap, err := types.NewMap(ctx, rom.vrw)
			if err != nil {
				return doltdb.TableName{}, id.Null, RootObjectID_None, err
			}
			typeColl = typecollection.CollectionFromMap(emptyMap)
		}
		resolvedID, err := typeColl.ResolveName(ctx, name.Schema, name.Name)
		if err != nil {
			return doltdb.TableName{}, id.Null, RootObjectID_None, err
		}
		if resolvedID.IsValid() {
			if resolvedObjID != RootObjectID_None {
				return doltdb.TableName{}, id.Null, RootObjectID_None, fmt.Errorf(`"%s" is ambiguous`, name.String())
			}
			resolvedName = doltdb.TableName{
				Name:   resolvedID.TypeName(),
				Schema: resolvedID.SchemaName(),
			}
			resolvedRawID = resolvedID.AsId()
			resolvedObjID = RootObjectID_Types
		}
	}

	return resolvedName, resolvedRawID, resolvedObjID, nil
}

// GetSequences returns the sequence collection.
func (rom RootObjectMap) GetSequences(ctx context.Context) (*sequences.Collection, error) {
	val, ok, err := rom.m.MaybeGet(ctx, types.Int(RootObjectID_Sequences))
	if err != nil {
		return nil, err
	}
	if !ok {
		newMap, err := types.NewMap(ctx, rom.vrw)
		return sequences.CollectionFromMap(newMap), err
	}
	return sequences.CollectionFromMap(val.(types.Map)), nil
}

// SetSequences updates the sequence collection.
func (rom RootObjectMap) SetSequences(ctx context.Context, coll *sequences.Collection) (newRom RootObjectMap, err error) {
	collMap, err := coll.Map(ctx)
	if err != nil {
		return rom, err
	}
	mapEditor := rom.m.Edit()
	defer func() {
		nErr := mapEditor.Close(ctx)
		if err == nil {
			err = nErr
		}
	}()
	newRomMap, err := mapEditor.Set(types.Int(RootObjectID_Sequences), collMap).Map(ctx)
	if err != nil {
		return rom, err
	}
	return ToRootObjectMap(ctx, newRomMap, rom.vrw), nil
}

// GetFunctions returns the function collection.
func (rom RootObjectMap) GetFunctions(ctx context.Context) (*functions.Collection, error) {
	val, ok, err := rom.m.MaybeGet(ctx, types.Int(RootObjectID_Functions))
	if err != nil {
		return nil, err
	}
	if !ok {
		newMap, err := types.NewMap(ctx, rom.vrw)
		return functions.CollectionFromMap(newMap, rom.vrw), err
	}
	return functions.CollectionFromMap(val.(types.Map), rom.vrw), nil
}

// SetFunctions updates the function collection.
func (rom RootObjectMap) SetFunctions(ctx context.Context, coll *functions.Collection) (newRom RootObjectMap, err error) {
	collMap, err := coll.Map(ctx)
	if err != nil {
		return rom, err
	}
	mapEditor := rom.m.Edit()
	defer func() {
		nErr := mapEditor.Close(ctx)
		if err == nil {
			err = nErr
		}
	}()
	newRomMap, err := mapEditor.Set(types.Int(RootObjectID_Functions), collMap).Map(ctx)
	if err != nil {
		return rom, err
	}
	return ToRootObjectMap(ctx, newRomMap, rom.vrw), nil
}

// GetTypes returns the types collection.
func (rom RootObjectMap) GetTypes(ctx context.Context) (*typecollection.TypeCollection, error) {
	val, ok, err := rom.m.MaybeGet(ctx, types.Int(RootObjectID_Types))
	if err != nil {
		return nil, err
	}
	if !ok {
		newMap, err := types.NewMap(ctx, rom.vrw)
		return typecollection.CollectionFromMap(newMap), err
	}
	return typecollection.CollectionFromMap(val.(types.Map)), nil
}

// SetTypes updates the types collection.
func (rom RootObjectMap) SetTypes(ctx context.Context, coll *typecollection.TypeCollection) (newRom RootObjectMap, err error) {
	collMap, err := coll.Map(ctx)
	if err != nil {
		return rom, err
	}
	mapEditor := rom.m.Edit()
	defer func() {
		nErr := mapEditor.Close(ctx)
		if err == nil {
			err = nErr
		}
	}()
	newRomMap, err := mapEditor.Set(types.Int(RootObjectID_Types), collMap).Map(ctx)
	if err != nil {
		return rom, err
	}
	return ToRootObjectMap(ctx, newRomMap, rom.vrw), nil
}
