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

// Privilege represents some permission for a database object.
// https://www.postgresql.org/docs/15/ddl-priv.html
type Privilege string

const (
	Privilege_SELECT       = "r"
	Privilege_INSERT       = "a"
	Privilege_UPDATE       = "w"
	Privilege_DELETE       = "d"
	Privilege_TRUNCATE     = "D"
	Privilege_REFERENCES   = "x"
	Privilege_TRIGGER      = "t"
	Privilege_CREATE       = "C"
	Privilege_CONNECT      = "c"
	Privilege_TEMPORARY    = "T"
	Privilege_EXECUTE      = "X"
	Privilege_USAGE        = "U"
	Privilege_SET          = "s"
	Privilege_ALTER_SYSTEM = "A"
	Privilege_DROP         = "Y"
)

// PrivilegeObject is the database object that privileges are applied to.
// https://www.postgresql.org/docs/15/ddl-priv.html
type PrivilegeObject byte

const (
	PrivilegeObject_DATABASE PrivilegeObject = iota
	PrivilegeObject_DOMAIN
	PrivilegeObject_FUNCTION // Also applies to procedures and routines
	PrivilegeObject_FOREIGN_DATA_WRAPPER
	PrivilegeObject_FOREIGN_SERVER
	PrivilegeObject_LANGUAGE
	PrivilegeObject_LARGE_OBJECT
	PrivilegeObject_PARAMETER
	PrivilegeObject_SCHEMA
	PrivilegeObject_SEQUENCE
	PrivilegeObject_TABLE
	PrivilegeObject_TABLE_COLUMN
	PrivilegeObject_TABLESPACE
	PrivilegeObject_TYPE
)

// GrantedPrivilege specifies details.
type GrantedPrivilege struct {
	Privilege
	GrantedBy RoleID
}

// GetAllPrivileges returns every Privilege.
func GetAllPrivileges() []Privilege {
	return []Privilege{
		Privilege_SELECT,
		Privilege_INSERT,
		Privilege_UPDATE,
		Privilege_DELETE,
		Privilege_TRUNCATE,
		Privilege_REFERENCES,
		Privilege_TRIGGER,
		Privilege_CREATE,
		Privilege_CONNECT,
		Privilege_TEMPORARY,
		Privilege_EXECUTE,
		Privilege_USAGE,
		Privilege_SET,
		Privilege_ALTER_SYSTEM,
		Privilege_DROP,
	}
}

// GetAllPrivilegeObjects returns every PrivilegeObject.
func GetAllPrivilegeObjects() []PrivilegeObject {
	return []PrivilegeObject{
		PrivilegeObject_DATABASE,
		PrivilegeObject_DOMAIN,
		PrivilegeObject_FUNCTION,
		PrivilegeObject_FOREIGN_DATA_WRAPPER,
		PrivilegeObject_FOREIGN_SERVER,
		PrivilegeObject_LANGUAGE,
		PrivilegeObject_LARGE_OBJECT,
		PrivilegeObject_PARAMETER,
		PrivilegeObject_SCHEMA,
		PrivilegeObject_SEQUENCE,
		PrivilegeObject_TABLE,
		PrivilegeObject_TABLE_COLUMN,
		PrivilegeObject_TABLESPACE,
		PrivilegeObject_TYPE,
	}
}

// ACLAbbreviation returns the name of the privilege using the Access Control List abbreviation.
func (p Privilege) ACLAbbreviation() string {
	return string(p)
}

// String returns the name of the privilege (uppercased).
func (p Privilege) String() string {
	switch p {
	case Privilege_SELECT:
		return "SELECT"
	case Privilege_INSERT:
		return "INSERT"
	case Privilege_UPDATE:
		return "UPDATE"
	case Privilege_DELETE:
		return "DELETE"
	case Privilege_TRUNCATE:
		return "TRUNCATE"
	case Privilege_REFERENCES:
		return "REFERENCES"
	case Privilege_TRIGGER:
		return "TRIGGER"
	case Privilege_CREATE:
		return "CREATE"
	case Privilege_CONNECT:
		return "CONNECT"
	case Privilege_TEMPORARY:
		return "TEMPORARY"
	case Privilege_EXECUTE:
		return "EXECUTE"
	case Privilege_USAGE:
		return "USAGE"
	case Privilege_SET:
		return "SET"
	case Privilege_ALTER_SYSTEM:
		return "ALTER SYSTEM"
	case Privilege_DROP:
		return "DROP"
	default:
		return "UNKNOWN"
	}
}

// AllPrivileges returns all valid privileges that may be applied to this object.
func (po PrivilegeObject) AllPrivileges() []Privilege {
	switch po {
	case PrivilegeObject_DATABASE:
		return []Privilege{Privilege_CREATE, Privilege_TEMPORARY, Privilege_CONNECT, Privilege_DROP}
	case PrivilegeObject_DOMAIN:
		return []Privilege{Privilege_USAGE, Privilege_DROP}
	case PrivilegeObject_FUNCTION:
		return []Privilege{Privilege_EXECUTE, Privilege_DROP}
	case PrivilegeObject_FOREIGN_DATA_WRAPPER:
		return []Privilege{Privilege_USAGE, Privilege_DROP}
	case PrivilegeObject_FOREIGN_SERVER:
		return []Privilege{Privilege_USAGE, Privilege_DROP}
	case PrivilegeObject_LANGUAGE:
		return []Privilege{Privilege_USAGE, Privilege_DROP}
	case PrivilegeObject_LARGE_OBJECT:
		return []Privilege{Privilege_SELECT, Privilege_UPDATE, Privilege_DROP}
	case PrivilegeObject_PARAMETER:
		return []Privilege{Privilege_SET, Privilege_ALTER_SYSTEM, Privilege_DROP}
	case PrivilegeObject_SCHEMA:
		return []Privilege{Privilege_USAGE, Privilege_CREATE, Privilege_DROP}
	case PrivilegeObject_SEQUENCE:
		return []Privilege{Privilege_SELECT, Privilege_UPDATE, Privilege_USAGE, Privilege_DROP}
	case PrivilegeObject_TABLE:
		return []Privilege{Privilege_INSERT, Privilege_SELECT, Privilege_UPDATE, Privilege_DELETE, Privilege_TRUNCATE, Privilege_REFERENCES, Privilege_TRIGGER, Privilege_DROP}
	case PrivilegeObject_TABLE_COLUMN:
		return []Privilege{Privilege_INSERT, Privilege_SELECT, Privilege_UPDATE, Privilege_REFERENCES, Privilege_DROP}
	case PrivilegeObject_TABLESPACE:
		return []Privilege{Privilege_CREATE, Privilege_DROP}
	case PrivilegeObject_TYPE:
		return []Privilege{Privilege_USAGE, Privilege_DROP}
	default:
		panic("unknown privilege object")
	}
}

// DefaultPublicPrivileges return the default PUBLIC privileges for this object.
func (po PrivilegeObject) DefaultPublicPrivileges() []Privilege {
	switch po {
	case PrivilegeObject_DATABASE:
		return []Privilege{Privilege_TEMPORARY, Privilege_CONNECT}
	case PrivilegeObject_DOMAIN:
		return []Privilege{Privilege_USAGE}
	case PrivilegeObject_FUNCTION:
		return []Privilege{Privilege_EXECUTE}
	case PrivilegeObject_FOREIGN_DATA_WRAPPER:
		return nil
	case PrivilegeObject_FOREIGN_SERVER:
		return nil
	case PrivilegeObject_LANGUAGE:
		return []Privilege{Privilege_USAGE}
	case PrivilegeObject_LARGE_OBJECT:
		return nil
	case PrivilegeObject_PARAMETER:
		return nil
	case PrivilegeObject_SCHEMA:
		return nil
	case PrivilegeObject_SEQUENCE:
		return nil
	case PrivilegeObject_TABLE:
		return nil
	case PrivilegeObject_TABLE_COLUMN:
		return nil
	case PrivilegeObject_TABLESPACE:
		return nil
	case PrivilegeObject_TYPE:
		return []Privilege{Privilege_USAGE}
	default:
		panic("unknown privilege object")
	}
}

// IsValid returns whether the given Privilege is valid for the PrivilegeObject, as not all privileges are valid for all
// objects.
func (po PrivilegeObject) IsValid(privilege Privilege) bool {
	for _, poPriv := range po.AllPrivileges() {
		if privilege == poPriv {
			return true
		}
	}
	return false
}

// String returns the name of the privilege (uppercased).
func (po PrivilegeObject) String() string {
	switch po {
	case PrivilegeObject_DATABASE:
		return "DATABASE"
	case PrivilegeObject_DOMAIN:
		return "DOMAIN"
	case PrivilegeObject_FUNCTION:
		return "FUNCTION"
	case PrivilegeObject_FOREIGN_DATA_WRAPPER:
		return "FOREIGN DATA WRAPPER"
	case PrivilegeObject_FOREIGN_SERVER:
		return "FOREIGN SERVER"
	case PrivilegeObject_LANGUAGE:
		return "LANGUAGE"
	case PrivilegeObject_LARGE_OBJECT:
		return "LARGE_OBJECT"
	case PrivilegeObject_PARAMETER:
		return "PARAMETER"
	case PrivilegeObject_SCHEMA:
		return "SCHEMA"
	case PrivilegeObject_SEQUENCE:
		return "SEQUENCE"
	case PrivilegeObject_TABLE:
		return "TABLE"
	case PrivilegeObject_TABLE_COLUMN:
		return "TABLE COLUMN"
	case PrivilegeObject_TABLESPACE:
		return "TABLESPACE"
	case PrivilegeObject_TYPE:
		return "TYPE"
	default:
		return "UNKNOWN"
	}
}
