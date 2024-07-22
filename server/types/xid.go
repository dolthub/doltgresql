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

package types

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// Xid is a data type used for internal transaction IDs. It is implemented as an unsigned 32 bit integer.
var Xid = XidType{}

// XidType is the extended type implementation of the PostgreSQL xid.
type XidType struct{}

var _ DoltgresType = XidType{}

// Alignment implements the DoltgresType interface.
func (b XidType) Alignment() TypeAlignment {
	return TypeAlignment_Int
}

// BaseID implements the DoltgresType interface.
func (b XidType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Xid
}

// BaseName implements the DoltgresType interface.
func (b XidType) BaseName() string {
	return "xid"
}

// Category implements the DoltgresType interface.
func (b XidType) Category() TypeCategory {
	return TypeCategory_UserDefinedTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b XidType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b XidType) Compare(v1 any, v2 any) (int, error) {
	return compareUint32(b, v1, v2)
}

// Convert implements the DoltgresType interface.
func (b XidType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case uint32:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b XidType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b XidType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b XidType) GetSerializationID() SerializationID {
	return SerializationID_Xid
}

// IoInput implements the DoltgresType interface.
func (b XidType) IoInput(ctx *sql.Context, input string) (any, error) {
	val, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil {
		return uint32(0), nil
	}
	return uint32(val), nil
}

// IoOutput implements the DoltgresType interface.
func (b XidType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return strconv.FormatUint(uint64(converted.(uint32)), 10), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b XidType) IsPreferredType() bool {
	return false
}

// IsUnbounded implements the DoltgresType interface.
func (b XidType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b XidType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b XidType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 4
}

// OID implements the DoltgresType interface.
func (b XidType) OID() uint32 {
	return uint32(oid.T_xid)
}

// Promote implements the DoltgresType interface.
func (b XidType) Promote() sql.Type {
	return b
}

// SerializedCompare implements the DoltgresType interface.
func (b XidType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresType interface.
func (b XidType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.IoOutput(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b XidType) String() string {
	return "xid"
}

// ToArrayType implements the DoltgresType interface.
func (b XidType) ToArrayType() DoltgresArrayType {
	return XidArray
}

// Type implements the DoltgresType interface.
func (b XidType) Type() query.Type {
	return sqltypes.Uint32
}

// ValueType implements the DoltgresType interface.
func (b XidType) ValueType() reflect.Type {
	return reflect.TypeOf(uint32(0))
}

// Zero implements the DoltgresType interface.
func (b XidType) Zero() any {
	return uint32(0)
}

// SerializeType implements the DoltgresType interface.
func (b XidType) SerializeType() ([]byte, error) {
	return SerializationID_Xid.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b XidType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Xid, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b XidType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	retVal := make([]byte, 4)
	binary.BigEndian.PutUint32(retVal, uint32(converted.(uint32)))
	return retVal, nil
}

// DeserializeValue implements the DoltgresType interface.
func (b XidType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return uint32(binary.BigEndian.Uint32(val)), nil
}
