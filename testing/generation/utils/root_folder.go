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

package utils

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/errors"
)

// RootFolderLocation is the location of this project's root folder.
type RootFolderLocation struct {
	path string
}

// GetRootFolder returns the location of this project's root folder. This is useful to locate relative directories,
// which will often be read from and written to for generation. It is assumed that this will always be called from
// within an IDE.
func GetRootFolder() (RootFolderLocation, error) {
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		return RootFolderLocation{}, errors.Errorf("failed to fetch the location of the current file")
	}
	return RootFolderLocation{filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation), "../../.."))}, nil
}

// MoveRoot returns a new RootFolderLocation that defines the root at the new path. The parameter should be relative to
// the current root.
func (root RootFolderLocation) MoveRoot(relativePath string) RootFolderLocation {
	return RootFolderLocation{filepath.Clean(filepath.Join(root.path, relativePath))}
}

// GetAbsolutePath returns the absolute path of the given path, which should be relative to the project's root
// folder.
func (root RootFolderLocation) GetAbsolutePath(relativePath string) string {
	return filepath.ToSlash(filepath.Join(root.path, relativePath))
}

// ReadDir is equivalent to os.ReadDir, except that it uses the root path and the given relative path.
func (root RootFolderLocation) ReadDir(relativePath string) ([]os.DirEntry, error) {
	return os.ReadDir(filepath.ToSlash(filepath.Join(root.path, relativePath)))
}

// ReadFile is equivalent to os.ReadFile, except that it uses the root path and the given relative path.
func (root RootFolderLocation) ReadFile(relativePath string) ([]byte, error) {
	return os.ReadFile(filepath.ToSlash(filepath.Join(root.path, relativePath)))
}

// ReadFileFromDirectory is similar to ReadFile, except that it takes in a relative path for the directory containing
// the file, along with the file name to be read. This is purely for convenience, as the same behavior can be
// accomplished using ReadFile.
func (root RootFolderLocation) ReadFileFromDirectory(relativeDirectoryPath string, fileName string) ([]byte, error) {
	return os.ReadFile(filepath.ToSlash(filepath.Join(root.path, relativeDirectoryPath, fileName)))
}

// WriteFile is equivalent to os.WriteFile, except that it uses the root path and the given relative path.
func (root RootFolderLocation) WriteFile(relativePath string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filepath.ToSlash(filepath.Join(root.path, relativePath)), data, perm)
}

// WriteFileToDirectory is similar to WriteFile, except that it takes in a relative path for the directory containing
// the file, along with the file name to be written. This is purely for convenience, as the same behavior can be
// accomplished using WriteFile.
func (root RootFolderLocation) WriteFileToDirectory(relativeDirectoryPath string, fileName string, data []byte, perm os.FileMode) error {
	return os.WriteFile(filepath.ToSlash(filepath.Join(root.path, relativeDirectoryPath, fileName)), data, perm)
}
