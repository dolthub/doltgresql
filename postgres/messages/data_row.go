package messages

import (
	"bytes"
	"fmt"
	"strconv"
	"time"

	"github.com/dolthub/vitess/go/sqltypes"

	"github.com/shopspring/decimal"
)

// DataRow represents a row of data.
type DataRow struct {
	Values []DataRowValue
}

// DataRowValue represents a column's value in a DataRow.
type DataRowValue struct {
	Value any
}

// NewDataRow creates a new DataRow from the given rows.
func NewDataRow(row []sqltypes.Value) DataRow {
	values := make([]DataRowValue, len(row))
	for i, value := range row {
		values[i] = DataRowValue{value.ToString()}
	}
	return DataRow{
		Values: values,
	}
}

// Bytes returns DataRow as a byte slice, ready to be returned to the client.
func (dr DataRow) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('D')          // Message Type
	WriteNumber(&buf, int32(0)) // Message length, will be corrected later
	WriteNumber(&buf, int16(len(dr.Values)))
	for _, drv := range dr.Values {
		drv.Bytes(&buf)
	}
	return WriteLength(buf.Bytes())
}

// Bytes writes the value into the given buffer.
func (drv DataRowValue) Bytes(buf *bytes.Buffer) {
	var dataBytes []byte
	switch val := drv.Value.(type) {
	case int:
		dataBytes = []byte(strconv.FormatInt(int64(val), 10))
	case int8:
		dataBytes = []byte(strconv.FormatInt(int64(val), 10))
	case int16:
		dataBytes = []byte(strconv.FormatInt(int64(val), 10))
	case int32:
		dataBytes = []byte(strconv.FormatInt(int64(val), 10))
	case int64:
		dataBytes = []byte(strconv.FormatInt(val, 10))
	case uint:
		dataBytes = []byte(strconv.FormatUint(uint64(val), 10))
	case uint8:
		dataBytes = []byte(strconv.FormatUint(uint64(val), 10))
	case uint16:
		dataBytes = []byte(strconv.FormatUint(uint64(val), 10))
	case uint32:
		dataBytes = []byte(strconv.FormatUint(uint64(val), 10))
	case uint64:
		dataBytes = []byte(strconv.FormatUint(val, 10))
	case float32:
		dataBytes = []byte(strconv.FormatFloat(float64(val), 'g', -1, 32))
	case float64:
		dataBytes = []byte(strconv.FormatFloat(val, 'g', -1, 64))
	case decimal.NullDecimal:
		if !val.Valid {
			WriteNumber(buf, int32(-1))
			return
		}
		dataBytes = []byte(val.Decimal.String())
	case decimal.Decimal:
		dataBytes = []byte(val.String())
	case []byte:
		dataBytes = val
	case string:
		dataBytes = []byte(val)
	case bool:
		if val {
			dataBytes = []byte("true")
		} else {
			dataBytes = []byte("false")
		}
	case time.Time:
		dataBytes = []byte(val.Format(time.RFC3339))
	case nil:
		WriteNumber(buf, int32(-1))
		return
	default:
		panic(fmt.Errorf("unknown DataRow value type: %T", val))
	}
	WriteNumber(buf, int32(len(dataBytes))) // This length only covers the value's size
	buf.Write(dataBytes)
}
