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

package functions

import (
	"strings"
	"unicode"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initArray registers the functions to the catalog.
func initArray() {
	framework.RegisterFunction(array_in)
	framework.RegisterFunction(array_out)
	framework.RegisterFunction(array_recv)
	framework.RegisterFunction(array_send)
	framework.RegisterFunction(btarraycmp)
	framework.RegisterFunction(array_subscript_handler)
}

// array_in represents the PostgreSQL function of array type IO input.
var array_in = framework.Function3{
	Name:       "array_in",
	Return:     pgtypes.AnyArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input, err := framework.UnwrapString(ctx, val1)
		if err != nil {
			return nil, err
		}
		baseTypeOid := val2.(id.Id)
		baseType := pgtypes.IDToBuiltInDoltgresType[id.Type(baseTypeOid)]
		if baseType == nil {
			if typColl, err := pgtypes.GetTypesCollectionFromContext(ctx, ""); err == nil && typColl != nil {
				if t, err := typColl.GetType(ctx, id.Type(baseTypeOid)); err == nil {
					baseType = t
				}
			}
		}
		if baseType == nil {
			return nil, errors.Errorf("unknown array element type: %s", string(baseTypeOid))
		}
		typmod := val3.(int32)
		parser := arrayLiteralParser{input: input, runes: []rune(input), baseType: baseType.WithAttTypMod(typmod)}
		return parser.parse(ctx)
	},
}

// arrayLiteralParser parses the text representation of an array, which may be multidimensional. Malformed literals are
// critical errors that return a nil value, while element conversion errors are non-critical and are returned alongside
// the parsed value so that a higher layer (such as an explicit cast) may ignore them.
type arrayLiteralParser struct {
	input      string
	runes      []rune
	pos        int
	baseType   *pgtypes.DoltgresType
	elementErr error
}

// parse parses the entire input.
func (p *arrayLiteralParser) parse(ctx *sql.Context) (any, error) {
	p.skipWhitespace()
	if p.peek() != '{' {
		return nil, p.malformed()
	}
	vals, err := p.parseArray(ctx, false)
	if err != nil {
		return nil, err
	}
	p.skipWhitespace()
	if p.pos != len(p.runes) {
		return nil, p.malformed()
	}
	return vals, p.elementErr
}

// parseArray parses the array that starts at the current opening brace.
func (p *arrayLiteralParser) parseArray(ctx *sql.Context, nested bool) ([]any, error) {
	p.pos++
	p.skipWhitespace()
	if p.peek() == '}' && !nested {
		p.pos++
		return []any{}, nil
	}
	var vals []any
	for {
		p.skipWhitespace()
		var val any
		var err error
		if p.peek() == '{' {
			val, err = p.parseArray(ctx, true)
		} else {
			val, err = p.parseElement(ctx)
		}
		if err != nil {
			return nil, err
		}
		vals = append(vals, val)
		p.skipWhitespace()
		switch p.next() {
		case ',':
		case '}':
			if !p.baseType.IsVectorType() && !pgtypes.SameArrayDims(vals) {
				return nil, p.malformed()
			}
			return vals, nil
		default:
			return nil, p.malformed()
		}
	}
}

// parseElement parses a single, optionally quoted, element and converts it to the base type. An unquoted NULL is the
// null value.
func (p *arrayLiteralParser) parseElement(ctx *sql.Context) (any, error) {
	sb := strings.Builder{}
	pendingWhitespace := strings.Builder{}
	quoted := p.peek() == '"'
	if quoted {
		p.pos++
	}
	for {
		r := p.next()
		switch {
		case r == 0:
			return nil, p.malformed()
		case r == '\\':
			escaped := p.next()
			if escaped == 0 {
				return nil, p.malformed()
			}
			sb.WriteString(pendingWhitespace.String())
			pendingWhitespace.Reset()
			sb.WriteRune(escaped)
		case quoted && r == '"':
			return p.convert(ctx, sb.String())
		case quoted:
			sb.WriteRune(r)
		case r == ',' || r == '}':
			p.pos--
			if sb.Len() == 0 {
				return nil, p.malformed()
			}
			if strings.EqualFold(sb.String(), "null") {
				return nil, nil
			}
			return p.convert(ctx, sb.String())
		case r == '{' || r == '"':
			return nil, p.malformed()
		case unicode.IsSpace(r):
			pendingWhitespace.WriteRune(r)
		default:
			sb.WriteString(pendingWhitespace.String())
			pendingWhitespace.Reset()
			sb.WriteRune(r)
		}
	}
}

// convert converts the element text to the base type, recording the first conversion error.
func (p *arrayLiteralParser) convert(ctx *sql.Context, str string) (any, error) {
	val, err := p.baseType.IoInput(ctx, str)
	if err != nil && p.elementErr == nil {
		p.elementErr = err
	}
	return val, nil
}

// skipWhitespace advances past any whitespace.
func (p *arrayLiteralParser) skipWhitespace() {
	for p.pos < len(p.runes) && unicode.IsSpace(p.runes[p.pos]) {
		p.pos++
	}
}

// peek returns the current rune without advancing, or zero at the end of the input.
func (p *arrayLiteralParser) peek() rune {
	if p.pos >= len(p.runes) {
		return 0
	}
	return p.runes[p.pos]
}

// next returns the current rune and advances, or returns zero at the end of the input.
func (p *arrayLiteralParser) next() rune {
	r := p.peek()
	p.pos++
	return r
}

// malformed returns the error for an invalid literal.
func (p *arrayLiteralParser) malformed() error {
	return errors.Errorf(`malformed array literal: "%s"`, p.input)
}

// array_out represents the PostgreSQL function of array type IO output.
var array_out = framework.Function1{
	Name:       "array_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		arrType := t[0]
		baseType := arrType.ArrayBaseType()
		return pgtypes.ArrToString(ctx, val.([]any), baseType, false)
	},
}

// array_recv represents the PostgreSQL function of array type IO receive.
var array_recv = framework.Function3{
	Name:       "array_recv",
	Return:     pgtypes.AnyArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable:   array_recv_callable,
}

// array_recv_callable is the function definition of array_recv.
func array_recv_callable(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
	data, err := framework.UnwrapBytes(ctx, val1)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	typeColl, err := core.GetTypesCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	reader := utils.NewWireReader(data)
	dimensions := reader.ReadInt32()
	_ = reader.ReadInt32() // Whether the array has a null, doesn't seem useful
	baseTypeID := id.Type(id.Cache().ToInternal(reader.ReadUint32()))
	baseType, err := typeColl.GetType(ctx, baseTypeID)
	if err != nil {
		return nil, err
	}
	if baseType == nil {
		return nil, pgtypes.ErrTypeDoesNotExist.New(baseTypeID.TypeName())
	}
	dims := make([]int32, dimensions)
	elementsCount := int32(0)
	for dimensionIdx := range dims {
		dims[dimensionIdx] = reader.ReadInt32()
		_ = reader.ReadInt32() // Lower bound, not sure what to do with this
		if dimensionIdx == 0 {
			elementsCount = dims[dimensionIdx]
		} else {
			elementsCount *= dims[dimensionIdx]
		}
	}
	var vals []any
	for i := int32(0); i < elementsCount; i++ {
		elementLen := reader.ReadInt32()
		if elementLen != -1 {
			valBytes := reader.ReadBytes(uint32(elementLen))
			val, err := baseType.CallReceive(ctx, valBytes)
			if err != nil {
				return nil, err
			}
			vals = append(vals, val)
		} else {
			vals = append(vals, nil)
		}
	}
	return pgtypes.InflateArray(vals, dims), nil
}

// array_send represents the PostgreSQL function of array type IO send.
var array_send = framework.Function1{
	Name:       "array_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		if wrapper, ok := val.(sql.AnyWrapper); ok {
			var err error
			val, err = wrapper.UnwrapAny(ctx)
			if err != nil {
				return nil, err
			}
			if val == nil {
				return nil, nil
			}
		}
		dims := pgtypes.ArrayDims(val.([]any), t[0].BaseType())
		vals := pgtypes.FlattenArray(val.([]any), t[0].BaseType())
		// Check for nulls first
		hasNull := false
		for _, val := range vals {
			if val == nil {
				hasNull = true
				break
			}
		}
		writer := utils.NewWireWriter()
		writer.WriteInt32(int32(len(dims))) // Write the number of dimensions
		if hasNull {
			writer.WriteInt32(1)
		} else {
			writer.WriteInt32(0)
		}
		writer.WriteUint32(id.Cache().ToOID(t[0].BaseType().ID.AsId())) // Element OID
		for _, dim := range dims {
			writer.WriteInt32(dim) // Elements in this dimension
			if t[0].IsArrayType() {
				writer.WriteInt32(1) // Lower bound, or what index number we start at (seems to always be 1?)
			} else {
				writer.WriteInt32(0)
			}
		}
		for _, val := range vals {
			if val == nil {
				writer.WriteInt32(-1)
			} else {
				valBytes, err := t[0].BaseType().CallSend(ctx, val)
				if err != nil {
					return nil, err
				}
				writer.WriteInt32(int32(len(valBytes)))
				writer.WriteBytes(valBytes)
			}
		}
		return writer.BufferData(), nil
	},
}

// btarraycmp represents the PostgreSQL function of array type byte compare.
var btarraycmp = framework.Function2{
	Name:       "btarraycmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		at := t[0]
		bt := t[1]
		if !at.Equals(bt) {
			// TODO: currently, types should match.
			// Technically, does not have to e.g.: float4 vs float8
			return nil, errors.Errorf("different type comparison is not supported yet")
		}

		res, err := at.Compare(ctx, val1, val2)
		if err != nil {
			return nil, err
		}
		return int32(res), nil
	},
}

// array_subscript_handler represents the PostgreSQL function of array type subscript handler.
var array_subscript_handler = framework.Function1{
	Name:       "array_subscript_handler",
	Return:     pgtypes.Internal,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return []byte{}, nil
	},
}
