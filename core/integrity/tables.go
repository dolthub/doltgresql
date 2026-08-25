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

package integrity

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb/durable"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/val"
	"github.com/dolthub/go-mysql-server/sql"
)

// TableInfo describes one table in one root value, along with the tuple descriptors needed to scan it.
type TableInfo struct {
	Name    doltdb.TableName
	Tbl     *doltdb.Table
	Sch     schema.Schema
	KeyDesc *val.TupleDesc
	ValDesc *val.TupleDesc

	// AdaptiveValueCols are the names of value (non-PK) columns with an adaptive encoding.
	AdaptiveValueCols []string
	// AdaptiveKeyCols are the names of key (PK) columns with an adaptive encoding.
	AdaptiveKeyCols []string
}

// Impacted returns whether this table's schema can be affected by the value_address_offsets corruption:
// its value tuples contain one or more adaptive-encoded fields.
func (ti *TableInfo) Impacted() bool {
	return len(ti.AdaptiveValueCols) > 0
}

// KeyImpacted returns whether this table's key tuples contain adaptive-encoded fields. Out-of-band
// values in key tuples cannot be recorded in the ProllyTreeNode message format at all, so they are
// scanned and reported, but cannot be repaired by rewriting nodes.
func (ti *TableInfo) KeyImpacted() bool {
	return len(ti.AdaptiveKeyCols) > 0
}

// RowMap returns the primary row storage of the table as a prolly map.
func (ti *TableInfo) RowMap(ctx context.Context) (prolly.Map, error) {
	idx, err := ti.Tbl.GetRowData(ctx)
	if err != nil {
		return prolly.Map{}, err
	}
	return durable.ProllyMapFromIndex(idx)
}

// TablesForRoot enumerates all tables in all database schemas of the given root value.
// |sctx| must be a *sql.Context: doltgres extended type deserialization requires one.
func TablesForRoot(sctx *sql.Context, root doltdb.RootValue, ns tree.NodeStore) ([]*TableInfo, error) {
	schemaNames, err := schemaNamesForRoot(sctx, root)
	if err != nil {
		return nil, err
	}

	var infos []*TableInfo
	for _, schemaName := range schemaNames {
		tableNames, err := root.GetTableNames(sctx, schemaName, false)
		if err != nil {
			return nil, err
		}
		for _, tableName := range tableNames {
			tname := doltdb.TableName{Name: tableName, Schema: schemaName}
			tbl, ok, err := root.GetTable(sctx, tname)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get table %s", tname.String())
			}
			if !ok {
				continue
			}
			sch, err := tbl.GetSchema(sctx)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to get schema for table %s", tname.String())
			}
			info := &TableInfo{
				Name:    tname,
				Tbl:     tbl,
				Sch:     sch,
				KeyDesc: sch.GetKeyDescriptor(ns),
				ValDesc: sch.GetValueDescriptor(ns),
			}
			for _, col := range sch.GetNonPKCols().GetColumns() {
				if col.Virtual {
					continue
				}
				if val.IsAdaptiveEncoding(col.TypeInfo.Encoding()) {
					info.AdaptiveValueCols = append(info.AdaptiveValueCols, col.Name)
				}
			}
			for _, col := range sch.GetPKCols().GetColumns() {
				if val.IsAdaptiveEncoding(col.TypeInfo.Encoding()) {
					info.AdaptiveKeyCols = append(info.AdaptiveKeyCols, col.Name)
				}
			}
			infos = append(infos, info)
		}
	}
	return infos, nil
}

// schemaNamesForRoot returns the names of all database schemas in the root. For roots that don't record
// any database schemas (plain dolt databases), the default (empty) schema name is returned.
func schemaNamesForRoot(ctx context.Context, root doltdb.RootValue) ([]string, error) {
	dbSchemas, err := root.GetDatabaseSchemas(ctx)
	if err != nil {
		return nil, err
	}
	if len(dbSchemas) == 0 {
		return []string{doltdb.DefaultSchemaName}, nil
	}
	names := make([]string, len(dbSchemas))
	for i, s := range dbSchemas {
		names[i] = s.Name
	}
	return names, nil
}
