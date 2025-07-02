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

package pg_extension

import (
	"fmt"
	"sync"
)

// Library is a fully-loaded extension library.
type Library struct {
	Magic    PgMagicStruct
	Funcs    map[string]Function
	Version  Version
	internal InternalLoadedLibrary
}

// InternalLoadedLibrary is an interface that is implemented by the specific platform to handle library operations.
type InternalLoadedLibrary interface {
	Lookup(sym string) (uintptr, error)
	Close() error
}

// Function represents an internal library function.
type Function struct {
	Name       string
	Ptr        uintptr
	Args       []int
	APIVersion int
	// TODO: return type?
}

// PgFunctionInfo is a stand-in for the C struct that reports the function information.
type PgFunctionInfo struct {
	APIVersion int32
}

// PgMagicStruct is a stand-in for the C struct that reports the information of the library.
type PgMagicStruct struct {
	Len          int32
	Version      int32
	FuncMaxArgs  int32
	IndexMaxKeys int32
	NameDataLen  int32
	Float4ByVal  int32
	Float8ByVal  int32
}

var (
	// loadedLibraries contains all of the loaded libraries.
	// TODO: need to close all of these before the program ends
	loadedLibraries = make(map[string]*Library)
	// loadedLibrariesMutex gates access to the cached libraries.
	loadedLibrariesMutex = &sync.Mutex{}
)

// LoadLibrary loads the library of the extension, along with preloading all of the functions given.
func LoadLibrary(path string, funcNames []string) (*Library, error) {
	loadedLibrariesMutex.Lock()
	defer loadedLibrariesMutex.Unlock()

	if lib, ok := loadedLibraries[path]; ok {
		return lib, nil
	}
	internalLib, err := loadLibraryInternal(path)
	if err != nil {
		return nil, err
	}
	magicPtr, err := internalLib.Lookup("Pg_magic_func")
	if err != nil {
		return nil, err
	}
	// We don't free the magic struct since it's a pointer to static memory
	magicStructDatum, isNotNull := CallFmgrFunction(magicPtr)
	if !isNotNull {
		return nil, fmt.Errorf("unable to find magic function for `%s`", path)
	}
	magicStruct := *(FromDatum[PgMagicStruct](magicStructDatum))
	lib := &Library{
		Magic:    magicStruct,
		Funcs:    make(map[string]Function),
		internal: internalLib,
	}
	for _, funcName := range funcNames {
		finfoPtr, err := internalLib.Lookup(fmt.Sprintf("pg_finfo_%s", funcName))
		if err != nil {
			return nil, err
		}
		// We don't free finfo since it's a pointer to static memory
		finfoDatum, isNotNull := CallFmgrFunction(finfoPtr)
		apiVersion := 0
		if isNotNull {
			apiVersion = int(FromDatum[PgFunctionInfo](finfoDatum).APIVersion)
		}
		funcPtr, err := internalLib.Lookup(funcName)
		if err != nil {
			return nil, err
		}
		lib.Funcs[funcName] = Function{
			Name:       funcName,
			Ptr:        funcPtr,
			Args:       nil,
			APIVersion: apiVersion,
		}
	}
	loadedLibraries[path] = lib
	return lib, nil
}
