// Copyright 2024 Dolthub, Inc.
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

package auth

import (
	"errors"
	"fmt"

	"github.com/dolthub/doltgresql/utils"
)

// PersistChanges will save the state of the global database to disk (assuming we are not using the pure in-memory
// implementation).
func PersistChanges() error {
	if fileSystem != nil {
		return fileSystem.WriteFile(authFileName, globalDatabase.serialize(), 0644)
	}
	return nil
}

// serialize returns the Database as a byte slice.
func (db *Database) serialize() []byte {
	writer := utils.NewWriter(16384)
	// Write the version
	writer.Uint32(0)
	// Write the roles
	writer.Uint32(uint32(len(db.rolesByID)))
	for _, role := range db.rolesByID {
		role.serialize(writer)
	}
	// Write the ownership
	db.ownership.serialize(writer)
	// Write the database privileges
	db.databasePrivileges.serialize(writer)
	// Write the schema privileges
	db.schemaPrivileges.serialize(writer)
	// Write the table privileges
	db.tablePrivileges.serialize(writer)
	// Write the role chain
	db.roleMembership.serialize(writer)
	return writer.Data()
}

// deserialize creates a Database from a byte slice.
func (db *Database) deserialize(data []byte) error {
	if len(data) < 4 {
		return errors.New("invalid auth database format")
	}
	reader := utils.NewReader(data)
	version := reader.Uint32()
	switch version {
	case 0:
		return db.deserializeV0(reader)
	default:
		return fmt.Errorf("Authorization database format %d is not supported, please upgrade Doltgres", version)
	}
}

// deserialize creates a Database from a byte slice. Expects a reader that has already read the version.
func (db *Database) deserializeV0(reader *utils.Reader) error {
	// Read the roles
	clear(db.rolesByName)
	clear(db.rolesByID)
	roleCount := reader.Uint32()
	for i := uint32(0); i < roleCount; i++ {
		r := Role{}
		r.deserialize(0, reader)
		db.rolesByName[r.Name] = r.id
		db.rolesByID[r.id] = r
	}
	// Read the ownership
	db.ownership.deserialize(0, reader)
	// Read the database privileges
	db.databasePrivileges.deserialize(0, reader)
	// Read the schema privileges
	db.schemaPrivileges.deserialize(0, reader)
	// Read the table privileges
	db.tablePrivileges.deserialize(0, reader)
	// Read the role chain
	db.roleMembership.deserialize(0, reader)
	return nil
}
