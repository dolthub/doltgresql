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

package core

import (
	"context"
	"fmt"

	doltserial "github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/schema"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/shim"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"

	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// rootStorage is the FlatBuffer interface for the storage format.
type rootStorage struct {
	srv *serial.RootValue
}

// SetFunctions sets the function hash and returns a new storage object.
func (r rootStorage) SetFunctions(ctx context.Context, h hash.Hash) (rootStorage, error) {
	if len(r.srv.FunctionsBytes()) > 0 {
		ret := r.clone()
		copy(ret.srv.FunctionsBytes(), h[:])
		return ret, nil
	} else {
		return r.clone(), nil
	}
}

// SetSequences sets the sequence hash and returns a new storage object.
func (r rootStorage) SetSequences(ctx context.Context, h hash.Hash) (rootStorage, error) {
	if len(r.srv.SequencesBytes()) > 0 {
		ret := r.clone()
		copy(ret.srv.SequencesBytes(), h[:])
		return ret, nil
	} else {
		dbSchemas, err := r.GetSchemas(ctx)
		if err != nil {
			return rootStorage{}, err
		}
		msg, err := r.serializeRootValue(r.srv.TablesBytes(), dbSchemas, h[:])
		if err != nil {
			return rootStorage{}, err
		}
		return rootStorage{msg}, nil
	}
}

// SetForeignKeyMap sets the foreign key and returns a new storage object.
func (r rootStorage) SetForeignKeyMap(ctx context.Context, vrw types.ValueReadWriter, v types.Value) (rootStorage, error) {
	var h hash.Hash
	isempty, err := doltdb.EmptyForeignKeyCollection(v.(types.SerialMessage))
	if err != nil {
		return rootStorage{}, err
	}
	if !isempty {
		ref, err := vrw.WriteValue(ctx, v)
		if err != nil {
			return rootStorage{}, err
		}
		h = ref.TargetHash()
	}
	ret := r.clone()
	copy(ret.srv.ForeignKeyAddrBytes(), h[:])
	return ret, nil
}

// SetFeatureVersion sets the feature version and returns a new storage object.
func (r rootStorage) SetFeatureVersion(v doltdb.FeatureVersion) (rootStorage, error) {
	ret := r.clone()
	ret.srv.MutateFeatureVersion(int64(v))
	return ret, nil
}

// SetCollation sets the collation and returns a new storage object.
func (r rootStorage) SetCollation(ctx context.Context, collation schema.Collation) (rootStorage, error) {
	ret := r.clone()
	ret.srv.MutateCollation(serial.Collation(collation))
	return ret, nil
}

// GetSchemas returns all schemas.
func (r rootStorage) GetSchemas(ctx context.Context) ([]schema.DatabaseSchema, error) {
	numSchemas := r.srv.SchemasLength()
	schemas := make([]schema.DatabaseSchema, numSchemas)
	for i := 0; i < numSchemas; i++ {
		dbSchema := new(serial.DatabaseSchema)
		_, err := r.srv.TrySchemas(dbSchema, i)
		if err != nil {
			return nil, err
		}

		schemas[i] = schema.DatabaseSchema{
			Name: string(dbSchema.Name()),
		}
	}

	return schemas, nil
}

// SetSchemas sets the given schemas and returns a new storage object.
func (r rootStorage) SetSchemas(ctx context.Context, dbSchemas []schema.DatabaseSchema) (rootStorage, error) {
	msg, err := r.serializeRootValue(r.srv.TablesBytes(), dbSchemas, r.srv.SequencesBytes())
	if err != nil {
		return rootStorage{}, err
	}
	return rootStorage{msg}, nil
}

// GetFunctions returns the functions hash.
func (r rootStorage) GetFunctions() hash.Hash {
	hashBytes := r.srv.FunctionsBytes()
	if len(hashBytes) == 0 {
		return hash.Hash{}
	}
	return hash.New(hashBytes)
}

// GetSequences returns the sequence hash.
func (r rootStorage) GetSequences() hash.Hash {
	hashBytes := r.srv.SequencesBytes()
	if len(hashBytes) == 0 {
		return hash.Hash{}
	}
	return hash.New(hashBytes)
}

// GetTypes returns the domain hash.
func (r rootStorage) GetTypes() hash.Hash {
	hashBytes := r.srv.TypesBytes()
	if len(hashBytes) == 0 {
		return hash.Hash{}
	}
	return hash.New(hashBytes)
}

// SetTypes sets the domain hash and returns a new storage object.
func (r rootStorage) SetTypes(ctx context.Context, h hash.Hash) (rootStorage, error) {
	if len(r.srv.TypesBytes()) > 0 {
		ret := r.clone()
		copy(ret.srv.TypesBytes(), h[:])
		return ret, nil
	} else {
		dbSchemas, err := r.GetSchemas(ctx)
		if err != nil {
			return rootStorage{}, err
		}
		msg, err := r.serializeRootValue(r.srv.TablesBytes(), dbSchemas, h[:])
		if err != nil {
			return rootStorage{}, err
		}
		return rootStorage{msg}, nil
	}
}

// clone returns a clone of the calling storage.
func (r rootStorage) clone() rootStorage {
	bs := make([]byte, len(r.srv.Table().Bytes))
	copy(bs, r.srv.Table().Bytes)
	var ret serial.RootValue
	ret.Init(bs, r.srv.Table().Pos)
	return rootStorage{&ret}
}

// DebugString returns the storage as a printable string.
func (r rootStorage) DebugString(ctx context.Context) string {
	return fmt.Sprintf("rootStorage[%d, %s, %s]",
		r.srv.FeatureVersion(),
		"...",
		hash.New(r.srv.ForeignKeyAddrBytes()).String())
}

// nomsValue returns the storage as a noms value.
func (r rootStorage) nomsValue() types.Value {
	return types.SerialMessage(r.srv.Table().Bytes)
}

// GetFeatureVersion returns the feature version for this storage object.
func (r rootStorage) GetFeatureVersion() doltdb.FeatureVersion {
	return doltdb.FeatureVersion(r.srv.FeatureVersion())
}

// getAddressMap returns the address map from within this storage object.
func (r rootStorage) getAddressMap(vrw types.ValueReadWriter, ns tree.NodeStore) (prolly.AddressMap, error) {
	tbytes := r.srv.TablesBytes()
	node, _, err := shim.NodeFromValue(types.SerialMessage(tbytes))
	if err != nil {
		return prolly.AddressMap{}, err
	}
	return prolly.NewAddressMap(node, ns)
}

// GetTablesMap returns the tables map from within this storage object.
func (r rootStorage) GetTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, databaseSchema string) (rootTableMap, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return rootTableMap{}, err
	}
	return rootTableMap{AddressMap: am, schemaName: databaseSchema}, nil
}

// GetForeignKeys returns the types.SerialMessage representing the foreign keys.
func (r rootStorage) GetForeignKeys(ctx context.Context, vr types.ValueReader) (types.Value, bool, error) {
	addr := hash.New(r.srv.ForeignKeyAddrBytes())
	if addr.IsEmpty() {
		return types.SerialMessage{}, false, nil
	}
	v, err := vr.ReadValue(ctx, addr)
	if err != nil {
		return types.SerialMessage{}, false, err
	}
	return v.(types.SerialMessage), true, nil
}

// GetCollation returns the collation declared within storage.
func (r rootStorage) GetCollation(ctx context.Context) (schema.Collation, error) {
	collation := r.srv.Collation()
	// Pre-existing repositories will return invalid here
	if collation == serial.Collationinvalid {
		return schema.Collation_Default, nil
	}
	return schema.Collation(collation), nil
}

// EditTablesMap edits the table map within storage.
func (r rootStorage) EditTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, edits []tableEdit) (rootStorage, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return rootStorage{}, err
	}
	ae := am.Editor()
	for _, e := range edits {
		if e.old_name.Name != "" {
			oldaddr, err := am.Get(ctx, encodeTableNameForAddressMap(e.old_name))
			if err != nil {
				return rootStorage{}, err
			}
			newaddr, err := am.Get(ctx, encodeTableNameForAddressMap(e.name))
			if err != nil {
				return rootStorage{}, err
			}
			if oldaddr.IsEmpty() {
				return rootStorage{}, doltdb.ErrTableNotFound
			}
			if !newaddr.IsEmpty() {
				return rootStorage{}, doltdb.ErrTableExists
			}
			err = ae.Delete(ctx, encodeTableNameForAddressMap(e.old_name))
			if err != nil {
				return rootStorage{}, err
			}
			err = ae.Update(ctx, encodeTableNameForAddressMap(e.name), oldaddr)
			if err != nil {
				return rootStorage{}, err
			}
		} else {
			if e.ref == nil {
				err := ae.Delete(ctx, encodeTableNameForAddressMap(e.name))
				if err != nil {
					return rootStorage{}, err
				}
			} else {
				err := ae.Update(ctx, encodeTableNameForAddressMap(e.name), e.ref.TargetHash())
				if err != nil {
					return rootStorage{}, err
				}
			}
		}
	}
	am, err = ae.Flush(ctx)
	if err != nil {
		return rootStorage{}, err
	}

	ambytes := []byte(tree.ValueFromNode(am.Node()).(types.SerialMessage))
	dbSchemas, err := r.GetSchemas(ctx)
	if err != nil {
		return rootStorage{}, err
	}

	msg, err := r.serializeRootValue(ambytes, dbSchemas, r.srv.SequencesBytes())
	if err != nil {
		return rootStorage{}, err
	}
	return rootStorage{msg}, nil
}

// serializeRootValue serializes a new serial.RootValue object.
func (r rootStorage) serializeRootValue(addressMapBytes []byte, dbSchemas []schema.DatabaseSchema, seqHash []byte) (*serial.RootValue, error) {
	builder := flatbuffers.NewBuilder(80)
	tablesOffset := builder.CreateByteVector(addressMapBytes)
	schemasOffset := serializeDatabaseSchemas(builder, dbSchemas)
	fkOffset := builder.CreateByteVector(r.srv.ForeignKeyAddrBytes())
	seqOffset := builder.CreateByteVector(seqHash)

	serial.RootValueStart(builder)
	serial.RootValueAddFeatureVersion(builder, r.srv.FeatureVersion())
	serial.RootValueAddCollation(builder, r.srv.Collation())
	serial.RootValueAddTables(builder, tablesOffset)
	serial.RootValueAddForeignKeyAddr(builder, fkOffset)
	serial.RootValueAddSequences(builder, seqOffset)
	if schemasOffset > 0 {
		serial.RootValueAddSchemas(builder, schemasOffset)
	}

	bs := doltserial.FinishMessage(builder, serial.RootValueEnd(builder), []byte(doltserial.DoltgresRootValueFileID))
	msg, err := serial.TryGetRootAsRootValue(bs, doltserial.MessagePrefixSz)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

// serializeDatabaseSchemas serialzes the schemas into an offset within the given builder.
func serializeDatabaseSchemas(b *flatbuffers.Builder, dbSchemas []schema.DatabaseSchema) flatbuffers.UOffsetT {
	// if we have no schemas, do not serialize an empty vector
	if len(dbSchemas) == 0 {
		return 0
	}

	offsets := make([]flatbuffers.UOffsetT, len(dbSchemas))
	for i := len(dbSchemas) - 1; i >= 0; i-- {
		dbSchema := dbSchemas[i]

		nameOff := b.CreateString(dbSchema.Name)
		serial.DatabaseSchemaStart(b)
		serial.DatabaseSchemaAddName(b, nameOff)
		offsets[i] = serial.DatabaseSchemaEnd(b)
	}

	serial.RootValueStartSchemasVector(b, len(offsets))
	for i := len(offsets) - 1; i >= 0; i-- {
		b.PrependUOffsetT(offsets[i])
	}
	return b.EndVector(len(offsets))
}

// encodeTableNameForAddressMap encodes the given table name for writing into storage.
func encodeTableNameForAddressMap(name doltdb.TableName) string {
	if name.Schema == "" {
		return name.Name
	}
	return fmt.Sprintf("\000%s\000%s", name.Schema, name.Name)
}

// decodeTableNameForAddressMap decodes a previously-encoded table name from storage.
func decodeTableNameForAddressMap(encodedName, schemaName string) (string, bool) {
	if schemaName == "" && encodedName[0] != 0 {
		return encodedName, true
	} else if schemaName != "" && encodedName[0] == 0 &&
		len(encodedName) > len(schemaName)+2 &&
		encodedName[1:len(schemaName)+1] == schemaName {
		return encodedName[len(schemaName)+2:], true
	}
	return "", false
}

// rootTableMap is an address map alongside a schema name.
type rootTableMap struct {
	prolly.AddressMap
	schemaName string
}

// Get returns the hash of the table with the given case-sensitive name.
func (m rootTableMap) Get(ctx context.Context, name string) (hash.Hash, error) {
	return m.AddressMap.Get(ctx, encodeTableNameForAddressMap(doltdb.TableName{Name: name, Schema: m.schemaName}))
}

// Iter calls the given callback for each table and hash contained in the map.
func (m rootTableMap) Iter(ctx context.Context, cb func(string, hash.Hash) (bool, error)) error {
	var stop bool
	return m.AddressMap.IterAll(ctx, func(n string, a hash.Hash) error {
		n, ok := decodeTableNameForAddressMap(n, m.schemaName)
		if !stop && ok {
			var err error
			stop, err = cb(n, a)
			return err
		}
		return nil
	})
}
