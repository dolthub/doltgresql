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

package extensions

import (
	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	uuid_ossp "github.com/dolthub/doltgresql/server/extensions/uuid-ossp/v1_1"
)

// registry holds every extension that Doltgres emulates, keyed by its case-sensitive name.
var registry = map[string]*extdef.Extension{}

// implementations holds every registered extension's routines, keyed by the extension's name and the routine's symbol.
var implementations = map[string]map[string]extdef.Function{}

// Init adds every emulated extension to the registry, making them installable through CREATE EXTENSION.
func Init() {
	Register(uuid_ossp.Extension())
}

// Register adds the given extension to the registry.
func Register(ext *extdef.Extension) {
	if _, ok := registry[ext.Name]; ok {
		panic(errors.Errorf(`extension "%s" has already been registered`, ext.Name))
	}
	symbols := make(map[string]extdef.Function, len(ext.Routines))
	for _, routine := range ext.Routines {
		if _, ok := symbols[routine.Symbol]; ok {
			panic(errors.Errorf(`extension "%s" declares the symbol "%s" twice`, ext.Name, routine.Symbol))
		}
		symbols[routine.Symbol] = routine.Impl
	}
	registry[ext.Name] = ext
	implementations[ext.Name] = symbols
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
	symbols, ok := implementations[extensionName]
	if !ok {
		return nil, errors.Errorf(`extension "%s" is not available`, extensionName)
	}
	f, ok := symbols[symbol]
	if !ok {
		return nil, errors.Errorf(`extension "%s" does not declare the function "%s"`, extensionName, symbol)
	}
	return f, nil
}
