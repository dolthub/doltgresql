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

package tree

//CREATE [ OR REPLACE ] [ CONSTRAINT ] TRIGGER name { BEFORE | AFTER | INSTEAD OF } { event [ OR ... ] }
//    ON table_name
//    [ FROM referenced_table_name ]
//    [ NOT DEFERRABLE | [ DEFERRABLE ] [ INITIALLY IMMEDIATE | INITIALLY DEFERRED ] ]
//    [ REFERENCING { { OLD | NEW } TABLE [ AS ] transition_relation_name } [ ... ] ]
//    [ FOR [ EACH ] { ROW | STATEMENT } ]
//    [ WHEN ( condition ) ]
//    EXECUTE { FUNCTION | PROCEDURE } function_name ( arguments )
//
//where event can be one of:
//
//    INSERT
//    UPDATE [ OF column_name [, ... ] ]
//    DELETE
//    TRUNCATE

var _ Statement = &CreateTrigger{}

// CreateTrigger represents a CREATE TRIGGER statement.
type CreateTrigger struct {
	Replace    bool
	Constraint bool
	Name       Name
	Time       TriggerTime
	Events     TriggerEvents
	OnTable    TableName
	RefTable   Name
	Deferrable TriggerDeferrableMode
	Relations  TriggerRelations
	ForEachRow bool
	When       Expr
	FuncName   *UnresolvedObjectName
	Args       NameList
}

func (node *CreateTrigger) Format(ctx *FmtCtx) {
	//TODO implement me
	panic("implement me")
}

type TriggerTime int

const (
	TriggerTimeBefore TriggerTime = iota
	TriggerTimeAfter
	TriggerTimeInsteadOf
)

type TriggerEvents []TriggerEvent

type TriggerEvent struct {
	Type TriggerEventType
	Cols NameList // used only for UPDATE event type
}

func (node TriggerEvents) Format(ctx *FmtCtx) {
	for i, event := range node {
		if i != 0 {
			ctx.WriteString(" OR ")
		}
		switch event.Type {
		case TriggerEventInsert:
			ctx.WriteString("INSERT")
		case TriggerEventUpdate:
			ctx.WriteString("UPDATE")
			if event.Cols != nil {
				ctx.WriteString(" OF ")
				ctx.FormatNode(&event.Cols)
			}
		case TriggerEventDelete:
			ctx.WriteString("DELETE")
		case TriggerEventTruncate:
			ctx.WriteString("TRUNCATE")
		}
	}
}

type TriggerEventType int

const (
	TriggerEventInsert TriggerEventType = iota
	TriggerEventUpdate
	TriggerEventDelete
	TriggerEventTruncate
)

type TriggerRelations []TriggerRelation

// TriggerRelation describes { { OLD | NEW } TABLE [ AS ] transition_relation_name }
type TriggerRelation struct {
	IsOld bool
	Name  string
}

func (t TriggerRelations) Format(ctx *FmtCtx) {
	for i, rel := range t {
		if i != 0 {
			ctx.WriteString(", ")
		}
		if rel.IsOld {
			ctx.WriteString("OLD")
		} else {
			ctx.WriteString("NEW")
		}
		ctx.WriteString(" TABLE AS ")
		ctx.WriteString(rel.Name)
	}
}

type TriggerDeferrableMode int

const (
	// TriggerDeferrable is used for { DEFERRABLE [ INITIALLY IMMEDIATE ] }
	TriggerDeferrable TriggerDeferrableMode = iota
	// TriggerNotDeferrable is used as default when not specified and for { NOT DEFERRABLE }
	// Validation cannot be specified as it's always INITIALLY IMMEDIATE for not deferrable constraint.
	TriggerNotDeferrable
	// TriggerInitiallyDeferred is used for { [ DEFERRABLE ] INITIALLY DEFERRED }
	TriggerInitiallyDeferred
	// NOTE: cases other than these are invalid.
)
