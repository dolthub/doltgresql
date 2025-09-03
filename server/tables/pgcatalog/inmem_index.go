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

package pgcatalog

import (
	"io"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/btree"
)

type sqlLookupIter struct {
	lookup   sql.IndexLookup
	classes  *pgClassCache
	rangeIdx int
	nextChan chan *pgClass
}

func (l *sqlLookupIter) Next(ctx *sql.Context) (sql.Row, error) {
	nextClass, err := l.NextClassItem()
	if err != nil {
		return nil, err
	}

	return pgClassToRow(*nextClass), nil
}

func (l sqlLookupIter) Close(context *sql.Context) error {
	return nil
}

func (l *sqlLookupIter) NextClassItem() (*pgClass, error) {
	if l.rangeIdx >= l.lookup.Ranges.Len() {
		return nil, io.EOF
	}

	if l.nextChan != nil {
		class, ok := <-l.nextChan
		if !ok {
			l.nextChan = nil
			l.rangeIdx++
			return l.NextClassItem()
		}
		return class, nil
	}

	l.nextChan = make(chan *pgClass)
	go func() {
		idx, gte, lte := l.getIndexScanRange()
		itr := func(item *pgClass) bool {
			l.nextChan <- item
			return true
		}

		if gte != nil && lte != nil {
			idx.AscendRange(gte, lte, itr)
		} else if gte != nil {
			idx.AscendGreaterOrEqual(gte, itr)
		} else if lte != nil {
			idx.AscendLessThan(lte, itr)
		} else {
			idx.Ascend(itr)
		}

		// because the above call uses a closed range for its upper end, we just return the last item at the end rather
		// than trying to generate a greater one
		upperRange, ok := idx.Get(lte)
		if ok {
			l.nextChan <- upperRange
		}

		close(l.nextChan)
	}()

	return l.NextClassItem()
}

func (l sqlLookupIter) getIndexScanRange() (*btree.BTreeG[*pgClass], *pgClass, *pgClass) {
	rng := l.lookup.Ranges.ToRanges()[l.rangeIdx]

	var gte, lte *pgClass
	var idx *btree.BTreeG[*pgClass]

	switch l.lookup.Index.(pgCatalogInMemIndex).name {
	case "pg_class_oid_index":
		idx = l.classes.oidIdx

		msrng := rng.(sql.MySQLRange)
		oidRng := msrng[0]
		var oidLower, oidUpper id.Id
		if oidRng.HasLowerBound() {
			oidLower = sql.GetMySQLRangeCutKey(oidRng.LowerBound).(id.Id)
			gte = &pgClass{
				oid: oidLower,
			}
		}
		if oidRng.HasUpperBound() {
			oidUpper = sql.GetMySQLRangeCutKey(oidRng.UpperBound).(id.Id)
			lte = &pgClass{
				oid: oidUpper,
			}
		}

	case "pg_class_relname_nsp_index":
		idx = l.classes.nameIdx
		msrng := rng.(sql.MySQLRange)
		relNameRange := msrng[0]
		schemaOidRange := msrng[1]
		var relnameLower, relnameUpper string
		var schemaOidLower, schemaOidUpper id.Id

		if relNameRange.HasLowerBound() {
			relnameLower = sql.GetMySQLRangeCutKey(relNameRange.LowerBound).(string)
		}
		if relNameRange.HasUpperBound() {
			relnameUpper = sql.GetMySQLRangeCutKey(relNameRange.UpperBound).(string)
		}
		if schemaOidRange.HasLowerBound() {
			lowerRangeCutKey := sql.GetMySQLRangeCutKey(schemaOidRange.LowerBound)
			schemaOidLower = id.Cache().ToInternal(lowerRangeCutKey.(id.Oid).OID())
		}
		if schemaOidRange.HasUpperBound() {
			upperRangeCutKey := sql.GetMySQLRangeCutKey(schemaOidRange.UpperBound)
			schemaOidUpper = id.Cache().ToInternal(upperRangeCutKey.(id.Oid).OID())
		}

		if relNameRange.HasLowerBound() || schemaOidRange.HasLowerBound() {
			gte = &pgClass{
				name:      relnameLower,
				schemaOid: schemaOidLower,
			}
		}

		if relNameRange.HasUpperBound() || schemaOidRange.HasUpperBound() {
			lte = &pgClass{
				name:      relnameUpper,
				schemaOid: schemaOidUpper,
			}
		}
	default:
		panic("unknown index name: " + l.lookup.Index.(pgCatalogInMemIndex).name)
	}

	return idx, gte, lte
}

type pgCatalogInMemIndex struct {
	name        string
	tblName     string
	dbName      string
	uniq        bool
	columnExprs []sql.ColumnExpressionType
}

func (p pgCatalogInMemIndex) ID() string {
	return p.name
}

func (p pgCatalogInMemIndex) Database() string {
	return p.dbName
}

func (p pgCatalogInMemIndex) Table() string {
	return p.tblName
}

func (p pgCatalogInMemIndex) Expressions() []string {
	exprs := make([]string, len(p.columnExprs))
	for i, expr := range p.columnExprs {
		exprs[i] = expr.Expression
	}
	return exprs
}

func (p pgCatalogInMemIndex) IsUnique() bool {
	return p.uniq
}

func (p pgCatalogInMemIndex) IsSpatial() bool {
	return false
}

func (p pgCatalogInMemIndex) IsFullText() bool {
	return false
}

func (p pgCatalogInMemIndex) IsVector() bool {
	return false
}

func (p pgCatalogInMemIndex) Comment() string {
	return ""
}

func (p pgCatalogInMemIndex) IndexType() string {
	return "BTREE"
}

func (p pgCatalogInMemIndex) IsGenerated() bool {
	return false
}

func (p pgCatalogInMemIndex) ColumnExpressionTypes() []sql.ColumnExpressionType {
	return p.columnExprs
}

func (p pgCatalogInMemIndex) CanSupport(context *sql.Context, r ...sql.Range) bool {
	return true // why not
}

func (p pgCatalogInMemIndex) CanSupportOrderBy(expr sql.Expression) bool {
	return true
}

func (p pgCatalogInMemIndex) PrefixLengths() []uint16 {
	return make([]uint16, len(p.columnExprs))
}

var _ sql.Index = (*pgCatalogInMemIndex)(nil)