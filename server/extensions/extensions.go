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

// Package extensions holds the registry of every Postgres extension that Doltgres emulates. Each
// emulated extension lives in its own subdirectory and is registered from Init().
package extensions

import (
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	uuid_ossp "github.com/dolthub/doltgresql/server/extensions/uuid-ossp"
)

// registry holds every extension that Doltgres emulates, keyed by its case-sensitive name.
var registry = map[string]*extdef.Extension{}

// Init adds every emulated extension to the registry, making them installable through CREATE EXTENSION.
func Init() {
	register(uuid_ossp.Extension())
}

// register adds the given extension to the registry, and strips the psql meta-commands from its Script.
func register(ext *extdef.Extension) {
	if _, ok := registry[ext.Name]; ok {
		panic(errors.Errorf(`extension "%s" has already been registered`, ext.Name))
	}
	ext.Script = stripMetaCommands(ext.Script)
	registry[ext.Name] = ext
}

// stripMetaCommands removes the psql meta-command lines from an installation script, such as the
// `\echo ... \quit` guard that nearly every script Postgres ships opens with.
func stripMetaCommands(script string) string {
	lines := strings.Split(script, "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, `\`) {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}

// Get returns the emulated extension with the given name, or an error if Doltgres does not emulate it.
func Get(name string) (*extdef.Extension, error) {
	ext, ok := registry[name]
	if !ok {
		return nil, errors.Errorf(`extension "%s" is not available`, name)
	}
	return ext, nil
}

// GetAll returns every extension that Doltgres emulates, keyed by name. The map must not be modified.
func GetAll() map[string]*extdef.Extension {
	return registry
}

// GetFunction returns the implementation of the given symbol within the given extension.
func GetFunction(extensionName string, symbol string) (extdef.Function, error) {
	ext, err := Get(extensionName)
	if err != nil {
		return nil, err
	}
	f, ok := ext.Functions[symbol]
	if !ok {
		return nil, errors.Errorf(`extension "%s" does not declare the function "%s"`, extensionName, symbol)
	}
	return f, nil
}
