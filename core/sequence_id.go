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

package core

import (
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// sequenceIDValidator is the internal ID validator for sequences.
func sequenceIDValidator(ctx *sql.Context, operation id.Operation, originalID id.Internal, newID id.Internal) error {
	switch originalID.Section() {
	case id.Section_ColumnDefault, id.Section_Table:
		switch operation {
		case id.Operation_Rename, id.Operation_Delete, id.Operation_Delete_Cascade:
			return nil
		default:
			return fmt.Errorf("sequence performer received unexpected operation `%s`", operation.String())
		}
	default:
		return fmt.Errorf("sequence performer received unexpected section `%s`", originalID.Section().String())
	}
}

// sequenceIDValidator is the internal ID performer for sequences, which modifies the sequence collection in response to
// the given operation and section of the original ID.
func sequenceIDPerformer(ctx *sql.Context, operation id.Operation, originalID id.Internal, newID id.Internal) error {
	switch originalID.Section() {
	case id.Section_ColumnDefault:
		originalIDCol := id.InternalColumnDefault(originalID)
		switch operation {
		case id.Operation_Rename:
			return nil
		case id.Operation_Delete, id.Operation_Delete_Cascade:
			collection, err := GetSequencesCollectionFromContext(ctx)
			if err != nil {
				return err
			}
			sequences := collection.GetSequencesWithTable(doltdb.TableName{
				Name:   originalIDCol.TableName(),
				Schema: originalIDCol.SchemaName(),
			})
			for _, sequence := range sequences {
				if sequence.OwnerColumn == originalIDCol.ColumnName() {
					if err = collection.DropSequence(sequence.Name); err != nil {
						return err
					}
				}
			}
			return nil
		default:
			return fmt.Errorf("sequence performer received unexpected operation `%s`", operation.String())
		}
	case id.Section_Table:
		originalIDTable := id.InternalTable(originalID)
		switch operation {
		case id.Operation_Rename, id.Operation_Delete, id.Operation_Delete_Cascade:
			collection, err := GetSequencesCollectionFromContext(ctx)
			if err != nil {
				return err
			}
			sequences := collection.GetSequencesWithTable(doltdb.TableName{
				Name:   originalIDTable.TableName(),
				Schema: originalIDTable.SchemaName(),
			})
			for _, sequence := range sequences {
				if err = collection.DropSequence(sequence.Name); err != nil {
					return err
				}
				if operation == id.Operation_Rename {
					sequence.OwnerTable = id.InternalTable(newID)
					if err = collection.CreateSequence(sequence.Name.SchemaName(), sequence); err != nil {
						return err
					}
				}
			}
			return nil
		default:
			return fmt.Errorf("sequence performer received unexpected operation `%s`", operation.String())
		}
	default:
		return fmt.Errorf("sequence performer received unexpected section `%s`", originalID.Section().String())
	}
}
