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

package settings

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
)

// GetCurrentSchemas returns all the schemas in the search_path setting, with elements like "$user" excluded
func GetCurrentSchemas(ctx *sql.Context) ([]string, error) {
	searchPathVar, err := ctx.GetSessionVariable(ctx, "search_path")
	if err != nil {
		return nil, err
	}

	pathElems := strings.Split(searchPathVar.(string), ",")
	var path []string

	for _, schemaName := range pathElems {
		schemaName = strings.Trim(schemaName, " ")
		if schemaName == "\"$user\"" {
			continue
		}
		path = append(path, schemaName)
	}

	return path, nil
}

// GetCurrentSchemasAsMap returns the schemas from the search_path setting as a map for easy lookup. Any
// elements like "$user" are excluded.
func GetCurrentSchemasAsMap(ctx *sql.Context) (map[string]struct{}, error) {
	schemas, err := GetCurrentSchemas(ctx)
	if err != nil {
		return nil, err
	}
	schemaMap := make(map[string]struct{}, len(schemas))
	for _, schema := range schemas {
		schemaMap[schema] = struct{}{}
	}
	return schemaMap, nil
}
