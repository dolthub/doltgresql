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
	"iter"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/google/btree"
)

// inMemIndexScanIter is a sql.RowIter that uses an in-memory btree index to satisfy index lookups
// on pg_catalog tables.
type inMemIndexScanIter[T any] struct {
	lookup         sql.IndexLookup
	rangeConverter RangeConverter[T]
	btreeAccess    BTreeStorageAccess[T]
	rowConverter   rowConverter[T]
	rangeIdx       int
	next           func() (T, bool)
	stop           func()
}

var _ sql.RowIter = (*inMemIndexScanIter[any])(nil)

// RangeConverter knows how to convert a Range to bounds for a btree scan. The two values returned are the
// greater-than-or-equal lower bound, and the less-than upper bound for this index.
type RangeConverter[T any] interface {
	getIndexScanRange(rng sql.Range, index sql.Index) (T, bool, T, bool)
}

// BTreeStorageAccess knows how to get a btree index by name. This interface needs two methods because
// unique and non-unique indexes have different types as stored in the btree package.
type BTreeStorageAccess[T any] interface {
	getIndex(name string) *inMemIndexStorage[T]
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
	if l.stop != nil {
		l.stop()
	}
	return nil
}

// nextItem returns the next item from the index lookup, or io.EOF if there are no more items.
// Needs to return a pointer to T so that we can return nil for EOF.
func (l *inMemIndexScanIter[T]) nextItem() (*T, error) {
	if l.rangeIdx >= l.lookup.Ranges.Len() {
		return nil, io.EOF
	}

	if l.next != nil {
		next, ok := l.next()
		if !ok {
			l.stop()
			l.next = nil
			l.stop = nil
			l.rangeIdx++
			return l.nextItem()
		}
		return &next, nil
	}

	inMemIndex := l.lookup.Index.(pgCatalogInMemIndex)
	rng := l.lookup.Ranges.ToRanges()[l.rangeIdx]
	gte, hasLowerBound, lt, hasUpperBound := l.rangeConverter.getIndexScanRange(rng, l.lookup.Index)
	idx := l.btreeAccess.getIndex(inMemIndex.name)
	if hasLowerBound && hasUpperBound {
		l.next, l.stop = idx.IterRange(gte, lt)
	} else if hasLowerBound {
		l.next, l.stop = idx.IterGreaterThanEqual(gte)
	} else if hasUpperBound {
		l.next, l.stop = idx.IterLessThan(lt)
	} else {
		// We don't support nil lookups for this kind of index, there are never nillable elements
		return nil, io.EOF
	}

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

var _ sql.Index = (*pgCatalogInMemIndex)(nil)

// ID implements the interface sql.Index.
func (p pgCatalogInMemIndex) ID() string {
	return p.name
}

// Database implements the interface sql.Index.
func (p pgCatalogInMemIndex) Database() string {
	return p.dbName
}

// Table implements the interface sql.Index.
func (p pgCatalogInMemIndex) Table() string {
	return p.tblName
}

// Expressions implements the interface sql.Index.
func (p pgCatalogInMemIndex) Expressions() []string {
	exprs := make([]string, len(p.columnExprs))
	for i, expr := range p.columnExprs {
		exprs[i] = expr.Expression
	}
	return exprs
}

// IsUnique implements the interface sql.Index.
func (p pgCatalogInMemIndex) IsUnique() bool {
	return p.uniq
}

// IsSpatial implements the interface sql.Index.
func (p pgCatalogInMemIndex) IsSpatial() bool {
	return false
}

// IsFullText implements the interface sql.Index.
func (p pgCatalogInMemIndex) IsFullText() bool {
	return false
}

// IsVector implements the interface sql.Index.
func (p pgCatalogInMemIndex) IsVector() bool {
	return false
}

// Comment implements the interface sql.Index.
func (p pgCatalogInMemIndex) Comment() string {
	return ""
}

// IndexType implements the interface sql.Index.
func (p pgCatalogInMemIndex) IndexType() string {
	return "BTREE"
}

// IsGenerated implements the interface sql.Index.
func (p pgCatalogInMemIndex) IsGenerated() bool {
	return false
}

// ColumnExpressionTypes implements the interface sql.Index.
func (p pgCatalogInMemIndex) ColumnExpressionTypes() []sql.ColumnExpressionType {
	return p.columnExprs
}

// CanSupport implements the interface sql.Index.
func (p pgCatalogInMemIndex) CanSupport(context *sql.Context, r ...sql.Range) bool {
	return true
}

// CanSupportOrderBy implements the interface sql.Index.
func (p pgCatalogInMemIndex) CanSupportOrderBy(expr sql.Expression) bool {
	return true
}

// PrefixLengths implements the interface sql.Index.
func (p pgCatalogInMemIndex) PrefixLengths() []uint16 {
	return make([]uint16, len(p.columnExprs))
}

var _ sql.Index = (*pgCatalogInMemIndex)(nil)

// inMemIndexPartition is a sql.Partition that represents the single partition for an in memory index lookup.
type inMemIndexPartition struct {
	idxName string
	lookup  sql.IndexLookup
}

var _ sql.Partition = (*inMemIndexPartition)(nil)

// Key implements the interface sql.Partition.
func (p inMemIndexPartition) Key() []byte {
	return []byte(p.idxName)
}

// inMemIndexPartIter is a sql.PartitionIter that returns a single partition for an in memory index lookup.
type inMemIndexPartIter struct {
	used bool
	part inMemIndexPartition
}

var _ sql.PartitionIter = (*inMemIndexPartIter)(nil)

// Close implements the interface sql.PartitionIter.
func (p inMemIndexPartIter) Close(context *sql.Context) error {
	return nil
}

// Next implements the interface sql.PartitionIter.
func (p *inMemIndexPartIter) Next(context *sql.Context) (sql.Partition, error) {
	if p.used {
		return nil, io.EOF
	}
	p.used = true
	return p.part, nil
}

// inMemIndexStorage is an in-memory storage for an index using a btree, abstracting away the differences between
// unique and non-unique indexes.
type inMemIndexStorage[T any] struct {
	uniqTree    *btree.BTreeG[T]
	nonUniqTree *btree.BTreeG[[]T]
}

// NewUniqueInMemIndexStorage creates a new in-memory index storage for a unique index.
func NewUniqueInMemIndexStorage[T any](lessFunc func(a, b T) bool) *inMemIndexStorage[T] {
	return &inMemIndexStorage[T]{
		uniqTree: btree.NewG[T](2, lessFunc),
	}
}

// NewNonUniqueInMemIndexStorage creates a new in-memory index storage for a non-unique index.
func NewNonUniqueInMemIndexStorage[T any](lessFunc func(a, b []T) bool) *inMemIndexStorage[T] {
	return &inMemIndexStorage[T]{
		nonUniqTree: btree.NewG[[]T](2, lessFunc),
	}
}

// Add adds a value to the in-memory index storage.
func (s *inMemIndexStorage[T]) Add(val T) {
	if s.uniqTree != nil {
		s.uniqTree.ReplaceOrInsert(val)
	} else {
		existing, replaced := s.nonUniqTree.ReplaceOrInsert([]T{val})
		if replaced {
			existing = append(existing, val)
			s.nonUniqTree.ReplaceOrInsert(existing)
		}
	}
}

// IterRange implements an in-order iteration over the index values in the range [gte, lt). All values in the
// index in the range are sent to the channel
func (s *inMemIndexStorage[T]) IterRange(gte, lt T) (next func() (T, bool), stop func()) {
	if s.uniqTree != nil {
		return iter.Pull(func(yield func(T) bool) {
			s.uniqTree.AscendRange(gte, lt, yield)
		})
	} else {
		next, stop := iter.Pull(func(yield func([]T) bool) {
			s.nonUniqTree.AscendRange([]T{gte}, []T{lt}, yield)
		})
		return s.unnestIter(next, stop)
	}
}

// IterGreaterThanEqual implements an in-order iteration over the index values greater than or equal to the given value.
// All values in the index greater than or equal to the given value are sent to the channel.
func (s *inMemIndexStorage[T]) IterGreaterThanEqual(gte T) (next func() (T, bool), stop func()) {
	if s.uniqTree != nil {
		return iter.Pull(func(yield func(T) bool) {
			s.uniqTree.AscendGreaterOrEqual(gte, yield)
		})
	} else {
		next, stop := iter.Pull(func(yield func([]T) bool) {
			s.nonUniqTree.AscendGreaterOrEqual([]T{gte}, yield)
		})
		return s.unnestIter(next, stop)
	}
}

// IterLessThan implements an in-order iteration over the index values less than the given value.
// All values in the index less than or equal to the given value are sent to the channel.
func (s *inMemIndexStorage[T]) IterLessThan(lt T) (next func() (T, bool), stop func()) {
	if s.uniqTree != nil {
		return iter.Pull(func(yield func(T) bool) {
			s.uniqTree.AscendLessThan(lt, yield)
		})
	} else {
		next, stop := iter.Pull(func(yield func([]T) bool) {
			s.nonUniqTree.AscendLessThan([]T{lt}, yield)
		})
		return s.unnestIter(next, stop)
	}
}

// unnestIter takes an iterator that returns slices of T, and returns an iterator that returns individual T values.
func (s *inMemIndexStorage[T]) unnestIter(sNext func() ([]T, bool), sStop func()) (next func() (T, bool), stop func()) {
	return iter.Pull(func(yield func(T) bool) {
		defer sStop()
		for {
			items, ok := sNext()
			if !ok {
				return
			}
			for _, item := range items {
				if !yield(item) {
					return
				}
			}
		}
	})
}