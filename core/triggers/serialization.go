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

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Serialize returns the Trigger as a byte slice. If the Trigger is invalid, then this returns a nil slice.
func (trigger Trigger) Serialize(ctx context.Context) ([]byte, error) {
	if !trigger.ID.IsValid() {
		return nil, nil
	}

	// Initialize the writer and version
	writer := utils.NewWriter(256)
	writer.VariableUint(0) // Version
	// Write the trigger data
	writer.Id(trigger.ID.AsId())
	writer.Id(trigger.Function.AsId())
	writer.Uint8(uint8(trigger.Timing))
	writer.Bool(trigger.ForEachRow)
	// TODO: writer.Unknown(trigger.When)
	writer.Uint8(uint8(trigger.Deferrable))
	writer.Id(trigger.ReferencedTableName.AsId())
	writer.Bool(trigger.Constraint)
	writer.String(trigger.OldTransitionName)
	writer.String(trigger.NewTransitionName)
	writer.StringSlice(trigger.Arguments)
	writer.String(trigger.Definition)
	// Write the events
	writer.VariableUint(uint64(len(trigger.Events)))
	for _, event := range trigger.Events {
		writer.Uint8(uint8(event.Type))
		writer.StringSlice(event.ColumnNames)
	}
	// Returns the data
	return writer.Data(), nil
}

// DeserializeTrigger returns the Trigger that was serialized in the byte slice. Returns an empty Trigger (invalid ID)
// if data is nil or empty.
func DeserializeTrigger(ctx context.Context, data []byte) (Trigger, error) {
	if len(data) == 0 {
		return Trigger{}, nil
	}
	reader := utils.NewReader(data)
	version := reader.VariableUint()
	if version != 0 {
		return Trigger{}, errors.Errorf("version %d of triggers is not supported, please upgrade the server", version)
	}

	// Read from the reader
	t := Trigger{}
	t.ID = id.Trigger(reader.Id())
	t.Function = id.Function(reader.Id())
	t.Timing = TriggerTiming(reader.Uint8())
	t.ForEachRow = reader.Bool()
	// TODO: trigger.When = reader.Unknown()
	t.Deferrable = TriggerDeferrable(reader.Uint8())
	t.ReferencedTableName = id.Table(reader.Id())
	t.Constraint = reader.Bool()
	t.OldTransitionName = reader.String()
	t.NewTransitionName = reader.String()
	t.Arguments = reader.StringSlice()
	t.Definition = reader.String()
	// Read the events
	eventCount := reader.VariableUint()
	t.Events = make([]TriggerEvent, eventCount)
	for eventIdx := uint64(0); eventIdx < eventCount; eventIdx++ {
		t.Events[eventIdx].Type = TriggerEventType(reader.Uint8())
		t.Events[eventIdx].ColumnNames = reader.StringSlice()
	}
	if !reader.IsEmpty() {
		return Trigger{}, errors.Errorf("extra data found while deserializing a trigger")
	}
	// Return the deserialized object
	return t, nil
}
