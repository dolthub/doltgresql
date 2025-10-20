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

package procedures

import (
	"context"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/core/id"
	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

const (
	FIELD_NAME_PARAMETER_NAMES  = "parameter_names"
	FIELD_NAME_PARAMETER_MODES  = "parameter_argmodes"
	FIELD_NAME_DEFINITION       = "definition"
	FIELD_NAME_EXTENSION_NAME   = "extension_name"
	FIELD_NAME_EXTENSION_SYMBOL = "extension_symbol"
	FIELD_NAME_SQL_DEFINITION   = "sql_definition"
)

// DeserializeRootObject implements the interface objinterface.Collection.
func (pgp *Collection) DeserializeRootObject(ctx context.Context, data []byte) (objinterface.RootObject, error) {
	return DeserializeProcedure(ctx, data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgp *Collection) DiffRootObjects(ctx context.Context, fromHash string, o objinterface.RootObject, t objinterface.RootObject, a objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	// We ignore many fields when diffing, as differences in these fields would result in a different procedure due to overloading
	// For example, "proc_name(text)" and "proc_name(varchar)" cannot produce a conflict as they're different procedures
	ours := o.(Procedure)
	theirs := t.(Procedure)
	ancestor, hasAncestor := a.(Procedure)
	var diffs []objinterface.RootObjectDiff
	{
		ourParamNames := strings.Join(ours.ParameterNames, ",")
		theirParamNames := strings.Join(theirs.ParameterNames, ",")
		ancParamNames := strings.Join(ancestor.ParameterNames, ",")
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_PARAMETER_NAMES,
		}
		if pgmerge.DiffValues(&diff, ourParamNames, theirParamNames, ancParamNames, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.ParameterNames = strings.Split(diff.OurValue.(string), ",")
		}
	}
	{
		ourModes := ours.ParameterModesAsString()
		theirModes := theirs.ParameterModesAsString()
		ancModes := ancestor.ParameterModesAsString()
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_PARAMETER_MODES,
		}
		if pgmerge.DiffValues(&diff, ourModes, theirModes, ancModes, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			paramModes, err := ParameterModesFromString(diff.OurValue.(string))
			if err != nil {
				return nil, nil, err
			}
			ours.ParameterModes = paramModes
		}
	}
	if ours.Definition != theirs.Definition {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_DEFINITION,
		}
		if pgmerge.DiffValues(&diff, ours.GetInnerDefinition(), theirs.GetInnerDefinition(), ancestor.GetInnerDefinition(), hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.Definition = ours.ReplaceDefinition(diff.OurValue.(string))
		}
	}
	if ours.ExtensionName != theirs.ExtensionName {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_EXTENSION_NAME,
		}
		if pgmerge.DiffValues(&diff, ours.ExtensionName, theirs.ExtensionName, ancestor.ExtensionName, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.ExtensionName = diff.OurValue.(string)
		}
	}
	if ours.ExtensionSymbol != theirs.ExtensionSymbol {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_EXTENSION_SYMBOL,
		}
		if pgmerge.DiffValues(&diff, ours.ExtensionSymbol, theirs.ExtensionSymbol, ancestor.ExtensionSymbol, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.ExtensionSymbol = diff.OurValue.(string)
		}
	}
	if ours.SQLDefinition != theirs.SQLDefinition {
		diff := objinterface.RootObjectDiff{
			Type:      pgtypes.Text,
			FromHash:  fromHash,
			FieldName: FIELD_NAME_SQL_DEFINITION,
		}
		if pgmerge.DiffValues(&diff, ours.SQLDefinition, theirs.SQLDefinition, ancestor.SQLDefinition, hasAncestor) {
			diffs = append(diffs, diff)
		} else {
			ours.SQLDefinition = diff.OurValue.(string)
		}
	}
	return diffs, ours, nil
}

// DropRootObject implements the interface objinterface.Collection.
func (pgp *Collection) DropRootObject(ctx context.Context, identifier id.Id) error {
	if identifier.Section() != id.Section_Procedure {
		return errors.Errorf(`procedure %s does not exist`, identifier.String())
	}
	return pgp.DropProcedure(ctx, id.Procedure(identifier))
}

// GetFieldType implements the interface objinterface.Collection.
func (pgp *Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	switch fieldName {
	case FIELD_NAME_PARAMETER_NAMES:
		return pgtypes.Text
	case FIELD_NAME_PARAMETER_MODES:
		return pgtypes.Text
	case FIELD_NAME_DEFINITION:
		return pgtypes.Text
	case FIELD_NAME_EXTENSION_NAME:
		return pgtypes.Text
	case FIELD_NAME_EXTENSION_SYMBOL:
		return pgtypes.Text
	case FIELD_NAME_SQL_DEFINITION:
		return pgtypes.Text
	default:
		return nil
	}
}

// GetID implements the interface objinterface.Collection.
func (pgp *Collection) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Procedures
}

// GetRootObject implements the interface objinterface.Collection.
func (pgp *Collection) GetRootObject(ctx context.Context, identifier id.Id) (objinterface.RootObject, bool, error) {
	if identifier.Section() != id.Section_Procedure {
		return nil, false, nil
	}
	f, err := pgp.GetProcedure(ctx, id.Procedure(identifier))
	return f, err == nil && f.ID.IsValid(), err
}

// HasRootObject implements the interface objinterface.Collection.
func (pgp *Collection) HasRootObject(ctx context.Context, identifier id.Id) (bool, error) {
	if identifier.Section() != id.Section_Procedure {
		return false, nil
	}
	return pgp.HasProcedure(ctx, id.Procedure(identifier)), nil
}

// IDToTableName implements the interface objinterface.Collection.
func (pgp *Collection) IDToTableName(identifier id.Id) doltdb.TableName {
	if identifier.Section() != id.Section_Procedure {
		return doltdb.TableName{}
	}
	return ProcedureIDToTableName(id.Procedure(identifier))
}

// IterAll implements the interface objinterface.Collection.
func (pgp *Collection) IterAll(ctx context.Context, callback func(rootObj objinterface.RootObject) (stop bool, err error)) error {
	return pgp.IterateProcedures(ctx, func(f Procedure) (stop bool, err error) {
		return callback(f)
	})
}

// IterIDs implements the interface objinterface.Collection.
func (pgp *Collection) IterIDs(ctx context.Context, callback func(identifier id.Id) (stop bool, err error)) error {
	return pgp.iterateIDs(ctx, func(procID id.Procedure) (stop bool, err error) {
		return callback(procID.AsId())
	})
}

// PutRootObject implements the interface objinterface.Collection.
func (pgp *Collection) PutRootObject(ctx context.Context, rootObj objinterface.RootObject) error {
	f, ok := rootObj.(Procedure)
	if !ok {
		return errors.Newf("invalid procedure root object: %T", rootObj)
	}
	return pgp.AddProcedure(ctx, f)
}

// RenameRootObject implements the interface objinterface.Collection.
func (pgp *Collection) RenameRootObject(ctx context.Context, oldName id.Id, newName id.Id) error {
	if !oldName.IsValid() || !newName.IsValid() || oldName.Section() != newName.Section() || oldName.Section() != id.Section_Procedure {
		return errors.New("cannot rename procedure due to invalid name")
	}
	oldProcName := id.Procedure(oldName)
	newProcName := id.Procedure(newName)
	if oldProcName.ParameterCount() != newProcName.ParameterCount() {
		return errors.Newf(`old procedure id had "%d" parameters, new procedure id has "%d" parameters`,
			oldProcName.ParameterCount(), newProcName.ParameterCount())
	}
	proc, err := pgp.GetProcedure(ctx, oldProcName)
	if err != nil {
		return err
	}
	if err = pgp.DropProcedure(ctx, oldProcName); err != nil {
		return err
	}
	proc.ID = newProcName
	return pgp.AddProcedure(ctx, proc)
}

// ResolveName implements the interface objinterface.Collection.
func (pgp *Collection) ResolveName(ctx context.Context, name doltdb.TableName) (doltdb.TableName, id.Id, error) {
	rawID, err := pgp.resolveName(ctx, name.Schema, name.Name)
	if err != nil || !rawID.IsValid() {
		return doltdb.TableName{}, id.Null, err
	}
	return ProcedureIDToTableName(rawID), rawID.AsId(), nil
}

// TableNameToID implements the interface objinterface.Collection.
func (pgp *Collection) TableNameToID(name doltdb.TableName) id.Id {
	return pgp.tableNameToID(name.Schema, name.Name).AsId()
}

// UpdateField implements the interface objinterface.Collection.
func (pgp *Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	procedure := rootObject.(Procedure)
	switch fieldName {
	case FIELD_NAME_PARAMETER_NAMES:
		procedure.ParameterNames = strings.Split(newValue.(string), ",")
	case FIELD_NAME_PARAMETER_MODES:
		newModes, err := ParameterModesFromString(newValue.(string))
		if err != nil {
			return nil, err
		}
		procedure.ParameterModes = newModes
	case FIELD_NAME_DEFINITION:
		newDefinition := procedure.ReplaceDefinition(newValue.(string))
		parsedBody, err := plpgsql.Parse(newDefinition)
		if err != nil {
			return nil, err
		}
		procedure.Definition = newDefinition
		procedure.Operations = parsedBody
	case FIELD_NAME_EXTENSION_NAME:
		procedure.ExtensionName = newValue.(string)
	case FIELD_NAME_EXTENSION_SYMBOL:
		procedure.ExtensionSymbol = newValue.(string)
	case FIELD_NAME_SQL_DEFINITION:
		procedure.SQLDefinition = newValue.(string)
	default:
		return nil, errors.Newf("unknown field name: `%s`", fieldName)
	}
	return procedure, nil
}
