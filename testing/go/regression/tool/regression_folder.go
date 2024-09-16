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
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/utils"
)

var regressionFolder RegressionFolderLocation // regressionFolder is the disk location of the regression folder

// RegressionFolderLocation is the location of this project's root folder.
type RegressionFolderLocation struct {
	path string
}

// GetRegressionFolder returns the location of the regression folder (testing/go/regression). This is used to find the
// absolute position of our data files.
func GetRegressionFolder() (RegressionFolderLocation, error) {
	_, currentFileLocation, _, ok := runtime.Caller(0)
	if !ok {
		return RegressionFolderLocation{}, fmt.Errorf("failed to fetch the location of the current file")
	}
	regressionFolder = RegressionFolderLocation{filepath.Clean(filepath.Join(filepath.Dir(currentFileLocation), ".."))}
	return regressionFolder, nil
}

// MoveRoot returns a new RegressionFolderLocation that defines the root at the new path. The parameter should be relative to
// the current root.
func (root RegressionFolderLocation) MoveRoot(relativePath string) RegressionFolderLocation {
	return RegressionFolderLocation{filepath.Clean(filepath.Join(root.path, relativePath))}
}

// GetAbsolutePath returns the absolute path of the given path, which should be relative to the project's root
// folder.
func (root RegressionFolderLocation) GetAbsolutePath(relativePath string) string {
	return filepath.ToSlash(filepath.Join(root.path, relativePath))
}

// Exists returns whether the file or directory at the given path (relative to the root path) exists. Returns an error
// if the check was unable to be completed.
func (root RegressionFolderLocation) Exists(relativePath string) (bool, error) {
	_, err := os.Stat(root.GetAbsolutePath(relativePath))
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

// ReadDir is equivalent to os.ReadDir, except that it uses the root path and the given relative path.
func (root RegressionFolderLocation) ReadDir(relativePath string) ([]os.DirEntry, error) {
	return os.ReadDir(root.GetAbsolutePath(relativePath))
}

// ReadFile is equivalent to os.ReadFile, except that it uses the root path and the given relative path.
func (root RegressionFolderLocation) ReadFile(relativePath string) ([]byte, error) {
	return os.ReadFile(root.GetAbsolutePath(relativePath))
}

// ReadMessages reads the messages from the file at the given path (relative to the root path). It is assumed that this
// file was previously written to using WriteMessages.
func (root RegressionFolderLocation) ReadMessages(relativePath string) ([]pgproto3.Message, error) {
	fileData, err := root.ReadFile(relativePath)
	if err != nil {
		return nil, err
	}
	reader := utils.NewReader(fileData)
	messages := make([]pgproto3.Message, reader.Uint32())
	for i := 0; i < len(messages); i++ {
		var err error
		messages[i], err = FromMessageType(MessageType(reader.Uint16()))
		if err != nil {
			return nil, err
		}
		if err = messages[i].Decode(reader.ByteSlice()); err != nil {
			return nil, err
		}
		if query, ok := messages[i].(*pgproto3.Query); ok {
			messages[i], err = RewriteCopyToLocal(query)
			if err != nil {
				return nil, err
			}
		}
	}
	if !reader.IsEmpty() {
		return nil, fmt.Errorf("file has additional data")
	}
	return messages, nil
}

// ReadReplayTrackers reads the replay trackers from the file at the given path (relative to the root path). It is
// assumed that this file was previously written to using WriteReplayTrackers.
func (root RegressionFolderLocation) ReadReplayTrackers(relativePath string) ([]*ReplayTracker, error) {
	fileData, err := root.ReadFile(relativePath)
	if err != nil {
		return nil, err
	}
	return DeserializeTrackers(fileData)
}

// WriteFile is equivalent to os.WriteFile, except that it uses the root path and the given relative path.
func (root RegressionFolderLocation) WriteFile(relativePath string, data []byte, perm os.FileMode) error {
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

// WriteMessages writes the given messages to the file at the given path (relative to the root path). It is assumed that
// this file will be read using ReadMessages.
func (root RegressionFolderLocation) WriteMessages(relativePath string, messages []pgproto3.Message, perm os.FileMode) error {
	writer := utils.NewWriter(1048576)
	writer.Uint32(uint32(len(messages)))
	for _, message := range messages {
		messageType, err := ToMessageType(message)
		if err != nil {
			return err
		}
		writer.Uint16(uint16(messageType))
		data, err := EncodeMessage(message)
		if err != nil {
			return err
		}
		writer.ByteSlice(data)
	}
	return root.WriteFile(relativePath, writer.Data(), perm)
}

// WriteReplayTrackers writes the given replay trackers to the file at the given path (relative to the root path). It is
// assumed that this file will be read using ReadReplayTrackers.
func (root RegressionFolderLocation) WriteReplayTrackers(relativePath string, trackers []*ReplayTracker, perm os.FileMode) error {
	return root.WriteFile(relativePath, SerializeTrackers(trackers...), perm)
}

// init is used to load the location of the regression folder.
func init() {
	var err error
	regressionFolder, err = GetRegressionFolder()
	if err != nil {
		panic(err)
	}
}
