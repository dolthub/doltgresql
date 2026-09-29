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

package publications

import (
	"context"
	"encoding/json"

	"github.com/cockroachdb/errors"

	pgmerge "github.com/dolthub/doltgresql/core/merge"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

const fieldDefinition = "definition"

// definition returns the complete definition as text for root-object conflict inspection and resolution.
func (p Publication) definition() string {
	// All fields have JSON representations, so marshaling cannot fail.
	data, _ := json.Marshal(p)
	return string(data)
}

// DiffRootObjects implements the interface objinterface.Collection.
func (pgp *Collection) DiffRootObjects(ctx context.Context, fromHash string, ours, theirs, ancestor objinterface.RootObject) ([]objinterface.RootObjectDiff, objinterface.RootObject, error) {
	var base string
	if ancestor != nil {
		base = ancestor.(Publication).definition()
	}
	diff := objinterface.RootObjectDiff{Type: pgtypes.Text, FromHash: fromHash, FieldName: fieldDefinition}
	if pgmerge.DiffValues(&diff, ours.(Publication).definition(), theirs.(Publication).definition(), base, ancestor != nil) {
		return []objinterface.RootObjectDiff{diff}, ours, nil
	}
	merged, err := pgp.UpdateField(ctx, ours, fieldDefinition, diff.OurValue)
	return nil, merged, err
}

// GetFieldType implements the interface objinterface.Collection.
func (*Collection) GetFieldType(ctx context.Context, fieldName string) *pgtypes.DoltgresType {
	if fieldName == fieldDefinition {
		return pgtypes.Text
	}
	return nil
}

// UpdateField implements the interface objinterface.Collection.
func (*Collection) UpdateField(ctx context.Context, rootObject objinterface.RootObject, fieldName string, newValue any) (objinterface.RootObject, error) {
	if fieldName != fieldDefinition {
		return nil, errors.Errorf("unknown publication field: %s", fieldName)
	}
	text, ok := newValue.(string)
	if !ok {
		return nil, errors.New("publication definition must be text")
	}
	var p Publication
	if err := json.Unmarshal([]byte(text), &p); err != nil {
		return nil, err
	}
	if p.ID.AsId() != rootObject.GetID() {
		return nil, errors.New("publication conflict resolution cannot change the publication ID")
	}
	return p, nil
}
