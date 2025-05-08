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

package id

import "github.com/dolthub/go-mysql-server/sql"

// Operation represents an operation that is being performed or validated.
type Operation uint8

const (
	Operation_Rename Operation = iota
	Operation_Delete
	Operation_Delete_Cascade
)

// registry is the implementation of the global registry. This holds all functions that operate or validate a change on
// an ID.
type registry struct {
	listeners [][]Listener
}

// globalRegistry is the variable that is referenced for the registry.
var globalRegistry = &registry{
	listeners: make([][]Listener, section_count),
}

type Listener interface {
	// OperationPerformer is a function that performs the given operation on the original ID. Some operations, such as
	// renames, will use the new ID.
	OperationPerformer(ctx *sql.Context, operation Operation, databaseName string, originalID Id, newID Id) error
	// OperationValidator is a function that validates the given operation on the original ID. Some operations, such as
	// renames, will use the new ID. A validator is not required, and is intended for operations that may be relatively
	// expensive to perform, but able to check quickly for failures. In addition, validators should not perform any
	// modifications. If a validator is not required, then this should just return nil.
	OperationValidator(ctx *sql.Context, operation Operation, databaseName string, originalID Id, newID Id) error
}

// RegisterListener registers the given listener for the given sections.
//
// For example, sequences are related to tables. Whenever a table operation is performed that changes its ID, sequences
// will also need to update their IDs that reference the table. This is accomplished by registering a performer that
// accepts a table section, where the performer modifies sequences as needed.
//
// Performers should not register sections that are directly related to themselves. For example, a sequence performer
// should not register itself under the sequence section, as it will be the one broadcasting that section, and therefore
// could cause a loop.
func RegisterListener(listener Listener, sections ...Section) {
	for _, section := range sections {
		if section == Section_Null {
			continue
		}
		globalRegistry.listeners[section] = append(globalRegistry.listeners[section], listener)
	}
}

// PerformOperation calls all registered performers that are associated with the given section. This does not call any
// validators, which should be done using ValidateOperation. This returns the first error that is encountered.
func PerformOperation(ctx *sql.Context, targetSection Section, operation Operation, databaseName string, originalID Id, newID Id) error {
	for _, listener := range globalRegistry.listeners[targetSection] {
		if err := listener.OperationPerformer(ctx, operation, databaseName, originalID, newID); err != nil {
			return err
		}
	}
	// TODO: need to look for tables that store OIDs in their columns and UPDATE them to the new value
	//  it will be relatively slow, but that's the price a user pays to store OIDs in their tables
	return nil
}

// ValidateOperation calls all registered validators that are associated with the given section.
func ValidateOperation(ctx *sql.Context, targetSection Section, operation Operation, databaseName string, originalID Id, newID Id) error {
	for _, listener := range globalRegistry.listeners[targetSection] {
		if err := listener.OperationValidator(ctx, operation, databaseName, originalID, newID); err != nil {
			return err
		}
	}
	return nil
}

// String returns the name of the operation.
func (op Operation) String() string {
	switch op {
	case Operation_Rename:
		return "Rename"
	case Operation_Delete:
		return "Delete"
	case Operation_Delete_Cascade:
		return "DeleteCascade"
	default:
		return "UNKNOWN_OPERATION"
	}
}
