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

package extensions

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/extensions/pg_extension"
)

var (
	cachedError   error                                   // TODO: is it better for this to be a local object instead of a global since it's always returned?
	allExtensions map[string]*pg_extension.ExtensionFiles // TODO: should use id.Extension instead of a string
	allLibraries  map[string]*pg_extension.Library        // TODO: close these at some point
	extMutex      = &sync.Mutex{}
	libMutex      = &sync.Mutex{}
)

// Version 0 identifiers are encoded like the following:
// 00AAA1111111111BBB
// `00` is a two-digit version specifier, and we only have version 0 for now
// `AAA` is the three-letter platform specifier
// `1111111111` is the ten-digit library version specifier (i.e. an encoded form of 1.0, 1.5, etc.)
// `BBB` is the extension name, and may be of any length (will have at least 1 character)

// LibraryIdentifier points to a specific extension, as extension functions are dependent on the extension name,
// version, and originating platform (as different platforms may encode data differently).
type LibraryIdentifier string

// InvalidIdentifierReason gives a reason as to why an Identifier is invalid.
type InvalidIdentifierReason uint8

const (
	InvalidIdentifierReason_MismatchedPlatform InvalidIdentifierReason = iota
	InvalidIdentifierReason_MissingLibrary
	InvalidIdentifierReason_InvalidVersion
)

// InvalidIdentifier represents an invalid LibraryIdentifier, providing both the LibraryIdentifier and the reason that
// it is invalid.
type InvalidIdentifier struct {
	Identifier LibraryIdentifier
	Reason     InvalidIdentifierReason
}

// GetExtension returns the extension matching the given name. Returns an error if the extension cannot be found.
func GetExtension(name string) (_ *pg_extension.ExtensionFiles, err error) {
	extMutex.Lock()
	defer extMutex.Unlock()

	if cachedError != nil {
		return nil, cachedError
	}
	if allExtensions == nil {
		allLibraries = make(map[string]*pg_extension.Library)
		allExtensions, err = pg_extension.LoadExtensions()
		if err != nil {
			allExtensions = make(map[string]*pg_extension.ExtensionFiles)
			cachedError = err
			return nil, err
		}
	}
	ext, ok := allExtensions[name]
	if !ok {
		return nil, errors.Errorf(`could not open extension control file "%s.control"`, name)
	}
	return ext, nil
}

// GetExtensionFunction returns the function inside the extension matching the given names. Returns an error if the
// extension or function cannot be found.
func GetExtensionFunction(identifier LibraryIdentifier, funcName string) (_ pg_extension.Function, err error) {
	libMutex.Lock()
	defer libMutex.Unlock()

	extName := identifier.ExtensionName()
	if identifier.Platform() != pg_extension.PLATFORM {
		return pg_extension.Function{}, errors.Errorf(
			`function "%s" was initialized through "%s" on a different platform, which is not supported`, funcName, extName)
	}
	lib, ok := allLibraries[extName]
	if !ok {
		ext, err := GetExtension(extName)
		if err != nil {
			return pg_extension.Function{}, err
		}
		lib, err = ext.LoadLibrary()
		if err != nil {
			return pg_extension.Function{}, err
		}
		lib.Version = ext.Control.DefaultVersion
		allLibraries[extName] = lib
	}
	if lib.Version != identifier.Version() {
		return pg_extension.Function{}, errors.Errorf(
			`function "%s" was initialized through "%s" v%s, the current platform only supports v%s"`,
			funcName, extName, identifier.Version().String(), lib.Version.String())
	}
	libFunc, ok := lib.Funcs[funcName]
	if !ok {
		return pg_extension.Function{}, errors.Errorf(`extension "%s" does not declare the function "%s"`, extName, funcName)
	}
	return libFunc, nil
}

// FindInvalidIdentifiers returns all identifiers that are not valid for the current environment. If the return is
// empty, then that means all of the given identifiers are valid.
func FindInvalidIdentifiers(identifiers ...LibraryIdentifier) []InvalidIdentifier {
	extMutex.Lock()
	defer extMutex.Unlock()

	var invalidIdentifiers []InvalidIdentifier
	if cachedError != nil {
		invalidIdentifiers = make([]InvalidIdentifier, 0, len(identifiers))
		for _, identifier := range identifiers {
			invalidIdentifiers = append(invalidIdentifiers, InvalidIdentifier{
				Identifier: identifier,
				Reason:     InvalidIdentifierReason_MissingLibrary,
			})
		}
		return invalidIdentifiers
	}
	for _, identifier := range identifiers {
		if identifier.Platform() != pg_extension.PLATFORM {
			invalidIdentifiers = append(invalidIdentifiers, InvalidIdentifier{
				Identifier: identifier,
				Reason:     InvalidIdentifierReason_MismatchedPlatform,
			})
			continue
		}
		ext, ok := allExtensions[identifier.ExtensionName()]
		if !ok {
			invalidIdentifiers = append(invalidIdentifiers, InvalidIdentifier{
				Identifier: identifier,
				Reason:     InvalidIdentifierReason_MissingLibrary,
			})
			continue
		}
		if ext.Control.DefaultVersion != identifier.Version() {
			invalidIdentifiers = append(invalidIdentifiers, InvalidIdentifier{
				Identifier: identifier,
				Reason:     InvalidIdentifierReason_InvalidVersion,
			})
			continue
		}
	}
	return invalidIdentifiers
}

// GetPlatform returns the current platform that Doltgres is being executed on. This is encoded as a three-letter string.
func GetPlatform() string {
	return pg_extension.PLATFORM
}

// CreateLibraryIdentifier creates a LibraryIdentifier using the given information.
func CreateLibraryIdentifier(name string, version pg_extension.Version) LibraryIdentifier {
	return LibraryIdentifier(fmt.Sprintf("00%s%010d%s", pg_extension.PLATFORM, version, name))
}

// Platform returns the platform that this LibraryIdentifier was created on.
func (id LibraryIdentifier) Platform() string {
	return string(id[2:5])
}

// Version returns the library version that this LibraryIdentifier references.
func (id LibraryIdentifier) Version() pg_extension.Version {
	val, err := strconv.ParseUint(string(id[5:15]), 10, 32)
	if err != nil {
		// We'll panic for now since this should never happen
		panic(err)
	}
	return pg_extension.Version(val)
}

// ExtensionName returns the extension referenced by this LibraryIdentifier.
func (id LibraryIdentifier) ExtensionName() string {
	return string(id[15:])
}

// DisplayString returns the identifier as a human-readable string.
func (id LibraryIdentifier) DisplayString() string {
	return fmt.Sprintf("%s--%s:%s", id.ExtensionName(), id.Version().String(), id.Platform())
}
