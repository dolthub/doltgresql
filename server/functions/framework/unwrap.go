// Copyright 2026 Dolthub, Inc.
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

package framework

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
)

// UnwrapString converts a string-typed function argument into a string. Large string values may arrive as a wrapper
// (e.g. *val.TextStorage) whose contents are stored out-of-band, in which case this loads the full value into memory.
// An error is returned if the value is neither a string nor a wrapper around one.
func UnwrapString(ctx *sql.Context, val any) (string, error) {
	str, ok, err := sql.Unwrap[string](ctx, val)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", errors.Errorf("expected string value, got %T", val)
	}
	return str, nil
}

// UnwrapBytes converts a bytes-typed function argument into a []byte. Large bytea values may arrive as a wrapper
// (e.g. *val.ByteArray) whose contents are stored out-of-band, in which case this loads the full value into memory.
// An error is returned if the value is neither a []byte nor a wrapper around one.
func UnwrapBytes(ctx *sql.Context, val any) ([]byte, error) {
	data, ok, err := sql.Unwrap[[]byte](ctx, val)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("expected []byte value, got %T", val)
	}
	return data, nil
}

// UnwrapByteLength returns the length in bytes of a string- or bytes-typed function argument. Large values may arrive
// as a wrapper (e.g. *val.ByteArray) whose contents are stored out-of-band; when such a wrapper already knows its
// exact byte length, that length is used rather than loading the full value into memory.
func UnwrapByteLength(ctx *sql.Context, val any) (int64, error) {
	if wrapper, ok := val.(sql.AnyWrapper); ok {
		if wrapper.IsExactLength() {
			return wrapper.MaxByteLength(), nil
		}
		unwrapped, err := wrapper.UnwrapAny(ctx)
		if err != nil {
			return 0, err
		}
		val = unwrapped
	}
	switch val := val.(type) {
	case string:
		return int64(len(val)), nil
	case []byte:
		return int64(len(val)), nil
	default:
		return 0, errors.Errorf("expected string or []byte value, got %T", val)
	}
}
