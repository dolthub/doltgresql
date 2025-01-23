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
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/jackc/pgx/v5/pgproto3"
)

var copyFrom, _ = regexp.Compile(`COPY\s+\S+\s+FROM\s+'([a-zA-Z0-9_\/\\%#@!~+=:.-]+)'`)

// RewriteCopyToLocal rewrites a COPY ... FROM query so that it points to the data file that is located in our data
// directory.
func RewriteCopyToLocal(query *pgproto3.Query) (*pgproto3.Query, error) {
	matches := copyFrom.FindStringSubmatch(query.String)
	if len(matches) != 2 {
		return query, nil
	}
	fileName := filepath.Base(matches[1])
	relativeFileName := "data/" + fileName
	if ok, err := regressionFolder.Exists(relativeFileName); err != nil {
		return query, err
	} else if !ok {
		return query, errors.Errorf("file does not exist: '%s'", regressionFolder.GetAbsolutePath(relativeFileName))
	}
	return &pgproto3.Query{
		String: strings.ReplaceAll(query.String, matches[1], regressionFolder.GetAbsolutePath(relativeFileName)),
	}, nil
}

// RewriteCopyToFileOnly rewrites a COPY ... FROM query so that it only contains the file name. This is mainly used so
// that serialization does not any local information from the user's machine.
func RewriteCopyToFileOnly(query *pgproto3.Query) *pgproto3.Query {
	matches := copyFrom.FindStringSubmatch(query.String)
	if len(matches) != 2 {
		return query
	}
	return &pgproto3.Query{
		String: strings.ReplaceAll(query.String, matches[1], filepath.Base(matches[1])),
	}
}
