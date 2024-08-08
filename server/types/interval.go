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
	"encoding/gob"
	"fmt"
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// Interval is the interval type.
var Interval = IntervalType{}

// IntervalType is the extended type implementation of the PostgreSQL interval.
type IntervalType struct{}

var _ DoltgresType = IntervalType{}

// Alignment implements the DoltgresType interface.
func (b IntervalType) Alignment() TypeAlignment {
	return TypeAlignment_Double
}

// BaseID implements the DoltgresType interface.
func (b IntervalType) BaseID() DoltgresTypeBaseID {
	return DoltgresTypeBaseID_Interval
}

// BaseName implements the DoltgresType interface.
func (b IntervalType) BaseName() string {
	return "interval"
}

// Category implements the DoltgresType interface.
func (b IntervalType) Category() TypeCategory {
	return TypeCategory_TimespanTypes
}

// CollationCoercibility implements the DoltgresType interface.
func (b IntervalType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (b IntervalType) Compare(v1 any, v2 any) (int, error) {
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

	ab := ac.(duration.Duration)
	bb := bc.(duration.Duration)
	return ab.Compare(bb), nil
}

// Convert implements the DoltgresType interface.
func (b IntervalType) Convert(val any) (any, sql.ConvertInRange, error) {
	switch val := val.(type) {
	case duration.Duration:
		return val, sql.InRange, nil
	case nil:
		return nil, sql.InRange, nil
	default:
		return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", b.String(), val)
	}
}

// Equals implements the DoltgresType interface.
func (b IntervalType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(b), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the DoltgresType interface.
func (b IntervalType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return b.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (b IntervalType) GetSerializationID() SerializationID {
	return SerializationID_Interval
}

// IoInput implements the DoltgresType interface.
func (b IntervalType) IoInput(ctx *sql.Context, input string) (any, error) {
	dInterval, err := tree.ParseDInterval(input)
	if err != nil {
		return nil, err
	}
	return dInterval.Duration, nil
}

// IoOutput implements the DoltgresType interface.
func (b IntervalType) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := b.Convert(output)
	if err != nil {
		return "", err
	}
	// TODO: depends on `intervalStyle` configuration variable. Defaults to `postgres`.
	d := converted.(duration.Duration)
	return d.String(), nil
}

// IsPreferredType implements the DoltgresType interface.
func (b IntervalType) IsPreferredType() bool {
	return true
}

// IsUnbounded implements the DoltgresType interface.
func (b IntervalType) IsUnbounded() bool {
	return false
}

// MaxSerializedWidth implements the DoltgresType interface.
func (b IntervalType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (b IntervalType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return 64
}

// OID implements the DoltgresType interface.
func (b IntervalType) OID() uint32 {
	return uint32(oid.T_interval)
}

// Promote implements the DoltgresType interface.
func (b IntervalType) Promote() sql.Type {
	return Interval
}

// SerializedCompare implements the DoltgresType interface.
func (b IntervalType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}

	var d1, d2 duration.Duration
	dec := gob.NewDecoder(bytes.NewReader(v1))
	err := dec.Decode(&d1)
	if err != nil {
		return 0, err
	}

	dec = gob.NewDecoder(bytes.NewReader(v2))
	err = dec.Decode(&d2)
	if err != nil {
		return 0, err
	}

	return d1.Compare(d2), nil
}

// SQL implements the DoltgresType interface.
func (b IntervalType) SQL(ctx *sql.Context, dest []byte, v any) (sqltypes.Value, error) {
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
func (b IntervalType) String() string {
	return "interval"
}

// ToArrayType implements the DoltgresType interface.
func (b IntervalType) ToArrayType() DoltgresArrayType {
	return IntervalArray
}

// Type implements the DoltgresType interface.
func (b IntervalType) Type() query.Type {
	return sqltypes.Text
}

// ValueType implements the DoltgresType interface.
func (b IntervalType) ValueType() reflect.Type {
	return reflect.TypeOf(duration.MakeDuration(0, 0, 0))
}

// Zero implements the DoltgresType interface.
func (b IntervalType) Zero() any {
	return duration.MakeDuration(0, 0, 0)
}

// SerializeType implements the DoltgresType interface.
func (b IntervalType) SerializeType() ([]byte, error) {
	return SerializationID_Interval.ToByteSlice(0), nil
}

// deserializeType implements the DoltgresType interface.
func (b IntervalType) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	switch version {
	case 0:
		return Interval, nil
	default:
		return nil, fmt.Errorf("version %d is not yet supported for %s", version, b.String())
	}
}

// SerializeValue implements the DoltgresType interface.
func (b IntervalType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := b.Convert(val)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err = enc.Encode(converted.(duration.Duration))
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DeserializeValue implements the DoltgresType interface.
func (b IntervalType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	var deserialized duration.Duration
	dec := gob.NewDecoder(bytes.NewReader(val))
	err := dec.Decode(&deserialized)
	return deserialized, err
}
