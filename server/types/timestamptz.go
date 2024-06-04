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
	"fmt"
	"reflect"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
)

// TimestampTZ is the timestamp with a time zone. Precision is unbounded.
var TimestampTZ = TimestampTZType{-1}

// TimestampTZType is the extended type implementation of the PostgreSQL timestamp with time zone.
type TimestampTZType struct {
	// TODO: implement precision
	Precision int8
}

var _ DoltgresType = TimestampTZType{}

// BaseID implements the DoltgresType interface.
func (b TimestampTZType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_TimestampTZ
}

// CollationCoercibility implements the DoltgresType interface.
func (b TimestampTZType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b TimestampTZType) Compare(v1 any, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ac, _, err := b.Convert(v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := b.Convert(v2)
	if err != nil {
		return 0, err
	}

	ab := ac.(time.Time)
	bb := bc.(time.Time)
	return ab.Compare(bb), nil
}

// Convert implements the DoltgresType interface.
func (b TimestampTZType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case time.Time:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b TimestampTZType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatSerializedValue implements the DoltgresType interface.
func (b TimestampTZType) FormatSerializedValue(val []byte) (string, error) {
	deserialized, err := b.DeserializeValue(val)
	if err != nil {
		return "", err
	}
	return b.FormatValue(deserialized)
}

// FormatValue implements the DoltgresType interface.
func (b TimestampTZType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return "", err
	}
	// TODO: this always displays the time with an offset relevant to the server location
	return converted.(time.Time).Format("2006-01-02 15:04:05-07"), nil
}

// GetSerializationID implements the DoltgresType interface.
func (b TimestampTZType) GetSerializationID() SerializationID {
	return SerializationID_TimestampTZ
}

// IoInput implements the DoltgresType interface.
func (b TimestampTZType) IoInput(input string) (any, error) {
	if t, err := time.Parse("2006-01-02 15:04:05-0700", input); err == nil {
		return t, nil
	} else if t, err = time.Parse("2006-01-02 15:04:05-07:00", input); err == nil {
		return t, nil
	} else if t, err = time.Parse("2006-01-02 15:04:05-07", input); err == nil {
		return t, nil
	} else if t, err = time.Parse("January 01 15:04:05 2006 -0700", input); err == nil {
		return t, nil
	} else if t, err = time.Parse("January 01 15:04:05 2006 -07:00", input); err == nil {
		return t, nil
	} else if t, err = time.Parse("January 01 15:04:05 2006 -07", input); err == nil {
		return t, nil
	}
	return nil, fmt.Errorf("invalid format for timestamptz")
}

// IoOutput implements the DoltgresType interface.
func (b TimestampTZType) IoOutput(output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	return converted.(time.Time).Format("2006-01-02 15:04:05-07"), nil
}

// IsUnbounded implements the DoltgresType interface.
func (b TimestampTZType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b TimestampTZType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b TimestampTZType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 40
}

// OID implements the DoltgresType interface.
func (b TimestampTZType) OID() uint32 {
	return uint32(oid.T_timestamptz)
}

// Promote implements the DoltgresType interface.
func (b TimestampTZType) Promote() sql.Type {
	return TimestampTZ
}

// SerializedCompare implements the DoltgresType interface.
func (b TimestampTZType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	// The marshalled time format is byte-comparable
	return bytes.Compare(v1, v2), nil
}

// SQL implements the DoltgresType interface.
func (b TimestampTZType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := b.FormatValue(v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the DoltgresType interface.
func (b TimestampTZType) String() string {
	if b.Precision == -1 {
		return "timestamptz"
	}
	return fmt.Sprintf("timestamptz(%d)", b.Precision)
}

// ToArrayType implements the DoltgresType interface.
func (b TimestampTZType) ToArrayType() DoltgresArrayType {
	return createArrayType(b, SerializationID_TimestampTZArray, oid.T__timestamptz)
}

// Type implements the DoltgresType interface.
func (b TimestampTZType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b TimestampTZType) ValueType() reflect.Type {
	return reflect.TypeOf(time.Time{})
}

// Zero implements the DoltgresType interface.
func (b TimestampTZType) Zero() any {
	return time.Time{}
}

// SerializeType implements the DoltgresType interface.
func (b TimestampTZType) SerializeType() ([]byte, error) {
	t := make([]byte, serializationIDHeaderSize+1)
	copy(t, SerializationID_TimestampTZ.ToByteSlice(0))
	t[serializationIDHeaderSize] = byte(b.Precision)
	return t, nil
}

// deserializeType implements the DoltgresType interface.
func (b TimestampTZType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return TimestampTZType{
			Precision: int8(metadata[0]),
		}, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b TimestampTZType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	return converted.(time.Time).MarshalBinary()
}

// DeserializeValue implements the DoltgresType interface.
func (b TimestampTZType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	t := time.Time{}
	if err := t.UnmarshalBinary(val); err != nil {
		return nil, err
	}
	return t, nil
}
