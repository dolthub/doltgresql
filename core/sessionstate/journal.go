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

package sessionstate

// Journal records session and transaction-local changes to one value. T must
// be a value (or treated as immutable): snapshots are copied by assignment.
// It is independent of SQL, so settings can use the same journal per value.
type Journal[T any] struct {
	current    T
	base       T
	inTx       bool
	changes    []journalChange[T]
	savepoints []journalSavepoint
}

type journalChange[T any] struct {
	value T
	local bool
}

type journalSavepoint struct {
	name string
	at   int
}

func NewJournal[T any](value T) Journal[T] { return Journal[T]{current: value} }

func (j *Journal[T]) Current() *T         { return &j.current }
func (j *Journal[T]) InTransaction() bool { return j.inTx }

// Begin snapshots the session value. A duplicate start leaves the active
// transaction alone; a completed transaction may be followed by another.
func (j *Journal[T]) Begin() {
	if j.inTx {
		return
	}
	j.base = j.current
	j.inTx = true
	j.changes = nil
	j.savepoints = nil
}

// SetSession survives a successful commit, including when a preceding LOCAL
// change exists. Outside a transaction it takes effect immediately.
func (j *Journal[T]) SetSession(value T) {
	if j.inTx {
		j.changes = append(j.changes, journalChange[T]{value: value})
	}
	j.current = value
}

// SetLocal lasts to transaction end. PostgreSQL ignores it outside a block.
func (j *Journal[T]) SetLocal(value T) {
	if !j.inTx {
		return
	}
	j.changes = append(j.changes, journalChange[T]{value: value, local: true})
	j.current = value
}

func (j *Journal[T]) Commit() {
	if !j.inTx {
		return
	}
	j.current = j.sessionValue()
	j.end()
}

func (j *Journal[T]) Rollback() {
	if !j.inTx {
		return
	}
	j.current = j.base
	j.end()
}

func (j *Journal[T]) end() {
	j.inTx = false
	j.changes = nil
	j.savepoints = nil
	var zero T
	j.base = zero
}

func (j *Journal[T]) Savepoint(name string) {
	if j.inTx {
		j.savepoints = append(j.savepoints, journalSavepoint{name: name, at: len(j.changes)})
	}
}

// RollbackTo retains the named savepoint, so it can be used again. Names may
// repeat; the most recent matching savepoint wins.
func (j *Journal[T]) RollbackTo(name string) bool {
	for i := len(j.savepoints) - 1; j.inTx && i >= 0; i-- {
		if j.savepoints[i].name == name {
			j.changes = j.changes[:j.savepoints[i].at]
			j.savepoints = j.savepoints[:i+1]
			j.replay()
			return true
		}
	}
	return false
}

// Release merges the named savepoint and its children into the parent scope.
func (j *Journal[T]) Release(name string) bool {
	for i := len(j.savepoints) - 1; j.inTx && i >= 0; i-- {
		if j.savepoints[i].name == name {
			j.savepoints = j.savepoints[:i]
			return true
		}
	}
	return false
}

func (j *Journal[T]) replay() {
	j.current = j.base
	for _, change := range j.changes {
		j.current = change.value
	}
}

func (j *Journal[T]) sessionValue() T {
	value := j.base
	for _, change := range j.changes {
		if !change.local {
			value = change.value
		}
	}
	return value
}
