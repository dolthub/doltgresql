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

package auth

import (
	"sort"
	"strings"

	"github.com/dolthub/doltgresql/utils"
)

// RoleSettings contains the configuration parameter defaults set through ALTER ROLE ... SET and ALTER DATABASE ... SET.
// This mirrors Postgres' pg_db_role_setting catalog.
type RoleSettings struct {
	Data map[RoleSettingKey][]RoleSetting
}

// RoleSettingKey identifies the target of a group of settings. An invalid Role applies the settings to every role, and
// an empty Database applies the settings to every database. ALTER DATABASE d SET is therefore stored identically to
// ALTER ROLE ALL IN DATABASE d SET, as is done in Postgres.
type RoleSettingKey struct {
	Role     RoleID
	Database string
}

// RoleSetting is a single configuration parameter and its value, stored as the text that will be applied when a
// session starts.
type RoleSetting struct {
	Name  string
	Value string
}

// NewRoleSettings returns a new *RoleSettings.
func NewRoleSettings() *RoleSettings {
	return &RoleSettings{Data: make(map[RoleSettingKey][]RoleSetting)}
}

// SetRoleSetting sets the parameter for the given key. Parameter names are case-insensitive, and a parameter that is
// already set keeps its position, which matches the ordering of setconfig in Postgres. Callers must hold the write lock.
func SetRoleSetting(key RoleSettingKey, name string, value string) {
	settings := globalDatabase.roleSettings.Data[key]
	for i := range settings {
		if strings.EqualFold(settings[i].Name, name) {
			settings[i].Name = name
			settings[i].Value = value
			return
		}
	}
	globalDatabase.roleSettings.Data[key] = append(settings, RoleSetting{Name: name, Value: value})
}

// ResetRoleSetting removes the parameter from the given key. This is a no-op if the parameter is not set. Callers must
// hold the write lock.
func ResetRoleSetting(key RoleSettingKey, name string) {
	settings := globalDatabase.roleSettings.Data[key]
	for i := range settings {
		if strings.EqualFold(settings[i].Name, name) {
			settings = append(settings[:i], settings[i+1:]...)
			break
		}
	}
	if len(settings) == 0 {
		delete(globalDatabase.roleSettings.Data, key)
	} else {
		globalDatabase.roleSettings.Data[key] = settings
	}
}

// ResetAllRoleSettings removes every parameter from the given key. Callers must hold the write lock.
func ResetAllRoleSettings(key RoleSettingKey) {
	delete(globalDatabase.roleSettings.Data, key)
}

// RemoveDatabaseRoleSettings removes every setting that targets the given database, which is used when the database is
// dropped. Database names are case-insensitive, so this matches the name case-insensitively. Callers must hold the write
// lock.
func RemoveDatabaseRoleSettings(database string) {
	for key := range globalDatabase.roleSettings.Data {
		if len(key.Database) > 0 && strings.EqualFold(key.Database, database) {
			delete(globalDatabase.roleSettings.Data, key)
		}
	}
}

// RoleSettingsForKey returns a copy of the settings for the given key. Callers must hold the read or write lock.
func RoleSettingsForKey(key RoleSettingKey) []RoleSetting {
	return append([]RoleSetting(nil), globalDatabase.roleSettings.Data[key]...)
}

// SessionRoleSettings returns the settings that apply to a session for the given role connected to the given database.
// Settings are returned from least to most specific, so applying them in order gives precedence to the more specific
// ones: every role and database, the database, the role, and finally the role within the database. This handles
// locking internally.
func SessionRoleSettings(roleName string, database string) []RoleSetting {
	var settings []RoleSetting
	LockRead(func() {
		roleID := globalDatabase.rolesByName[roleName]
		keys := []RoleSettingKey{{}, {Database: database}, {Role: roleID}, {Role: roleID, Database: database}}
		for i, key := range keys {
			// Skip keys that are duplicates of an earlier key, which happens for a missing role or database
			if (len(database) == 0 && i%2 == 1) || (!roleID.IsValid() && i >= 2) {
				continue
			}
			settings = append(settings, globalDatabase.roleSettings.Data[key]...)
		}
	})
	return settings
}

// RoleSettingsEntry is a single row of pg_db_role_setting.
type RoleSettingsEntry struct {
	Key      RoleSettingKey
	RoleName string
	Settings []RoleSetting
}

// AllRoleSettings returns every group of settings, sorted by database and then role name. Callers must hold the read or
// write lock.
func AllRoleSettings() []RoleSettingsEntry {
	entries := make([]RoleSettingsEntry, 0, len(globalDatabase.roleSettings.Data))
	for key, settings := range globalDatabase.roleSettings.Data {
		entry := RoleSettingsEntry{Key: key, Settings: append([]RoleSetting(nil), settings...)}
		if key.Role.IsValid() {
			entry.RoleName = globalDatabase.rolesByID[key.Role].Name
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Key.Database != entries[j].Key.Database {
			return entries[i].Key.Database < entries[j].Key.Database
		}
		return entries[i].RoleName < entries[j].RoleName
	})
	return entries
}

// removeRoleSettings removes every setting that targets the given role.
func (db *Database) removeRoleSettings(roleID RoleID) {
	for key := range db.roleSettings.Data {
		if key.Role == roleID {
			delete(db.roleSettings.Data, key)
		}
	}
}

// serialize writes the RoleSettings to the given writer.
func (rs *RoleSettings) serialize(writer *utils.Writer) {
	// Version 0
	writer.Uint64(uint64(len(rs.Data)))
	for key, settings := range rs.Data {
		writer.Uint64(uint64(key.Role))
		writer.String(key.Database)
		writer.Uint64(uint64(len(settings)))
		for _, setting := range settings {
			writer.String(setting.Name)
			writer.String(setting.Value)
		}
	}
}

// deserialize reads the RoleSettings from the given reader.
func (rs *RoleSettings) deserialize(version uint32, reader *utils.Reader) {
	rs.Data = make(map[RoleSettingKey][]RoleSetting)
	switch version {
	case 0, 1:
		// Role settings were added in version 2, so there is nothing to read
	case 2:
		keyCount := reader.Uint64()
		for keyIdx := uint64(0); keyIdx < keyCount; keyIdx++ {
			key := RoleSettingKey{Role: RoleID(reader.Uint64()), Database: reader.String()}
			settingCount := reader.Uint64()
			settings := make([]RoleSetting, settingCount)
			for settingIdx := range settings {
				settings[settingIdx].Name = reader.String()
				settings[settingIdx].Value = reader.String()
			}
			rs.Data[key] = settings
		}
	default:
		panic("unexpected version in RoleSettings")
	}
}
