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

package main

import (
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/dolthub/doltgresql/utils"
)

// CompareRowsOrdered compares the two rows, enforcing that the order matches between the two rows.
func CompareRowsOrdered(aRowDesc, bRowDesc *pgproto3.RowDescription, aRows, bRows []*pgproto3.DataRow) error {
	if len(aRows) != len(bRows) {
		return errors.Errorf("expected a row count of %d but received %d", len(aRows), len(bRows))
	}
	aReadRows := ReadRows(aRowDesc, aRows)
	bReadRows := ReadRows(bRowDesc, bRows)
	for rowIdx := range aReadRows {
		if len(aReadRows[rowIdx]) != len(bReadRows[rowIdx]) {
			return errors.Errorf("expected a row column count of %d but received %d",
				len(aReadRows[rowIdx]), len(bReadRows[rowIdx]))
		}
		for colIdx := range aReadRows[rowIdx] {
			if aReadRows[rowIdx][colIdx] != bReadRows[rowIdx][colIdx] {
				if len(aReadRows)+len(bReadRows) < 8 {
					return errors.Errorf("row sets differ:\n%s", rowsToErrorString(aReadRows, bReadRows))
				} else {
					return errors.Errorf("rows differ\n    Postgres:\n        {%s}\n    Doltgres:\n        {%s}",
						RowToString(aReadRows[rowIdx]), RowToString(bReadRows[rowIdx]))
				}
			}
		}
	}
	return nil
}

// CompareRowsUnordered compares the two rows. Order is not enforced, however if there are any duplicate rows, then it
// is expected that the duplicate counts match.
func CompareRowsUnordered(aRowDesc, bRowDesc *pgproto3.RowDescription, aRows, bRows []*pgproto3.DataRow) error {
	if len(aRows) != len(bRows) {
		return errors.Errorf("expected a row count of %d but received %d", len(aRows), len(bRows))
	}
	// It's possible that two different rows can hash to the same result, but we're not concerned with that.
	// The same row will always output the same hash, and that's the only property that we really care about.
	aMap := make(map[string]int)
	bMap := make(map[string]int)
	aReadRows := ReadRows(aRowDesc, aRows)
	bReadRows := ReadRows(bRowDesc, bRows)
	for rowIdx := range aReadRows {
		// Column counts should always match, so this is a sanity check
		if len(aReadRows[rowIdx]) != len(bReadRows[rowIdx]) {
			return errors.Errorf("expected a row column count of %d but received %d",
				len(aReadRows[rowIdx]), len(bReadRows[rowIdx]))
		}
		// We'll use the string form of each row as though it were a hash
		aHash := RowToString(aReadRows[rowIdx])
		bHash := RowToString(bReadRows[rowIdx])
		if count, ok := aMap[aHash]; ok {
			aMap[aHash] = count + 1
		} else {
			aMap[aHash] = 1
		}
		if count, ok := bMap[bHash]; ok {
			bMap[bHash] = count + 1
		} else {
			bMap[bHash] = 1
		}
	}
	aKVs := utils.GetMapKVsSorted(aMap)
	bKVs := utils.GetMapKVsSorted(bMap)
	// One map may have duplicates that the other does not have, so we have to do a length check again
	if len(aKVs) != len(bKVs) {
		aCount := 0
		bCount := 0
		for _, aKV := range aKVs {
			aCount += aKV.Value
		}
		for _, bKV := range bKVs {
			bCount += bKV.Value
		}
		if aCount+bCount < 8 {
			return errors.Errorf("row sets differ:\n%s", rowKVsToErrorString(aKVs, bKVs))
		} else {
			return errors.New("row sets differ (too large to display)")
		}
	}
	for i := range aKVs {
		if aKVs[i].Key != bKVs[i].Key || aKVs[i].Value != bKVs[i].Value {
			aCount := 0
			bCount := 0
			for _, aKV := range aKVs {
				aCount += aKV.Value
			}
			for _, bKV := range bKVs {
				bCount += bKV.Value
			}
			if aCount+bCount < 8 {
				return errors.Errorf("row sets differ:\n%s", rowKVsToErrorString(aKVs, bKVs))
			} else {
				if aKVs[i].Key != bKVs[i].Key {
					return errors.Errorf("could not find the following row in the result set:\n        {%s}", aKVs[i].Key)
				} else {
					return errors.Errorf("for the following row, expected to find %d duplicates but found %d:\n {%s}",
						aKVs[i].Value, bKVs[i].Value, aKVs[i].Key)
				}
			}
		}
	}
	return nil
}

// ReadRows reads the given rows into their native Go types.
func ReadRows(rowDesc *pgproto3.RowDescription, rows []*pgproto3.DataRow) []sql.Row {
	if len(rows) == 0 {
		return nil
	}
	typeMap := pgtype.NewMap()
	results := make([]sql.Row, len(rows))
	for rowIdx, row := range rows {
		resultRow := make(sql.Row, len(rowDesc.Fields))
		// This should never happen, but we can't be too safe since it's technically possible
		if len(rowDesc.Fields) != len(row.Values) {
			continue
		}
		for colIdx := range rowDesc.Fields {
			field := rowDesc.Fields[colIdx]
			if err := typeMap.Scan(field.DataTypeOID, field.Format, row.Values[colIdx], &resultRow[colIdx]); err != nil {
				resultRow[colIdx] = string(row.Values[colIdx])
			}
			switch val := resultRow[colIdx].(type) {
			case bool:
			case int:
				resultRow[colIdx] = int64(val)
			case int8:
			case int16:
			case int32:
			case int64:
			case uint:
				resultRow[colIdx] = uint64(val)
			case uint8:
			case uint16:
			case uint32:
			case uint64:
			case float32:
			case float64:
			case string:
			case []byte:
				resultRow[colIdx] = hex.EncodeToString(val)
			case [16]byte:
				resultRow[colIdx] = hex.EncodeToString(val[:])
			case time.Time:
			case pgtype.InfinityModifier:
				switch val {
				case pgtype.Infinity:
					resultRow[colIdx] = math.Inf(1)
				case pgtype.Finite:
					resultRow[colIdx] = float64(0)
				case pgtype.NegativeInfinity:
					resultRow[colIdx] = math.Inf(-1)
				}
			case pgtype.Range[interface{}]:
				resultRow[colIdx] = string(row.Values[colIdx])
			case pgtype.Multirange[pgtype.Range[interface{}]]:
				resultRow[colIdx] = string(row.Values[colIdx])
			case map[string]any:
				resultRow[colIdx], _ = json.Marshal(val)
			case []any:
				resultRow[colIdx] = "[" + RowToString(val) + "]"
			case nil:
			default:
				if driverValue, ok := val.(driver.Valuer); ok {
					resultRow[colIdx], _ = driverValue.Value()
				} else if stringer, ok := val.(fmt.Stringer); ok {
					resultRow[colIdx] = stringer.String()
				} else {
					// This makes it much simpler for the sake of comparison, but may be wrong.
					// At the time of writing (meaning no further updates to the test files), all types are covered.
					resultRow[colIdx] = nil
				}
			}
		}
		results[rowIdx] = resultRow
	}
	return results
}

// RowToString returns the row as a string, which may be used for printing.
func RowToString(row sql.Row) string {
	values := make([]string, len(row))
	for i := range row {
		switch val := row[i].(type) {
		case bool:
			if val {
				values[i] = "true"
			} else {
				values[i] = "false"
			}
		case int:
			values[i] = fmt.Sprintf("%d", val)
		case int8:
			values[i] = fmt.Sprintf("%d", val)
		case int16:
			values[i] = fmt.Sprintf("%d", val)
		case int32:
			values[i] = fmt.Sprintf("%d", val)
		case int64:
			values[i] = fmt.Sprintf("%d", val)
		case uint:
			values[i] = fmt.Sprintf("%d", val)
		case uint8:
			values[i] = fmt.Sprintf("%d", val)
		case uint16:
			values[i] = fmt.Sprintf("%d", val)
		case uint32:
			values[i] = fmt.Sprintf("%d", val)
		case uint64:
			values[i] = fmt.Sprintf("%d", val)
		case float32:
			values[i] = fmt.Sprintf("%f", val)
		case float64:
			values[i] = fmt.Sprintf("%f", val)
		case string:
			values[i] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(val, `"`, `\"`))
		case time.Time:
			values[i] = val.String()
		case nil:
			values[i] = "\uFFFD"
		default:
			values[i] = fmt.Sprintf("%v", val)
		}
	}
	return strings.Join(values, ", ")
}

// rowsToErrorString returns the given rows formatted to be displayed in the regression comment.
func rowsToErrorString(postgresRows []sql.Row, doltgresRows []sql.Row) string {
	sb := strings.Builder{}
	sb.WriteString("    Postgres:\n")
	for _, row := range postgresRows {
		sb.WriteString("        {")
		sb.WriteString(RowToString(row))
		sb.WriteString("}\n")
	}
	sb.WriteString("    Doltgres:\n")
	for _, row := range doltgresRows {
		sb.WriteString("        {")
		sb.WriteString(RowToString(row))
		sb.WriteString("}\n")
	}
	returnErr := sb.String()
	// Removing the last newline since it's not necessary
	return returnErr[:len(returnErr)-1]
}

// rowKVsToErrorString returns the given row KVs formatted to be displayed in the regression comment.
func rowKVsToErrorString(postgresRows []utils.KeyValue[string, int], doltgresRows []utils.KeyValue[string, int]) string {
	sb := strings.Builder{}
	sb.WriteString("    Postgres:\n")
	for _, kv := range postgresRows {
		for i := 0; i < kv.Value; i++ {
			sb.WriteString("        {")
			sb.WriteString(kv.Key)
			sb.WriteString("}\n")
		}
	}
	sb.WriteString("    Doltgres:\n")
	for _, kv := range doltgresRows {
		for i := 0; i < kv.Value; i++ {
			sb.WriteString("        {")
			sb.WriteString(kv.Key)
			sb.WriteString("}\n")
		}
	}
	returnErr := sb.String()
	// Removing the last newline since it's not necessary
	return returnErr[:len(returnErr)-1]
}
