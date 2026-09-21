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
	"slices"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql/expression/function/vector"

	"github.com/dolthub/doltgresql/server/extensions/extdef"
	uuid_ossp "github.com/dolthub/doltgresql/server/extensions/uuid-ossp/v1_1"
	pgvector "github.com/dolthub/doltgresql/server/extensions/vector/v0_8_6"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// registry holds every extension that Doltgres emulates, keyed by its case-sensitive name.
var registry = map[string]*extdef.Extension{}

// implementations holds every registered extension's routines, keyed by the extension's name and the routine's symbol.
var implementations = map[string]map[string]extdef.Function{}

// codecs holds every registered type's storage codec, keyed by the name of the type's send routine.
var codecs = map[string]*pgtypes.TypeCodec{}

// distanceTypes holds the distance type of every tagged distance routine, keyed by the extension's name and the
// routine's symbol.
var distanceTypes = map[string]map[string]vector.DistanceType{}

// operatorClasses holds every registered extension's operator classes, keyed by name.
var operatorClasses = map[string]extdef.OperatorClass{}

// accessMethods holds every registered extension's index access methods, keyed by name.
var accessMethods = map[string]extdef.AccessMethod{}

// Init adds every emulated extension to the registry, making them installable through CREATE EXTENSION.
func Init() {
	pgtypes.LoadTypeCodec = getTypeCodec
	Register(uuid_ossp.Extension())
	Register(pgvector.Extension())
}

// Register adds the given extension to the registry.
func Register(ext *extdef.Extension) {
	if _, ok := registry[ext.Name]; ok {
		panic(errors.Errorf(`extension "%s" has already been registered`, ext.Name))
	}
	symbols := make(map[string]extdef.Function, len(ext.Routines))
	names := make(map[string]string, len(ext.Routines))
	distances := map[string]vector.DistanceType{}
	for _, routine := range ext.Routines {
		if _, ok := symbols[routine.Symbol]; ok {
			panic(errors.Errorf(`extension "%s" declares the symbol "%s" twice`, ext.Name, routine.Symbol))
		}
		symbols[routine.Symbol] = routine.Impl
		names[routine.Symbol] = routine.Name
		if routine.DistanceType != nil {
			distances[routine.Symbol] = routine.DistanceType
		}
	}
	for _, declared := range ext.Types {
		if declared.Codec == nil {
			continue
		}
		sendName, ok := names[declared.Send]
		if !ok {
			panic(errors.Errorf(`extension "%s" declares a codec for the type "%s", which has no send routine`, ext.Name, declared.Name))
		}
		if _, ok = codecs[sendName]; ok {
			panic(errors.Errorf(`a codec has already been registered under the send routine "%s"`, sendName))
		}
		codecs[sendName] = declared.Codec
	}
	for _, opclass := range ext.OperatorClasses {
		if _, ok := operatorClasses[opclass.Name]; ok {
			panic(errors.Errorf(`the operator class "%s" has already been registered`, opclass.Name))
		}
		operatorClasses[opclass.Name] = opclass
	}
	for _, am := range ext.AccessMethods {
		if _, ok := accessMethods[am.Name]; ok {
			panic(errors.Errorf(`the access method "%s" has already been registered`, am.Name))
		}
		accessMethods[am.Name] = am
	}
	registry[ext.Name] = ext
	implementations[ext.Name] = symbols
	distanceTypes[ext.Name] = distances
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

// GetRoutine returns the routine matching the given name and parameter count within any registered extension.
func GetRoutine(name string, paramCount int) (extdef.Routine, bool) {
	for _, ext := range registry {
		for _, routine := range ext.Routines {
			if routine.Name == name && len(routine.Parameters) == paramCount {
				return routine, true
			}
		}
	}
	return extdef.Routine{}, false
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

// getTypeCodec returns the storage codec registered under the given send routine name, or nil when no registered type
// uses that name.
func getTypeCodec(sendName string) *pgtypes.TypeCodec {
	return codecs[sendName]
}

// GetDistanceType returns the distance type of the given routine, or nil when the routine is not a distance routine.
func GetDistanceType(extensionName string, symbol string) vector.DistanceType {
	return distanceTypes[extensionName][symbol]
}

// GetOperatorClass returns the operator class registered under the given name.
func GetOperatorClass(name string) (extdef.OperatorClass, bool) {
	opclass, ok := operatorClasses[name]
	return opclass, ok
}

// GetDefaultOperatorClass returns the default operator class for the given type under the given access method.
func GetDefaultOperatorClass(accessMethod string, typeName string) (extdef.OperatorClass, bool) {
	for _, opclass := range operatorClasses {
		if opclass.Type == typeName && slices.Contains(opclass.DefaultFor, accessMethod) {
			return opclass, true
		}
	}
	return extdef.OperatorClass{}, false
}

// GetOperatorClassForIndex returns the operator class declared for the given type whose distance type serves the
// given metric.
func GetOperatorClassForIndex(typeName string, distanceType vector.DistanceType) (extdef.OperatorClass, bool) {
	if distanceType == nil {
		return extdef.OperatorClass{}, false
	}
	for _, opclass := range operatorClasses {
		if opclass.Type == typeName && opclass.DistanceType != nil && opclass.DistanceType.CanEval(distanceType) {
			return opclass, true
		}
	}
	return extdef.OperatorClass{}, false
}

// GetAccessMethod returns the index access method registered under the given name.
func GetAccessMethod(name string) (extdef.AccessMethod, bool) {
	am, ok := accessMethods[name]
	return am, ok
}
