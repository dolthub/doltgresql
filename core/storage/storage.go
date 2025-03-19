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

package storage

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

// RootStorage is the FlatBuffer interface for the storage format.
type RootStorage struct {
	SRV *serial.RootValue
}

type TableEdit struct {
	Name doltdb.TableName
	Ref  *types.Ref

	// Used for rename.
	OldName doltdb.TableName
}

// RootObjectSerialization handles the allocation/preservation of bytes for root objects.
type RootObjectSerialization struct {
	Bytes        func(*serial.RootValue) []byte
	RootValueAdd func(builder *flatbuffers.Builder, sequences flatbuffers.UOffsetT)
}

// RootObjectSerializations contains all root object serializations. This should be set from the global initialization
// function.
var RootObjectSerializations []RootObjectSerialization

// SetForeignKeyMap sets the foreign key and returns a new storage object.
func (r RootStorage) SetForeignKeyMap(ctx context.Context, vrw types.ValueReadWriter, v types.Value) (RootStorage, error) {
	var h hash.Hash
	isempty, err := doltdb.EmptyForeignKeyCollection(v.(types.SerialMessage))
	if err != nil {
		return RootStorage{}, err
	}
	if !isempty {
		ref, err := vrw.WriteValue(ctx, v)
		if err != nil {
			return RootStorage{}, err
		}
		h = ref.TargetHash()
	}
	ret := r.Clone()
	copy(ret.SRV.ForeignKeyAddrBytes(), h[:])
	return ret, nil
}

// SetFeatureVersion sets the feature version and returns a new storage object.
func (r RootStorage) SetFeatureVersion(v doltdb.FeatureVersion) (RootStorage, error) {
	ret := r.Clone()
	ret.SRV.MutateFeatureVersion(int64(v))
	return ret, nil
}

// SetCollation sets the collation and returns a new storage object.
func (r RootStorage) SetCollation(ctx context.Context, collation schema.Collation) (RootStorage, error) {
	ret := r.Clone()
	ret.SRV.MutateCollation(serial.Collation(collation))
	return ret, nil
}

// GetSchemas returns all schemas.
func (r RootStorage) GetSchemas(ctx context.Context) ([]schema.DatabaseSchema, error) {
	numSchemas := r.SRV.SchemasLength()
	schemas := make([]schema.DatabaseSchema, numSchemas)
	for i := 0; i < numSchemas; i++ {
		dbSchema := new(serial.DatabaseSchema)
		_, err := r.SRV.TrySchemas(dbSchema, i)
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
func (r RootStorage) SetSchemas(ctx context.Context, dbSchemas []schema.DatabaseSchema) (RootStorage, error) {
	msg, err := r.serializeRootValue(r.SRV.TablesBytes(), dbSchemas)
	if err != nil {
		return RootStorage{}, err
	}
	return RootStorage{msg}, nil
}

// Clone returns a clone of the calling storage.
func (r RootStorage) Clone() RootStorage {
	bs := make([]byte, len(r.SRV.Table().Bytes))
	copy(bs, r.SRV.Table().Bytes)
	var ret serial.RootValue
	ret.Init(bs, r.SRV.Table().Pos)
	return RootStorage{&ret}
}

// DebugString returns the storage as a printable string.
func (r RootStorage) DebugString(ctx context.Context) string {
	return fmt.Sprintf("RootStorage[%d, %s, %s]",
		r.SRV.FeatureVersion(),
		"...",
		hash.New(r.SRV.ForeignKeyAddrBytes()).String())
}

// NomsValue returns the storage as a noms value.
func (r RootStorage) NomsValue() types.Value {
	return types.SerialMessage(r.SRV.Table().Bytes)
}

// GetFeatureVersion returns the feature version for this storage object.
func (r RootStorage) GetFeatureVersion() doltdb.FeatureVersion {
	return doltdb.FeatureVersion(r.SRV.FeatureVersion())
}

// getAddressMap returns the address map from within this storage object.
func (r RootStorage) getAddressMap(vrw types.ValueReadWriter, ns tree.NodeStore) (prolly.AddressMap, error) {
	tbytes := r.SRV.TablesBytes()
	node, _, err := shim.NodeFromValue(types.SerialMessage(tbytes))
	if err != nil {
		return prolly.AddressMap{}, err
	}
	return prolly.NewAddressMap(node, ns)
}

// GetTablesMap returns the tables map from within this storage object.
func (r RootStorage) GetTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, databaseSchema string) (RootTableMap, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return RootTableMap{}, err
	}
	return RootTableMap{AddressMap: am, schemaName: databaseSchema}, nil
}

// GetForeignKeys returns the types.SerialMessage representing the foreign keys.
func (r RootStorage) GetForeignKeys(ctx context.Context, vr types.ValueReader) (types.Value, bool, error) {
	addr := hash.New(r.SRV.ForeignKeyAddrBytes())
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
func (r RootStorage) GetCollation(ctx context.Context) (schema.Collation, error) {
	collation := r.SRV.Collation()
	// Pre-existing repositories will return invalid here
	if collation == serial.Collationinvalid {
		return schema.Collation_Default, nil
	}
	return schema.Collation(collation), nil
}

// EditTablesMap edits the table map within storage.
func (r RootStorage) EditTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, edits []TableEdit) (RootStorage, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return RootStorage{}, err
	}
	ae := am.Editor()
	for _, e := range edits {
		if e.OldName.Name != "" {
			oldaddr, err := am.Get(ctx, encodeTableNameForAddressMap(e.OldName))
			if err != nil {
				return RootStorage{}, err
			}
			newaddr, err := am.Get(ctx, encodeTableNameForAddressMap(e.Name))
			if err != nil {
				return RootStorage{}, err
			}
			if oldaddr.IsEmpty() {
				return RootStorage{}, doltdb.ErrTableNotFound
			}
			if !newaddr.IsEmpty() {
				return RootStorage{}, doltdb.ErrTableExists
			}
			err = ae.Delete(ctx, encodeTableNameForAddressMap(e.OldName))
			if err != nil {
				return RootStorage{}, err
			}
			err = ae.Update(ctx, encodeTableNameForAddressMap(e.Name), oldaddr)
			if err != nil {
				return RootStorage{}, err
			}
		} else {
			if e.Ref == nil {
				err := ae.Delete(ctx, encodeTableNameForAddressMap(e.Name))
				if err != nil {
					return RootStorage{}, err
				}
			} else {
				err := ae.Update(ctx, encodeTableNameForAddressMap(e.Name), e.Ref.TargetHash())
				if err != nil {
					return RootStorage{}, err
				}
			}
		}
	}
	am, err = ae.Flush(ctx)
	if err != nil {
		return RootStorage{}, err
	}

	ambytes := []byte(tree.ValueFromNode(am.Node()).(types.SerialMessage))
	dbSchemas, err := r.GetSchemas(ctx)
	if err != nil {
		return RootStorage{}, err
	}

	msg, err := r.serializeRootValue(ambytes, dbSchemas)
	if err != nil {
		return RootStorage{}, err
	}
	return RootStorage{msg}, nil
}

// serializeRootValue serializes a new serial.RootValue object.
func (r RootStorage) serializeRootValue(addressMapBytes []byte, dbSchemas []schema.DatabaseSchema) (*serial.RootValue, error) {
	builder := flatbuffers.NewBuilder(80)
	tablesOffset := builder.CreateByteVector(addressMapBytes)
	schemasOffset := serializeDatabaseSchemas(builder, dbSchemas)
	fkOffset := builder.CreateByteVector(r.SRV.ForeignKeyAddrBytes())
	rootObjOffsets := make([]flatbuffers.UOffsetT, len(RootObjectSerializations))
	for i := range RootObjectSerializations {
		rootObjOffset := RootObjectSerializations[i].Bytes(r.SRV)
		if len(rootObjOffset) == 0 {
			h := hash.Hash{}
			rootObjOffset = h[:]
		}
		rootObjOffsets[i] = builder.CreateByteVector(rootObjOffset)
	}

	serial.RootValueStart(builder)
	serial.RootValueAddFeatureVersion(builder, r.SRV.FeatureVersion())
	serial.RootValueAddCollation(builder, r.SRV.Collation())
	serial.RootValueAddTables(builder, tablesOffset)
	serial.RootValueAddForeignKeyAddr(builder, fkOffset)
	for i := range RootObjectSerializations {
		RootObjectSerializations[i].RootValueAdd(builder, rootObjOffsets[i])
	}
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

// RootTableMap is an address map alongside a schema name.
type RootTableMap struct {
	prolly.AddressMap
	schemaName string
}

// Get returns the hash of the table with the given case-sensitive name.
func (m RootTableMap) Get(ctx context.Context, name string) (hash.Hash, error) {
	return m.AddressMap.Get(ctx, encodeTableNameForAddressMap(doltdb.TableName{Name: name, Schema: m.schemaName}))
}

// Iter calls the given callback for each table and hash contained in the map.
func (m RootTableMap) Iter(ctx context.Context, cb func(string, hash.Hash) (bool, error)) error {
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
