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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/btree"
)

// inMemIndexScanIter is a sql.RowIter that uses an in-memory btree index to satisfy index lookups
// on pg_catalog tables.
type inMemIndexScanIter[T any] struct {
	lookup         sql.IndexLookup
	rangeConverter RangeConverter[T]
	btreeAccess    BTreeIndexAccess[T]
	rowConverter   rowConverter[T]
	rangeIdx       int
	nextChan       chan T
}

var _ sql.RowIter = (*inMemIndexScanIter[any])(nil)

// RangeConverter knows how to convert a Range to bounds for a btree scan.
type RangeConverter[T any] interface {
	getIndexScanRange(rng sql.Range, index sql.Index) (T, bool, T, bool)
}

// BTreeIndexAccess knows how to get a btree index by name.
type BTreeIndexAccess[T any] interface {
	getIndex(name string) *btree.BTreeG[T]
}

// rowConverter converts a value of type T to a sql.Row.
type rowConverter[T any] func(T) sql.Row

// Next implements the sql.RowIter interface.
func (l *inMemIndexScanIter[T]) Next(ctx *sql.Context) (sql.Row, error) {
	nextClass, err := l.nextItem()
	if err != nil {
		return nil, err
	}

	return l.rowConverter(*nextClass), nil
}

// Close implements the sql.RowIter interface.
func (l *inMemIndexScanIter[T]) Close(ctx *sql.Context) error {
	return nil
}

// nextItem returns the next item from the index lookup, or io.EOF if there are no more items.
// Needs to return a pointer to T so that we can return nil for EOF.
func (l *inMemIndexScanIter[T]) nextItem() (*T, error) {
	if l.rangeIdx >= l.lookup.Ranges.Len() {
		return nil, io.EOF
	}

	if l.nextChan != nil {
		next, ok := <-l.nextChan
		if !ok {
			l.nextChan = nil
			l.rangeIdx++
			return l.nextItem()
		}
		return &next, nil
	}

	l.nextChan = make(chan T)
	rng := l.lookup.Ranges.ToRanges()[l.rangeIdx]
	go func() {
		gte, hasLowerBound, lte, hasUpperBound := l.rangeConverter.getIndexScanRange(rng, l.lookup.Index)
		itr := func(item T) bool {
			l.nextChan <- item
			return true
		}

		idx := l.btreeAccess.getIndex(l.lookup.Index.(pgCatalogInMemIndex).name)
		if hasLowerBound && hasUpperBound {
			idx.AscendRange(gte, lte, itr)
		} else if hasLowerBound {
			idx.AscendGreaterOrEqual(gte, itr)
		} else if hasUpperBound {
			idx.AscendLessThan(lte, itr)
		} else {
			idx.Ascend(itr)
		}

		// because the above call uses a closed range for its upper end, we just return the last item at the end rather
		// than trying to generate a greater one for the upper bound.
		upperRange, ok := idx.Get(lte)
		if ok {
			l.nextChan <- upperRange
		}

		close(l.nextChan)
	}()

	return l.nextItem()
}

// pgCatalogInMemIndex is an in-memory implementation of sql.Index for pg_catalog tables.
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
	return true
}

func (p pgCatalogInMemIndex) CanSupportOrderBy(expr sql.Expression) bool {
	return true
}

func (p pgCatalogInMemIndex) PrefixLengths() []uint16 {
	return make([]uint16, len(p.columnExprs))
}

var _ sql.Index = (*pgCatalogInMemIndex)(nil)

type inMemIndexPartition struct {
	idxName string
	lookup  sql.IndexLookup
}

func (p inMemIndexPartition) Key() []byte {
	return []byte(p.idxName)
}

var _ sql.Partition = (*inMemIndexPartition)(nil)

type inMemIndexPartIter struct {
	used bool
	part inMemIndexPartition
}

func (p inMemIndexPartIter) Close(context *sql.Context) error {
	return nil
}

func (p *inMemIndexPartIter) Next(context *sql.Context) (sql.Partition, error) {
	if p.used {
		return nil, io.EOF
	}
	p.used = true
	return p.part, nil
}

var _ sql.PartitionIter = (*inMemIndexPartIter)(nil)
