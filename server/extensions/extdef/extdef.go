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

package extdef

import (
	"github.com/dolthub/go-mysql-server/sql"
)

// Function is the Go implementation of a single function that an extension provides.
type Function func(ctx *sql.Context, args ...any) (any, error)

// Control holds the information that an extension declares in its control file.
// https://www.postgresql.org/docs/15/extend-extensions.html#id-1.8.3.20.11
type Control struct {
	DefaultVersion string
	Comment        string
	Requires       []string
	Superuser      bool
	Trusted        bool
	Relocatable    bool
	Schema         string
}

// Extension is a Postgres extension that Doltgres emulates.
type Extension struct {
	Name     string
	Control  Control
	Routines []Routine
}

// Routine is a function that an extension provides. Symbol is its C link symbol, which is unique within the extension.
type Routine struct {
	Name       string
	Symbol     string
	Parameters []Parameter
	Returns    string
	Strict     bool
	Impl       Function
}

// Parameter is a single input parameter of a routine.
type Parameter struct {
	Name string
	Type string
}
