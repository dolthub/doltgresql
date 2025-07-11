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
	"cmp"
	"fmt"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// sqlFunctionCapture is a regex to capture the function name as defined in the library. We'll eventually replace this
// and use the nodes from the parser, but this is good enough for the default extensions.
var sqlFunctionCapture = regexp.MustCompile(`(?is)create\s+(?:or\s+replace\s+)?function\s+(.*?)\s*\(.*?\)\s+(?:.*?language c.*?as\s+'.*?'\s*,\s*'(.*?)'.*?;|.*?as\s+'.*?'\s*,\s*'(.*?)'.*?language c.*?;|.*?language c.*?;)`)

// createFunctionStart is a regex to find the beginning of a CREATE FUNCTION statement.
var createFunctionStart = regexp.MustCompile(`(?is)create\s+(?:or\s+replace\s+)?function`)

// ExtensionFiles contains all of the files that are related to or used by an extension.
type ExtensionFiles struct {
	Name            string
	ControlFileName string
	SQLFileNames    []string
	LibraryFileName string
	ControlFileDir  string
	LibraryFileDir  string
	Control         Control
}

// Control contains the contents of the control file.
// https://www.postgresql.org/docs/15/extend-extensions.html#id-1.8.3.20.11
type Control struct {
	Directory      string
	DefaultVersion Version
	Comment        string
	Encoding       string
	ModulePathname string
	Requires       []string
	Superuser      bool
	Trusted        bool
	Relocatable    bool
	Schema         string
	Extra          map[string]string // All entries in here could not be matched to an expected field
}

// Version specifies the major and minor version numbers for an extension.
type Version uint32

// FilenameVersions returns the versions that were encoded in a filename. `From` is the first number, while `To` is the
// second number. If a filename only specifies a single version, this both `From` and `To` will equal one another.
type FilenameVersions struct {
	From Version
	To   Version
}

// LoadExtensions loads information for all extensions that are in the extensions directory of a local Postgres installation.
func LoadExtensions() (map[string]*ExtensionFiles, error) {
	libDir, extDir, err := PostgresDirectories()
	if err != nil {
		return nil, err
	}
	dirEntries, err := os.ReadDir(extDir)
	if err != nil {
		return nil, err
	}
	libEntries, err := os.ReadDir(libDir)
	if err != nil {
		return nil, err
	}
	extensionFiles := make(map[string]*ExtensionFiles)
	// Look for the control files first
	for _, dirEntry := range dirEntries {
		fileName := dirEntry.Name()
		if !dirEntry.IsDir() && strings.HasSuffix(fileName, ".control") {
			extensionName := strings.TrimSuffix(fileName, ".control")
			extensionFiles[extensionName] = &ExtensionFiles{
				Name:            extensionName,
				ControlFileName: fileName,
				ControlFileDir:  extDir,
			}
		}
	}
	// Associate the SQL files and libraries
	for _, extFile := range extensionFiles {
		for _, dirEntry := range dirEntries {
			fileName := dirEntry.Name()
			if !dirEntry.IsDir() && strings.HasPrefix(fileName, extFile.Name+"--") && strings.HasSuffix(fileName, ".sql") {
				extFile.SQLFileNames = append(extFile.SQLFileNames, fileName)
			}
		}
		for _, libEntry := range libEntries {
			fileName := libEntry.Name()
			if !libEntry.IsDir() && strings.HasPrefix(fileName, extFile.Name+".") {
				extFile.LibraryFileName = fileName
				extFile.LibraryFileDir = libDir
			}
		}
		slices.SortFunc(extFile.SQLFileNames, func(aStr, bStr string) int {
			a := DecodeFilenameVersions(extFile.Name, aStr)
			b := DecodeFilenameVersions(extFile.Name, bStr)
			return cmp.Or(
				cmp.Compare(a.From, b.From),
				cmp.Compare(a.To, b.To),
			)
		})
		// Some SQL files are old migration files that won't apply to us, so we can remove them by starting at the first
		// non-migration file.
		for nextLoop := true; nextLoop; {
			nextLoop = false
			for i := 1; i < len(extFile.SQLFileNames); i++ {
				if strings.Count(extFile.SQLFileNames[i], "--") == 1 {
					extFile.SQLFileNames = extFile.SQLFileNames[i:]
					nextLoop = true
					break
				}
			}
		}
		// Load the control file
		if err = extFile.loadControl(); err != nil {
			return nil, err
		}
	}
	return extensionFiles, nil
}

// loadControl loads the control file of an extension.
func (extFile *ExtensionFiles) loadControl() error {
	data, err := os.ReadFile(fmt.Sprintf("%s/%s", extFile.ControlFileDir, extFile.ControlFileName))
	if err != nil {
		return err
	}
	extFile.Control = Control{ // These are the default values
		Directory:      "",
		DefaultVersion: 0,
		Comment:        "",
		Encoding:       "",
		ModulePathname: "",
		Requires:       nil,
		Superuser:      true,
		Trusted:        false,
		Relocatable:    false,
		Schema:         "",
		Extra:          make(map[string]string),
	}
	lines := strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n")
	for _, originalLine := range lines {
		line := strings.TrimSpace(originalLine)
		if commentIdx := strings.Index(line, "#"); commentIdx != -1 {
			line = line[:commentIdx]
		}
		// Line may be empty if it only contained a comment
		if len(line) == 0 {
			continue
		}
		equalsSplit := strings.Index(line, "=")
		if equalsSplit == -1 {
			return fmt.Errorf("malformed `%s.control`:\n%s", extFile.Name, string(data))
		}
		name := strings.ToLower(strings.TrimSpace(line[:equalsSplit]))
		value := line[equalsSplit+1:]
		switch name {
		case "directory":
			extFile.Control.Directory = removeStringQuotations(value)
		case "default_version":
			value = removeStringQuotations(value)
			separator := strings.Index(value, ".")
			if separator == -1 {
				return fmt.Errorf("malformed `%s.control` line:\n%s", extFile.Name, originalLine)
			}
			major, err := strconv.Atoi(value[:separator])
			if err != nil {
				return fmt.Errorf("malformed `%s.control` line:\n%s", extFile.Name, originalLine)
			}
			minor, err := strconv.Atoi(value[separator+1:])
			if err != nil {
				return fmt.Errorf("malformed `%s.control` line:\n%s", extFile.Name, originalLine)
			}
			extFile.Control.DefaultVersion = ToVersion(uint16(major), uint16(minor))
		case "comment":
			extFile.Control.Comment = removeStringQuotations(value)
		case "encoding":
			extFile.Control.Encoding = removeStringQuotations(value)
		case "module_pathname":
			extFile.Control.ModulePathname = removeStringQuotations(value)
		case "requires":
			value = removeStringQuotations(value)
			var entries []string
			for _, entry := range strings.Split(value, ",") {
				entries = append(entries, strings.TrimSpace(entry))
			}
			extFile.Control.Requires = entries
		case "superuser", "trusted", "relocatable":
			value = strings.ToLower(strings.TrimSpace(value))
			var boolValue bool
			if value == "true" {
				boolValue = true
			} else if value == "false" {
				boolValue = false
			} else {
				return fmt.Errorf("malformed `%s.control` line:\n%s", extFile.Name, originalLine)
			}
			switch name {
			case "superuser":
				extFile.Control.Superuser = boolValue
			case "trusted":
				extFile.Control.Trusted = boolValue
			case "relocatable":
				extFile.Control.Relocatable = boolValue
			}
		case "schema":
			extFile.Control.Schema = removeStringQuotations(value)
		default:
			extFile.Control.Extra[name] = value
		}
	}
	return nil
}

// LoadSQLFiles loads the contents of the SQL files used by the extension. These will be in the order that they need to
// be executed.
func (extFile *ExtensionFiles) LoadSQLFiles() ([]string, error) {
	sqlFiles := make([]string, len(extFile.SQLFileNames))
	for i, sqlFileName := range extFile.SQLFileNames {
		data, err := os.ReadFile(fmt.Sprintf("%s/%s", extFile.ControlFileDir, sqlFileName))
		if err != nil {
			return nil, err
		}
		sqlFiles[i] = string(data)
	}
	return sqlFiles, nil
}

// LoadSQLFunctionNames loads all of the library function names that are used by the extension.
func (extFile *ExtensionFiles) LoadSQLFunctionNames() ([]string, error) {
	funcNames := make(map[string]struct{})
	for _, sqlFileName := range extFile.SQLFileNames {
		data, err := os.ReadFile(fmt.Sprintf("%s/%s", extFile.ControlFileDir, sqlFileName))
		if err != nil {
			return nil, err
		}
		fileRemaining := string(data)
	OuterLoop:
		for {
			// We want to advance the file to the start of the next CREATE FUNCTION if one is present
			startIdx := createFunctionStart.FindStringIndex(fileRemaining)
			if startIdx == nil {
				break
			}
			fileRemaining = fileRemaining[startIdx[0]:]
			// We capture the ending semicolon so the regex doesn't match beyond the function definition's boundaries.
			endIdx := strings.IndexRune(fileRemaining, ';')
			if endIdx == -1 {
				break
			}
			matches := sqlFunctionCapture.FindStringSubmatch(fileRemaining[:endIdx+1])
			switch len(matches) {
			case 0:
				break OuterLoop
			case 4:
				if len(matches[2]) > 0 {
					funcNames[matches[2]] = struct{}{}
				} else if len(matches[3]) > 0 {
					funcNames[matches[3]] = struct{}{}
				} else {
					funcNames[matches[1]] = struct{}{}
				}
			default:
				return nil, fmt.Errorf("invalid CREATE FUNCTION string: %s", string(data))
			}
			// We nudge it forward to guarantee that our next CREATE FUNCTION search will grab the next one
			fileRemaining = fileRemaining[6:]
		}
	}
	sortedFuncNames := slices.Sorted(maps.Keys(funcNames))
	return sortedFuncNames, nil
}

// LoadLibrary loads the extension as a library.
func (extFile *ExtensionFiles) LoadLibrary() (*Library, error) {
	if len(extFile.LibraryFileName) == 0 {
		return nil, fmt.Errorf("extension `%s` does not reference a library", extFile.Name)
	}
	funcNames, err := extFile.LoadSQLFunctionNames()
	if err != nil {
		return nil, err
	}
	return LoadLibrary(fmt.Sprintf("%s/%s", extFile.LibraryFileDir, extFile.LibraryFileName), funcNames)
}

// ToVersion creates a version from the given major and minor version numbers.
func ToVersion(major uint16, minor uint16) Version {
	return (Version(major) << 16) + Version(minor)
}

// Major returns the encoded major version number.
func (v Version) Major() uint16 {
	return uint16(v >> 16)
}

// Minor returns the encoded minor version number.
func (v Version) Minor() uint16 {
	return uint16(v)
}

// String returns the version in the `major.minor` format.
func (v Version) String() string {
	return fmt.Sprintf("%d.%d", v.Major(), v.Minor())
}

// DecodeFilenameVersions decodes the version information within the file name. The `sqlFileName` should be the full
// file name (excluding the path), and the SQL file name should contain only the name as
func DecodeFilenameVersions(name string, fileName string) FilenameVersions {
	var versionSubsection string
	if strings.HasSuffix(fileName, ".sql") {
		versionSubsection = strings.TrimSuffix(fileName[len(name)+2: /* We add 2 to account for the -- */], ".sql")
	} else if strings.HasSuffix(fileName, ".control") {
		versionSubsection = strings.TrimSuffix(fileName[len(name)+2: /* We add 2 to account for the -- */], ".control")
	} else {
		// The given name is not a .SQL or .CONTROL file, so we'll just return
		return FilenameVersions{}
	}
	var from, to string
	if dashIdx := strings.Index(versionSubsection, "--"); dashIdx == -1 {
		from = versionSubsection
		to = versionSubsection
	} else {
		from = versionSubsection[:dashIdx]
		to = versionSubsection[dashIdx+2:]
	}
	fromSplit := strings.Index(from, ".")
	toSplit := strings.Index(to, ".")
	if fromSplit == -1 || toSplit == -1 {
		return FilenameVersions{}
	}
	fromMajor, err := strconv.Atoi(from[:fromSplit])
	if err != nil {
		return FilenameVersions{}
	}
	fromMinor, err := strconv.Atoi(from[fromSplit+1:])
	if err != nil {
		return FilenameVersions{}
	}
	toMajor, err := strconv.Atoi(to[:toSplit])
	if err != nil {
		return FilenameVersions{}
	}
	toMinor, err := strconv.Atoi(to[toSplit+1:])
	if err != nil {
		return FilenameVersions{}
	}
	return FilenameVersions{
		From: ToVersion(uint16(fromMajor), uint16(fromMinor)),
		To:   ToVersion(uint16(toMajor), uint16(toMinor)),
	}
}

// removeStringQuotations removes the single quotes that are used to specify that a value is a string.
func removeStringQuotations(str string) string {
	str = strings.TrimSpace(str)
	if strings.HasPrefix(str, "'") {
		return (str[:len(str)-1])[1:]
	}
	return str
}
