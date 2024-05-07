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

package rootvalue

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

type fbRvStorage struct {
	srv *serial.RootValue
}

type tableMap interface {
	Get(ctx context.Context, name string) (hash.Hash, error)
	Iter(ctx context.Context, cb func(name string, addr hash.Hash) (bool, error)) error
}

func tmIterAll(ctx context.Context, tm tableMap, cb func(name string, addr hash.Hash)) error {
	return tm.Iter(ctx, func(name string, addr hash.Hash) (bool, error) {
		cb(name, addr)
		return false, nil
	})
}

func (r fbRvStorage) SetForeignKeyMap(ctx context.Context, vrw types.ValueReadWriter, v types.Value) (fbRvStorage, error) {
	var h hash.Hash
	isempty, err := doltdb.EmptyForeignKeyCollection(v.(types.SerialMessage))
	if err != nil {
		return fbRvStorage{}, err
	}
	if !isempty {
		ref, err := vrw.WriteValue(ctx, v)
		if err != nil {
			return fbRvStorage{}, err
		}
		h = ref.TargetHash()
	}
	ret := r.clone()
	copy(ret.srv.ForeignKeyAddrBytes(), h[:])
	return ret, nil
}

func (r fbRvStorage) SetFeatureVersion(v doltdb.FeatureVersion) (fbRvStorage, error) {
	ret := r.clone()
	ret.srv.MutateFeatureVersion(int64(v))
	return ret, nil
}

func (r fbRvStorage) SetCollation(ctx context.Context, collation schema.Collation) (fbRvStorage, error) {
	ret := r.clone()
	ret.srv.MutateCollation(serial.Collation(collation))
	return ret, nil
}

func (r fbRvStorage) GetSchemas(ctx context.Context) ([]schema.DatabaseSchema, error) {
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

func (r fbRvStorage) SetSchemas(ctx context.Context, dbSchemas []schema.DatabaseSchema) (fbRvStorage, error) {
	msg, err := r.serializeRootValue(r.srv.TablesBytes(), dbSchemas)
	if err != nil {
		return fbRvStorage{}, err
	}
	return fbRvStorage{msg}, nil
}

func (r fbRvStorage) clone() fbRvStorage {
	bs := make([]byte, len(r.srv.Table().Bytes))
	copy(bs, r.srv.Table().Bytes)
	var ret serial.RootValue
	ret.Init(bs, r.srv.Table().Pos)
	return fbRvStorage{&ret}
}

func (r fbRvStorage) DebugString(ctx context.Context) string {
	return fmt.Sprintf("fbRvStorage[%d, %s, %s]",
		r.srv.FeatureVersion(),
		"...", // TODO: Print out tables map
		hash.New(r.srv.ForeignKeyAddrBytes()).String())
}

func (r fbRvStorage) nomsValue() types.Value {
	return types.SerialMessage(r.srv.Table().Bytes)
}

func (r fbRvStorage) GetFeatureVersion() (doltdb.FeatureVersion, bool, error) {
	return doltdb.FeatureVersion(r.srv.FeatureVersion()), true, nil
}

func (r fbRvStorage) getAddressMap(vrw types.ValueReadWriter, ns tree.NodeStore) (prolly.AddressMap, error) {
	tbytes := r.srv.TablesBytes()
	node, err := shim.NodeFromValue(types.SerialMessage(tbytes))
	if err != nil {
		return prolly.AddressMap{}, err
	}
	return prolly.NewAddressMap(node, ns)
}

func (r fbRvStorage) GetTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, databaseSchema string) (tableMap, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return nil, err
	}
	return fbTableMap{AddressMap: am, schemaName: databaseSchema}, nil
}

func (r fbRvStorage) GetForeignKeys(ctx context.Context, vr types.ValueReader) (types.Value, bool, error) {
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

func (r fbRvStorage) GetCollation(ctx context.Context) (schema.Collation, error) {
	collation := r.srv.Collation()
	// Pre-existing repositories will return invalid here
	if collation == serial.Collationinvalid {
		return schema.Collation_Default, nil
	}
	return schema.Collation(collation), nil
}

func (r fbRvStorage) EditTablesMap(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, edits []tableEdit) (fbRvStorage, error) {
	am, err := r.getAddressMap(vrw, ns)
	if err != nil {
		return fbRvStorage{}, err
	}
	ae := am.Editor()
	for _, e := range edits {
		if e.old_name != "" {
			oldaddr, err := am.Get(ctx, e.old_name)
			if err != nil {
				return fbRvStorage{}, err
			}
			newaddr, err := am.Get(ctx, encodeTableNameForAddressMap(e.name))
			if err != nil {
				return fbRvStorage{}, err
			}
			if oldaddr.IsEmpty() {
				return fbRvStorage{}, doltdb.ErrTableNotFound
			}
			if !newaddr.IsEmpty() {
				return fbRvStorage{}, doltdb.ErrTableExists
			}
			err = ae.Delete(ctx, e.old_name)
			if err != nil {
				return fbRvStorage{}, err
			}
			err = ae.Update(ctx, encodeTableNameForAddressMap(e.name), oldaddr)
			if err != nil {
				return fbRvStorage{}, err
			}
		} else {
			if e.ref == nil {
				err := ae.Delete(ctx, encodeTableNameForAddressMap(e.name))
				if err != nil {
					return fbRvStorage{}, err
				}
			} else {
				err := ae.Update(ctx, encodeTableNameForAddressMap(e.name), e.ref.TargetHash())
				if err != nil {
					return fbRvStorage{}, err
				}
			}
		}
	}
	am, err = ae.Flush(ctx)
	if err != nil {
		return fbRvStorage{}, err
	}

	ambytes := []byte(tree.ValueFromNode(am.Node()).(types.SerialMessage))
	dbSchemas, err := r.GetSchemas(ctx)
	if err != nil {
		return fbRvStorage{}, err
	}

	msg, err := r.serializeRootValue(ambytes, dbSchemas)
	if err != nil {
		return fbRvStorage{}, err
	}
	return fbRvStorage{msg}, nil
}

func (r fbRvStorage) serializeRootValue(addressMapBytes []byte, dbSchemas []schema.DatabaseSchema) (*serial.RootValue, error) {
	builder := flatbuffers.NewBuilder(80)
	tablesoff := builder.CreateByteVector(addressMapBytes)
	schemasOff := serializeDatabaseSchemas(builder, dbSchemas)

	fkoff := builder.CreateByteVector(r.srv.ForeignKeyAddrBytes())
	serial.RootValueStart(builder)
	serial.RootValueAddFeatureVersion(builder, r.srv.FeatureVersion())
	serial.RootValueAddCollation(builder, r.srv.Collation())
	serial.RootValueAddTables(builder, tablesoff)
	serial.RootValueAddForeignKeyAddr(builder, fkoff)
	if schemasOff > 0 {
		serial.RootValueAddSchemas(builder, schemasOff)
	}

	bs := doltserial.FinishMessage(builder, serial.RootValueEnd(builder), []byte(doltserial.RootValueFileID))
	msg, err := serial.TryGetRootAsRootValue(bs, doltserial.MessagePrefixSz)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

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

func encodeTableNameForAddressMap(name doltdb.TableName) string {
	if name.Schema == "" {
		return name.Name
	}
	return fmt.Sprintf("\000%s\000%s", name.Schema, name.Name)
}

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

type fbTableMap struct {
	prolly.AddressMap
	schemaName string
}

func (m fbTableMap) Get(ctx context.Context, name string) (hash.Hash, error) {
	return m.AddressMap.Get(ctx, encodeTableNameForAddressMap(doltdb.TableName{Name: name, Schema: m.schemaName}))
}

func (m fbTableMap) Iter(ctx context.Context, cb func(string, hash.Hash) (bool, error)) error {
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
