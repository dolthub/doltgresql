package _go

// Temporary instrumentation for the Rust port: when DUMP_TESTS_FILE is set, the test runners write their test
// definitions to that file as JSON lines instead of running them. Never committed.

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"
)

var dumpMutex sync.Mutex

// DumpTests writes the tests passed to a runner when DUMP_TESTS_FILE is set, returning true when it did.
func DumpTests(t *testing.T, runner string, tests any, extra map[string]any) bool {
	path, ok := os.LookupEnv("DUMP_TESTS_FILE")
	if !ok {
		return false
	}
	if !filepath.IsAbs(path) {
		_, self, _, _ := runtime.Caller(0)
		path = filepath.Join(filepath.Dir(self), path)
	}
	callers := []string{}
	for skip := 1; skip < 8; skip++ {
		_, file, line, ok := runtime.Caller(skip)
		if !ok {
			break
		}
		callers = append(callers, fmt.Sprintf("%s:%d", file, line))
	}
	record := map[string]any{
		"test":    t.Name(),
		"runner":  runner,
		"callers": callers,
		"extra":   extra,
		"tests":   dumpValue(reflect.ValueOf(tests)),
	}
	data, err := json.Marshal(record)
	if err != nil {
		panic(err)
	}
	dumpMutex.Lock()
	defer dumpMutex.Unlock()
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		panic(err)
	}
	return true
}

// dumpValue converts a value into JSON-friendly data, tagging every dynamically typed value with its Go type.
func dumpValue(v reflect.Value) any {
	if !v.IsValid() {
		return nil
	}
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			return map[string]any{"$type": "nil"}
		}
		inner := v.Elem()
		return map[string]any{"$type": inner.Type().String(), "$value": dumpConcrete(inner)}
	default:
		return dumpConcrete(v)
	}
}

// dumpConcrete converts a value of a concrete type.
func dumpConcrete(v reflect.Value) any {
	if !(v.Kind() == reflect.Pointer && v.IsNil()) && v.CanInterface() {
		if m, ok := v.Interface().(json.Marshaler); ok {
			data, err := m.MarshalJSON()
			if err == nil {
				var decoded any
				if json.Unmarshal(data, &decoded) == nil {
					return map[string]any{"$json": decoded, "$go": fmt.Sprintf("%#v", v.Interface())}
				}
			}
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return map[string]any{"$ptr": v.Elem().Type().String(), "$value": dumpConcrete(v.Elem())}
	case reflect.Interface:
		return dumpValue(v)
	case reflect.Struct:
		fields := map[string]any{"$struct": v.Type().String()}
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			fv := v.Field(i)
			if !field.IsExported() {
				fields["$unexported"] = fmt.Sprintf("%#v", v.Interface())
				continue
			}
			if fv.IsZero() {
				continue
			}
			fields[field.Name] = dumpValue(fv)
		}
		return fields
	case reflect.Slice:
		if v.IsNil() {
			return nil
		}
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return map[string]any{"$bytes": base64.StdEncoding.EncodeToString(v.Bytes())}
		}
		fallthrough
	case reflect.Array:
		if v.Type().Elem().Kind() == reflect.Uint8 && v.Kind() == reflect.Array {
			b := make([]byte, v.Len())
			for i := range b {
				b[i] = byte(v.Index(i).Uint())
			}
			return map[string]any{"$bytes": base64.StdEncoding.EncodeToString(b)}
		}
		items := make([]any, v.Len())
		for i := range items {
			items[i] = dumpValue(v.Index(i))
		}
		return items
	case reflect.Map:
		items := map[string]any{"$map": v.Type().String()}
		entries := []any{}
		iter := v.MapRange()
		for iter.Next() {
			entries = append(entries, []any{dumpValue(iter.Key()), dumpValue(iter.Value())})
		}
		items["$entries"] = entries
		return items
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return map[string]any{"$float": fmt.Sprintf("%v", f)}
		}
		return map[string]any{"$float": f}
	case reflect.Func:
		return map[string]any{"$func": runtime.FuncForPC(v.Pointer()).Name()}
	case reflect.Chan:
		return map[string]any{"$chan": v.Type().String()}
	default:
		return v.Interface()
	}
}
