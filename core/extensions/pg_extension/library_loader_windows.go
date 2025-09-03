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

//go:build windows

package pg_extension

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"
)

// PLATFORM specifies which platform applies to the current library loader. This will always be a three-letter string.
const PLATFORM = "WIN"

//go:embed output/postgres.exe
var libDefBytes []byte

//go:embed output/pg_extension.dll
var dllBytes []byte

// winLib is the Windows-specific implementation of InternalLoadedLibrary.
type winLib struct{ dll syscall.Handle }

var _ InternalLoadedLibrary = (*winLib)(nil)
var addPGBinDir = &sync.Once{}

// loadLibraryInternal handles the loading of an extension's DLL.
func loadLibraryInternal(path string) (InternalLoadedLibrary, error) {
	addPGBinDir.Do(func() {
		_, currentFileLocation, _, ok := runtime.Caller(0)
		if !ok || len(currentFileLocation) == 0 {
			panic("cannot find the directory where this file exists")
		}
		// There are three scenarios that we need to consider when attempting to load the DLL:
		// 1) The DLL exists in an output folder (this will be true for development)
		// 2) The DLL exists alongside the binary
		// 3) The DLL does not exist alongside the binary (or is the wrong version)
		// In the third situation, we write the contained DLL and definition file alongside the binary, so that we'll
		// always end up in the second situation. This enables both developmental and deployment workflows without
		// explicit configuration.
		var dllDir string
		if _, err := os.Stat(filepath.Join(filepath.Dir(currentFileLocation), "output", "postgres.exe")); err == nil {
			dllDir = filepath.Join(filepath.Dir(currentFileLocation), "output")
		} else {
			currentBinaryLocation, err := os.Executable()
			if err != nil {
				panic(fmt.Errorf("cannot find where the executable was launched:\n%s", err.Error()))
			}
			dllDir = filepath.Dir(currentBinaryLocation)
			shouldWriteFiles := false
			if _, err := os.Stat(filepath.Join(dllDir, "postgres.exe")); err != nil {
				shouldWriteFiles = true
			} else {
				func() {
					// If the DLL hash doesn't match our hash, then we overwrite it
					extDll, err := os.Open(filepath.Join(filepath.Dir(currentBinaryLocation), "pg_extension.dll"))
					if err != nil {
						shouldWriteFiles = true
						return
					}
					defer func() {
						_ = extDll.Close()
					}()
					dllSha := sha256.Sum256(dllBytes)
					extDllSha := sha256.New()
					_, _ = io.Copy(extDllSha, extDll)
					shouldWriteFiles = !bytes.Equal(extDllSha.Sum(nil), dllSha[:])
				}()
			}
			if shouldWriteFiles {
				writeLocation := filepath.Dir(currentBinaryLocation)
				_ = os.WriteFile(filepath.Join(writeLocation, "postgres.exe"), libDefBytes, 0755)
				_ = os.WriteFile(filepath.Join(writeLocation, "pg_extension.dll"), dllBytes, 0755)
			}
		}
		dirPtr, err := syscall.UTF16PtrFromString(dllDir)
		if err != nil {
			panic(err)
		}
		_, _, _ = syscall.MustLoadDLL("kernel32.dll").MustFindProc("SetDllDirectoryW").Call(uintptr(unsafe.Pointer(dirPtr)))
		_, _ = syscall.LoadLibrary(filepath.Join(dllDir, "pg_extension.dll"))
	})
	d, err := syscall.LoadLibrary(path)
	if err != nil {
		return nil, err
	}
	return &winLib{dll: d}, nil
}

// Lookup implements the interface InternalLoadedLibrary.
func (w *winLib) Lookup(sym string) (uintptr, error) {
	candidates := []string{
		sym,
		"_" + sym,
		sym + "@0",
		"_" + sym + "@0",
	}
	for bytes := 4; bytes <= 64; bytes += 4 {
		candidates = append(candidates,
			fmt.Sprintf("%s@%d", sym, bytes),
			fmt.Sprintf("_%s@%d", sym, bytes))
	}

	for _, name := range candidates {
		if p, err := syscall.GetProcAddress(w.dll, name); err == nil {
			return p, nil
		}
	}
	return 0, fmt.Errorf("symbol %s not found", sym)
}

// Close implements the interface InternalLoadedLibrary.
func (w *winLib) Close() error {
	return syscall.FreeLibrary(w.dll)
}
