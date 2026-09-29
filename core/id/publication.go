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

package id

// Publication is an Id wrapper for database-wide publications. This wrapper must not be returned to the client.
type Publication Id

// NullPublication is an empty, invalid publication ID.
const NullPublication Publication = ""

// NewPublication returns a publication ID. Publication names are not schema-qualified.
func NewPublication(name string) Publication {
	if name == "" {
		return NullPublication
	}
	return Publication(NewId(Section_Publication, name))
}

// Name returns the publication's name.
func (p Publication) Name() string { return Id(p).Segment(0) }

// IsValid returns whether the ID is valid.
func (p Publication) IsValid() bool { return Id(p).IsValid() }

// AsId returns the underlying ID.
func (p Publication) AsId() Id { return Id(p) }
