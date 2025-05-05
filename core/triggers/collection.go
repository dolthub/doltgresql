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

package triggers

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/rootobject/objinterface"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// Collection contains a collection of triggers.
type Collection struct {
	accessCache   map[id.Trigger]Trigger    // This cache is used for general access when you know the exact ID
	tableCache    map[id.Table][]id.Trigger // This cache is used to find triggers by table
	idCache       []id.Trigger              // This cache simply contains the name of every trigger
	mapHash       hash.Hash                 // This is cached so that we don't have to calculate the hash every time
	underlyingMap prolly.AddressMap
	ns            tree.NodeStore
	isReadOnly    bool
}

// TriggerTiming specifies the timing of the trigger's execution.
type TriggerTiming uint8

const (
	TriggerTiming_Before    TriggerTiming = 0
	TriggerTiming_After     TriggerTiming = 1
	TriggerTiming_InsteadOf TriggerTiming = 2
)

// TriggerDeferrable specifies whether the trigger is deferrable.
type TriggerDeferrable uint8

const (
	TriggerDeferrable_NotDeferrable       TriggerDeferrable = 0 // NOT DEFERRABLE
	TriggerDeferrable_DeferrableImmediate TriggerDeferrable = 1 // DEFERRABLE INITIALLY IMMEDIATE
	TriggerDeferrable_DeferrableDeferred  TriggerDeferrable = 2 // DEFERRABLE INITIALLY DEFERRED
)

// TriggerEventType specifies which type of event that the trigger applies to.
type TriggerEventType uint8

const (
	TriggerEventType_Insert   TriggerEventType = 0
	TriggerEventType_Update   TriggerEventType = 1
	TriggerEventType_Delete   TriggerEventType = 2
	TriggerEventType_Truncate TriggerEventType = 3
)

// TriggerEvent specifies the event type, along with column information for update events.
type TriggerEvent struct {
	Type        TriggerEventType
	ColumnNames []string
}

// Trigger represents a trigger.
type Trigger struct {
	ID                  id.Trigger
	Function            id.Function
	Timing              TriggerTiming
	Events              []TriggerEvent
	ForEachRow          bool // When false, represents FOR EACH STATEMENT
	When                []plpgsql.InterpreterOperation
	Deferrable          TriggerDeferrable
	ReferencedTableName id.Table // FROM referenced_table_name
	Constraint          bool
	OldTransitionName   string // REFERENCING OLD TABLE AS transition_relation_name
	NewTransitionName   string // REFERENCING NEW TABLE AS transition_relation_name
	Arguments           []string
	Definition          string
}

var _ objinterface.Collection = (*Collection)(nil)
var _ objinterface.RootObject = Trigger{}

// NewCollection returns a new Collection.
func NewCollection(ctx context.Context, underlyingMap prolly.AddressMap, ns tree.NodeStore) (*Collection, error) {
	collection := &Collection{
		accessCache:   make(map[id.Trigger]Trigger),
		tableCache:    make(map[id.Table][]id.Trigger),
		idCache:       nil,
		mapHash:       hash.Hash{},
		underlyingMap: underlyingMap,
		ns:            ns,
		isReadOnly:    false,
	}
	return collection, collection.reloadCaches(ctx)
}

// GetTrigger returns the trigger with the given ID. Returns a trigger with an invalid ID if it cannot be found
// (Trigger.ID.IsValid() == false).
func (pgt *Collection) GetTrigger(ctx context.Context, trigID id.Trigger) (Trigger, error) {
	if f, ok := pgt.accessCache[trigID]; ok {
		return f, nil
	}
	return Trigger{}, nil
}

// GetTriggerIDsForTable returns the trigger IDs for the given table.
func (pgt *Collection) GetTriggerIDsForTable(ctx context.Context, tableID id.Table) []id.Trigger {
	return pgt.tableCache[tableID]
}

// GetTriggersForTable returns the triggers for the given table.
func (pgt *Collection) GetTriggersForTable(ctx context.Context, tableID id.Table) []Trigger {
	triggerIDs := pgt.tableCache[tableID]
	triggers := make([]Trigger, len(triggerIDs))
	for i, trigID := range triggerIDs {
		triggers[i] = pgt.accessCache[trigID]
	}
	return triggers
}

// GetTriggersForTableByTiming returns the triggers for the given table, all matching the given timing. These triggers
// are also sorted by their name ascending.
func (pgt *Collection) GetTriggersForTableByTiming(ctx context.Context, tableID id.Table, timing TriggerTiming) []Trigger {
	triggers := pgt.GetTriggersForTable(ctx, tableID)
	timingTriggers := make([]Trigger, 0, len(triggers))
	for _, trig := range triggers {
		if trig.Timing == timing {
			timingTriggers = append(timingTriggers, trig)
		}
	}
	sort.Slice(timingTriggers, func(i, j int) bool {
		return timingTriggers[i].Name().String() < timingTriggers[j].Name().String()
	})
	return timingTriggers
}

// HasTrigger returns whether the trigger is present.
func (pgt *Collection) HasTrigger(ctx context.Context, trigID id.Trigger) bool {
	_, ok := pgt.accessCache[trigID]
	return ok
}

// AddTrigger adds a new trigger.
func (pgt *Collection) AddTrigger(ctx context.Context, t Trigger) error {
	if pgt.isReadOnly {
		return errors.New("cannot modify a read-only collection")
	}

	// First we'll check to see if it exists
	if _, ok := pgt.accessCache[t.ID]; ok {
		return errors.Errorf(`trigger "%s" for relation "%s" already exists`, t.ID.TriggerName(), t.ID.TableName())
	}

	// Now we'll add the trigger to our map
	data, err := t.Serialize(ctx)
	if err != nil {
		return err
	}
	h, err := pgt.ns.WriteBytes(ctx, data)
	if err != nil {
		return err
	}
	mapEditor := pgt.underlyingMap.Editor()
	if err = mapEditor.Add(ctx, string(t.ID), h); err != nil {
		return err
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgt.underlyingMap = newMap
	pgt.mapHash = pgt.underlyingMap.HashOf()
	return pgt.reloadCaches(ctx)
}

// DropTrigger drops an existing trigger.
func (pgt *Collection) DropTrigger(ctx context.Context, trigIDs ...id.Trigger) error {
	if pgt.isReadOnly {
		return errors.New("cannot modify a read-only collection")
	}
	if len(trigIDs) == 0 {
		return nil
	}
	// Check that each name exists before performing any deletions
	for _, trigID := range trigIDs {
		if _, ok := pgt.accessCache[trigID]; !ok {
			return errors.Errorf(`trigger "%s" for table "%s" does not exist`, trigID.TriggerName(), trigID.TableName())
		}
	}

	// Now we'll remove the triggers from the map
	mapEditor := pgt.underlyingMap.Editor()
	for _, trigID := range trigIDs {
		err := mapEditor.Delete(ctx, string(trigID))
		if err != nil {
			return err
		}
	}
	newMap, err := mapEditor.Flush(ctx)
	if err != nil {
		return err
	}
	pgt.underlyingMap = newMap
	pgt.mapHash = pgt.underlyingMap.HashOf()
	return pgt.reloadCaches(ctx)
}

// resolveName returns the fully resolved name of the given trigger. Returns an error if the name is ambiguous.
func (pgt *Collection) resolveName(ctx context.Context, schemaName string, formattedName string) (id.Trigger, error) {
	if len(pgt.accessCache) == 0 || len(formattedName) == 0 {
		return id.NullTrigger, nil
	}

	// Check for an exact match
	fullID := pgt.tableNameToID(schemaName, formattedName)
	if _, ok := pgt.accessCache[fullID]; ok {
		return fullID, nil
	}
	tableName := fullID.TableName()
	triggerName := fullID.TriggerName()

	// Otherwise we'll iterate over all the names
	var resolvedID id.Trigger
	for _, trigID := range pgt.idCache {
		if !strings.EqualFold(triggerName, trigID.TriggerName()) || !strings.EqualFold(tableName, trigID.TableName()) {
			continue
		}
		if len(schemaName) > 0 && !strings.EqualFold(schemaName, trigID.SchemaName()) {
			continue
		}
		// The above matches, so this counts as a match
		if resolvedID.IsValid() {
			trigTableName := TriggerIDToTableName(trigID)
			resolvedTableName := TriggerIDToTableName(resolvedID)
			return id.NullTrigger, fmt.Errorf("`%s.%s` is ambiguous, matches `%s` and `%s`",
				schemaName, formattedName, trigTableName.String(), resolvedTableName.String())
		}
		resolvedID = trigID
	}
	return resolvedID, nil
}

// iterateIDs iterates over all trigger IDs in the collection.
func (pgt *Collection) iterateIDs(ctx context.Context, callback func(trigID id.Trigger) (stop bool, err error)) error {
	for _, trigID := range pgt.idCache {
		stop, err := callback(trigID)
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// IterateTriggers iterates over all triggers in the collection.
func (pgt *Collection) IterateTriggers(ctx context.Context, callback func(t Trigger) (stop bool, err error)) error {
	for _, trigID := range pgt.idCache {
		stop, err := callback(pgt.accessCache[trigID])
		if err != nil {
			return err
		} else if stop {
			return nil
		}
	}
	return nil
}

// Clone returns a new *Collection with the same contents as the original. The returned collection will never be
// read-only.
func (pgt *Collection) Clone(ctx context.Context) *Collection {
	// We don't need to clone or copy the internal caches, as they're always rebuilt and therefore never modified
	return &Collection{
		accessCache:   pgt.accessCache,
		tableCache:    pgt.tableCache,
		idCache:       pgt.idCache,
		underlyingMap: pgt.underlyingMap,
		mapHash:       pgt.mapHash,
		ns:            pgt.ns,
		isReadOnly:    false,
	}
}

// Map writes any cached sequences to the underlying map, and then returns the underlying map.
func (pgt *Collection) Map(ctx context.Context) (prolly.AddressMap, error) {
	return pgt.underlyingMap, nil
}

// DiffersFrom returns true when the hash that is associated with the underlying map for this collection is different
// from the hash in the given root.
func (pgt *Collection) DiffersFrom(ctx context.Context, root objinterface.RootValue) bool {
	hashOnGivenRoot, err := pgt.LoadCollectionHash(ctx, root)
	if err != nil {
		return true
	}
	if pgt.mapHash.Equal(hashOnGivenRoot) {
		return false
	}
	// An empty map should match an uninitialized collection on the root
	count, err := pgt.underlyingMap.Count()
	if err == nil && count == 0 && hashOnGivenRoot.IsEmpty() {
		return false
	}
	return true
}

// reloadCaches writes the underlying map's contents to the caches.
func (pgt *Collection) reloadCaches(ctx context.Context) error {
	count, err := pgt.underlyingMap.Count()
	if err != nil {
		return err
	}

	pgt.accessCache = make(map[id.Trigger]Trigger, count)
	pgt.tableCache = make(map[id.Table][]id.Trigger, count)
	pgt.mapHash = pgt.underlyingMap.HashOf()
	pgt.idCache = make([]id.Trigger, 0, count)

	return pgt.underlyingMap.IterAll(ctx, func(_ string, h hash.Hash) error {
		if h.IsEmpty() {
			return nil
		}
		data, err := pgt.ns.ReadBytes(ctx, h)
		if err != nil {
			return err
		}
		t, err := DeserializeTrigger(ctx, data)
		if err != nil {
			return err
		}
		pgt.accessCache[t.ID] = t
		tableID := id.NewTable(t.ID.SchemaName(), t.ID.TableName())
		pgt.tableCache[tableID] = append(pgt.tableCache[tableID], t.ID)
		pgt.idCache = append(pgt.idCache, t.ID)
		return nil
	})
}

// tableNameToID returns the ID that was encoded via the Name() call, as the returned TableName contains additional
// information (which this is able to process).
func (pgt *Collection) tableNameToID(schemaName string, formattedName string) id.Trigger {
	names := strings.Split(formattedName, ".")
	if len(names) != 2 {
		return id.NullTrigger
	}
	return id.NewTrigger(schemaName, names[0], names[1])
}

// GetID implements the interface objinterface.RootObject.
func (trigger Trigger) GetID() objinterface.RootObjectID {
	return objinterface.RootObjectID_Triggers
}

// HashOf implements the interface objinterface.RootObject.
func (trigger Trigger) HashOf(ctx context.Context) (hash.Hash, error) {
	data, err := trigger.Serialize(ctx)
	if err != nil {
		return hash.Hash{}, err
	}
	return hash.Of(data), nil
}

// Name implements the interface objinterface.RootObject.
func (trigger Trigger) Name() doltdb.TableName {
	return TriggerIDToTableName(trigger.ID)
}

// TriggerIDToTableName returns the ID in a format that's better for user consumption.
func TriggerIDToTableName(trigID id.Trigger) doltdb.TableName {
	return doltdb.TableName{
		Name:   fmt.Sprintf("%s.%s", trigID.TableName(), trigID.TriggerName()),
		Schema: trigID.SchemaName(),
	}
}
