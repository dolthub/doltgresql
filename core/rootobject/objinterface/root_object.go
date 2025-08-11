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

package objinterface

import (
	"cmp"
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// RootObject is an expanded interface on Dolt's root objects.
type RootObject interface {
	doltdb.RootObject
	// GetID returns the root object ID.
	GetID() id.Id
	// GetRootObjectID returns the root object ID.
	GetRootObjectID() RootObjectID
	// Serialize returns the byte representation of the root object.
	Serialize(ctx context.Context) ([]byte, error)
}

// Conflict is an expanded interface on Dolt's conflict root object.
type Conflict interface {
	RootObject
	doltdb.ConflictRootObject
	// GetContainedRootObjectID returns the root object ID of the contained items.
	GetContainedRootObjectID() RootObjectID
	// Diffs returns the diffs for the conflict, along with the merged root object if there are no diffs.
	Diffs(ctx context.Context) ([]RootObjectDiff, RootObject, error)
	// FieldType returns the type associated with the given field name. Returns nil if the name does not match a field.
	FieldType(ctx context.Context, name string) *pgtypes.DoltgresType
}

// RootObjectDiffChange specifies the type of change that occurred from the ancestor value.
type RootObjectDiffChange uint8

const (
	RootObjectDiffChange_Added RootObjectDiffChange = iota
	RootObjectDiffChange_Deleted
	RootObjectDiffChange_Modified
	RootObjectDiffChange_NoChange
)

// RootObjectDiffSchema is the baseline schema that is returned for root object diffs.
var RootObjectDiffSchema = sql.Schema{
	{Name: "from_root_ish", Type: pgtypes.Text, Default: nil, Nullable: false},
	{Name: "base_value", Type: pgtypes.Text, Default: nil, Nullable: true},
	{Name: "our_value", Type: pgtypes.Text, Default: nil, Nullable: true},
	{Name: "our_diff_type", Type: pgtypes.Text, Default: nil, Nullable: false},
	{Name: "their_value", Type: pgtypes.Text, Default: nil, Nullable: true},
	{Name: "their_diff_type", Type: pgtypes.Text, Default: nil, Nullable: false},
	{Name: "dolt_conflict_id", Type: pgtypes.Text, Default: nil, Nullable: false},
}

// RootObjectDiff represents a diff between the ancestor value and our/their values. The field name uniquely identifies
// which part of a root object that this diff covers.
type RootObjectDiff struct {
	Type          *pgtypes.DoltgresType
	FromHash      string
	FieldName     string
	AncestorValue any
	OurValue      any
	TheirValue    any
	OurChange     RootObjectDiffChange
	TheirChange   RootObjectDiffChange
}

var _ doltdb.RootObjectDiff = RootObjectDiff{}

// CompareIds implements the interface doltdb.RootObjectDiff.
func (diff RootObjectDiff) CompareIds(ctx context.Context, o doltdb.RootObjectDiff) (int, error) {
	other, ok := o.(RootObjectDiff)
	if !ok {
		return 0, errors.Errorf("root object diff cannot compare with diff of type %T", o)
	}
	return cmp.Compare(diff.FieldName, other.FieldName), nil
}

// ToRow implements the interface doltdb.RootObjectDiff.
func (diff RootObjectDiff) ToRow(ctx *sql.Context) (_ sql.Row, err error) {
	var baseValue any
	var ourValue any
	var theirValue any
	var ourChange any
	var theirChange any
	if diff.AncestorValue != nil {
		baseValue, err = diff.Type.IoOutput(ctx, diff.AncestorValue)
		if err != nil {
			return nil, err
		}
	}
	if diff.OurValue != nil {
		ourValue, err = diff.Type.IoOutput(ctx, diff.OurValue)
		if err != nil {
			return nil, err
		}
	}
	if diff.TheirValue != nil {
		theirValue, err = diff.Type.IoOutput(ctx, diff.TheirValue)
		if err != nil {
			return nil, err
		}
	}
	switch diff.OurChange {
	case RootObjectDiffChange_Added:
		ourChange = "added"
	case RootObjectDiffChange_Deleted:
		ourChange = "deleted"
	case RootObjectDiffChange_Modified:
		ourChange = "modified"
	case RootObjectDiffChange_NoChange:
		ourChange = "no_change"
	}
	switch diff.TheirChange {
	case RootObjectDiffChange_Added:
		theirChange = "added"
	case RootObjectDiffChange_Deleted:
		theirChange = "deleted"
	case RootObjectDiffChange_Modified:
		theirChange = "modified"
	case RootObjectDiffChange_NoChange:
		ourChange = "no_change"
	}
	return sql.Row{diff.FromHash, baseValue, ourValue, ourChange, theirValue, theirChange, diff.FieldName}, nil
}

// DiffFromRow converts a row to a conflict diff.
func DiffFromRow(ctx *sql.Context, conflict doltdb.ConflictRootObject, row sql.Row) (_ doltdb.RootObjectDiff, err error) {
	if len(row) != len(RootObjectDiffSchema) {
		return nil, errors.Newf("expected root object row diff to have %d columns but had %d", len(RootObjectDiffSchema), len(row))
	}
	fieldName := row[6].(string)
	typ := conflict.(Conflict).FieldType(ctx, fieldName)
	if typ == nil {
		return nil, errors.Newf("cannot find a field named `%s`", fieldName)
	}
	var baseValue any
	var ourValue any
	var theirValue any
	var ourChange RootObjectDiffChange
	var theirChange RootObjectDiffChange
	if row[1] != nil {
		baseValue, err = typ.IoInput(ctx, row[1].(string))
		if err != nil {
			return nil, err
		}
	}
	if row[2] != nil {
		ourValue, err = typ.IoInput(ctx, row[2].(string))
		if err != nil {
			return nil, err
		}
	}
	if row[4] != nil {
		theirValue, err = typ.IoInput(ctx, row[4].(string))
		if err != nil {
			return nil, err
		}
	}
	switch row[3].(string) {
	case "added":
		ourChange = RootObjectDiffChange_Added
	case "deleted":
		ourChange = RootObjectDiffChange_Deleted
	case "modified":
		ourChange = RootObjectDiffChange_Modified
	case "no_change":
		ourChange = RootObjectDiffChange_NoChange
	}
	switch row[5].(string) {
	case "added":
		theirChange = RootObjectDiffChange_Added
	case "deleted":
		theirChange = RootObjectDiffChange_Deleted
	case "modified":
		theirChange = RootObjectDiffChange_Modified
	case "no_change":
		ourChange = RootObjectDiffChange_NoChange
	}
	return RootObjectDiff{
		Type:          typ,
		FromHash:      row[0].(string),
		FieldName:     fieldName,
		AncestorValue: baseValue,
		OurValue:      ourValue,
		TheirValue:    theirValue,
		OurChange:     ourChange,
		TheirChange:   theirChange,
	}, nil
}
