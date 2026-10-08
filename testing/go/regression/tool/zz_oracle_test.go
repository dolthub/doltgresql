package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
)

// TestOracleCells reads cells as JSON lines of {"oid": N, "v": base64 or null} from ORACLE_IN and writes, for each,
// the replay's RowToString of the normalized cell, its Go type, and for times the location kind.
func TestOracleCells(t *testing.T) {
	in, ok := os.LookupEnv("ORACLE_IN")
	if !ok {
		t.Skip()
	}
	f, err := os.Open(in)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	out, err := os.Create(os.Getenv("ORACLE_OUT"))
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	w := bufio.NewWriter(out)
	defer w.Flush()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64<<20), 64<<20)
	for scanner.Scan() {
		var cell struct {
			OID uint32  `json:"oid"`
			V   *string `json:"v"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &cell); err != nil {
			t.Fatal(err)
		}
		var value []byte
		if cell.V != nil {
			value, err = base64.StdEncoding.DecodeString(*cell.V)
			if err != nil {
				t.Fatal(err)
			}
		}
		result := map[string]any{}
		func() {
			defer func() {
				if r := recover(); r != nil {
					result["panic"] = fmt.Sprint(r)
				}
			}()
			rows := ReadRows(&pgproto3.RowDescription{Fields: []pgproto3.FieldDescription{{DataTypeOID: cell.OID}}},
				[]*pgproto3.DataRow{{Values: [][]byte{value}}})
			v := rows[0][0]
			result["key"] = RowToString(rows[0])
			result["type"] = fmt.Sprintf("%T", v)
			if tv, ok := v.(time.Time); ok {
				switch tv.Location() {
				case time.UTC:
					result["loc"] = "UTC"
				case time.Local:
					result["loc"] = "Local"
				default:
					_, offset := tv.Zone()
					result["loc"] = fmt.Sprintf("Fixed(%d)", offset)
				}
			}
			if fv, ok := v.(float64); ok && fv == 0 && 1/fv < 0 {
				result["negzero"] = true
			}
		}()
		line, _ := json.Marshal(result)
		w.Write(line)
		w.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
