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
	"context"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

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

// RootObjectDiffChange specifies the type of change that occurred from the ancestor value.
type RootObjectDiffChange uint8

const (
	RootObjectDiffChange_Added RootObjectDiffChange = iota
	RootObjectDiffChange_Deleted
	RootObjectDiffChange_Modified
)

// RootObjectDiff represents a diff between the ancestor value and our/their values. The field name uniquely identifies
// which part of a root object that this diff covers.
type RootObjectDiff struct {
	Type          *pgtypes.DoltgresType
	FieldName     string
	AncestorValue any
	OurValue      any
	TheirValue    any
	OurChange     RootObjectDiffChange
	TheirChange   RootObjectDiffChange
}
