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

package tree

// Vacuum represents a VACUUM statement.
type Vacuum struct {
	Options VacuumOptions
	TablesAndCols VacuumTableAndColsList
}

var _ Statement = &Vacuum{}

func (node *Vacuum) String() string {
	return "Vacuum" // TODO
}
 
func (node *Vacuum) StatementType() StatementType {
	return Ack
}

func (node *Vacuum) StatementTag() string {
	return "VACUUM"
}

// Format implements the NodeFormatter interface.
func (node *Vacuum) Format(ctx *FmtCtx) {
	ctx.WriteString("VACUUM") // TODO
}

type VacuumOption struct {
	Option string
	Value any
}

type VacuumOptions []*VacuumOption

type VacuumTableAndCols struct {
	Name *UnresolvedName
	Cols NameList
}

type VacuumTableAndColsList []*VacuumTableAndCols