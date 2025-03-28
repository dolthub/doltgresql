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
	"strings"

	"github.com/cockroachdb/errors"
	doltserial "github.com/dolthub/dolt/go/gen/fb/serial"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/store/hash"
	"github.com/dolthub/dolt/go/store/prolly"
	"github.com/dolthub/dolt/go/store/prolly/tree"
	"github.com/dolthub/dolt/go/store/types"
	flatbuffers "github.com/dolthub/flatbuffers/v23/go"

	"github.com/dolthub/doltgresql/core/storage"
	"github.com/dolthub/doltgresql/flatbuffers/gen/serial"
)

// emptyRootValue is Doltgres' implementation of doltdb.EmptyRootValue.
func emptyRootValue(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore) (doltdb.RootValue, error) {
	if vrw.Format().UsesFlatbuffers() {
		builder := flatbuffers.NewBuilder(80)

		emptyam, err := prolly.NewEmptyAddressMap(ns)
		if err != nil {
			return nil, err
		}
		ambytes := []byte(tree.ValueFromNode(emptyam.Node()).(types.SerialMessage))
		tablesoff := builder.CreateByteVector(ambytes)

		var empty hash.Hash
		fkoff := builder.CreateByteVector(empty[:])
		serial.RootValueStart(builder)
		serial.RootValueAddFeatureVersion(builder, int64(DoltgresFeatureVersion))
		serial.RootValueAddCollation(builder, serial.Collationutf8mb4_0900_bin)
		serial.RootValueAddTables(builder, tablesoff)
		serial.RootValueAddForeignKeyAddr(builder, fkoff)
		bs := doltserial.FinishMessage(builder, serial.RootValueEnd(builder), []byte(doltserial.DoltgresRootValueFileID))
		return newRootValue(ctx, vrw, ns, types.SerialMessage(bs))
	}

	empty, err := types.NewMap(ctx, vrw)
	if err != nil {
		return nil, err
	}

	sd := types.StructData{
		tablesKey:      empty,
		foreignKeyKey:  empty,
		featureVersKey: types.Int(DoltgresFeatureVersion),
	}

	st, err := types.NewStruct(vrw.Format(), ddbRootStructName, sd)
	if err != nil {
		return nil, err
	}

	return newRootValue(ctx, vrw, ns, st)
}

// newRootValue is Doltgres' implementation of doltdb.NewRootValue.
func newRootValue(ctx context.Context, vrw types.ValueReadWriter, ns tree.NodeStore, v types.Value) (doltdb.RootValue, error) {
	var st storage.RootStorage

	if !vrw.Format().UsesFlatbuffers() {
		return nil, errors.Errorf("unsupported vrw")
	}
	srv, err := serial.TryGetRootAsRootValue([]byte(v.(types.SerialMessage)), doltserial.MessagePrefixSz)
	if err != nil {
		return nil, err
	}
	st = storage.RootStorage{SRV: srv}
	ver := st.GetFeatureVersion()
	if DoltgresFeatureVersion < ver {
		return nil, doltdb.ErrClientOutOfDate{
			ClientVer: DoltgresFeatureVersion,
			RepoVer:   ver,
		}
	}

	return &RootValue{vrw, ns, st, nil, hash.Hash{}}, nil
}

// rootValueHumanReadableStringAtIndentationLevel is Doltgres' implementation of
// types.DoltgresRootValueHumanReadableStringAtIndentationLevel.
func rootValueHumanReadableStringAtIndentationLevel(sm types.SerialMessage, level int) string {
	msg, _ := serial.TryGetRootAsRootValue(sm, doltserial.MessagePrefixSz)
	ret := &strings.Builder{}
	printWithIndendationLevel(level, ret, "{\n")
	printWithIndendationLevel(level, ret, "\tFeatureVersion: %d\n", msg.FeatureVersion())
	printWithIndendationLevel(level, ret, "\tForeignKeys: #%s\n", hash.New(msg.ForeignKeyAddrBytes()).String())
	printWithIndendationLevel(level, ret, "\tTables: %s\n",
		types.SerialMessage(msg.TablesBytes()).HumanReadableStringAtIndentationLevel(level+1))
	printWithIndendationLevel(level, ret, "}")
	return ret.String()
}

// rootValueWalkAddrs is Doltgres' implementation of types.DoltgresRootValueWalkAddrs.
func rootValueWalkAddrs(sm types.SerialMessage, cb func(addr hash.Hash) error) error {
	var msg serial.RootValue
	err := serial.InitRootValueRoot(&msg, []byte(sm), doltserial.MessagePrefixSz)
	if err != nil {
		return err
	}
	err = types.SerialMessage(msg.TablesBytes()).WalkAddrs(types.Format_DOLT, cb)
	if err != nil {
		return err
	}
	addr := hash.New(msg.ForeignKeyAddrBytes())
	if !addr.IsEmpty() {
		if err = cb(addr); err != nil {
			return err
		}
	}
	return nil
}

// printWithIndendationLevel is a helper for rootValueHumanReadableStringAtIndentationLevel to print at the given
// indentation level.
func printWithIndendationLevel(level int, builder *strings.Builder, format string, a ...any) {
	fmt.Fprint(builder, strings.Repeat("\t", level))
	fmt.Fprintf(builder, format, a...)
}
