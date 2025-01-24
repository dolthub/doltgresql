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

package main

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/cockroachdb/errors"
)

var benchmarkFolder BenchmarkFolderLocation // benchmarkFolder is the disk location of the benchmark folder

// BenchmarkFolderLocation is the location of this project's root folder.
type BenchmarkFolderLocation struct {
	path string
}

// GetBenchmarkFolder returns the location of the benchmark folder (scripts/mini_benchmark). This is used to find the
// absolute position of our output files.
func GetBenchmarkFolder() (BenchmarkFolderLocation, error) {
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		return BenchmarkFolderLocation{}, errors.Errorf("failed to fetch the location of the current file")
	}
	benchmarkFolder = BenchmarkFolderLocation{filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation),
		"../../../scripts/mini_sysbench"))}
	return benchmarkFolder, nil
}

// MoveRoot returns a new BenchmarkFolderLocation that defines the root at the new path. The parameter should be
// relative to the current root.
func (root BenchmarkFolderLocation) MoveRoot(relativePath string) BenchmarkFolderLocation {
	return BenchmarkFolderLocation{filepath.Clean(filepath.Join(root.path, relativePath))}
}

// GetAbsolutePath returns the absolute path of the given path, which should be relative to the project's root
// folder.
func (root BenchmarkFolderLocation) GetAbsolutePath(relativePath string) string {
	return filepath.ToSlash(filepath.Join(root.path, relativePath))
}

// Exists returns whether the file or directory at the given path (relative to the root path) exists. Returns an error
// if the check was unable to be completed.
func (root BenchmarkFolderLocation) Exists(relativePath string) (bool, error) {
	_, err := os.Stat(root.GetAbsolutePath(relativePath))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// ReadDir is equivalent to os.ReadDir, except that it uses the root path and the given relative path.
func (root BenchmarkFolderLocation) ReadDir(relativePath string) ([]os.DirEntry, error) {
	return os.ReadDir(root.GetAbsolutePath(relativePath))
}

// ReadFile is equivalent to os.ReadFile, except that it uses the root path and the given relative path.
func (root BenchmarkFolderLocation) ReadFile(relativePath string) ([]byte, error) {
	return os.ReadFile(root.GetAbsolutePath(relativePath))
}

// WriteFile is equivalent to os.WriteFile, except that it uses the root path and the given relative path.
func (root BenchmarkFolderLocation) WriteFile(relativePath string, data []byte, perm os.FileMode) error {
	directory := filepath.ToSlash(filepath.Dir(relativePath))
	exists, err := root.Exists(directory)
	if err != nil {
		return err
	}
	if !exists {
		if err = os.MkdirAll(root.GetAbsolutePath(directory), 0644); err != nil {
			return err
		}
	}
	return os.WriteFile(root.GetAbsolutePath(relativePath), data, perm)
}

// init is used to load the location of the benchmark folder.
func init() {
	var err error
	benchmarkFolder, err = GetBenchmarkFolder()
	if err != nil {
		panic(err)
	}
}
