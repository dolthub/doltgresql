package framework

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewTextLiteral is the implementation for NewTextLiteral function
// that is being set from expression package to avoid circular dependencies.
var NewTextLiteral func(input string) sql.Expression

// NewLiteral is the implementation for NewLiteral function
// that is being set from expression package to avoid circular dependencies.
var NewLiteral func(input any, t pgtypes.DoltgresType) sql.Expression

// IoInput converts input string value to given type value.
func IoInput(ctx *sql.Context, t pgtypes.DoltgresType, input string) (any, error) {
	receivedVal := NewTextLiteral(input)
	var cf *CompiledFunction
	var ok bool
	var err error
	if bt, isArray := t.ArrayBaseType(); isArray {
		typmod := int32(0)
		if bt.ModInFunc != "-" {
			typmod = t.AttTypMod
		}
		cf, ok, err = GetFunction(t.InputFunc, receivedVal, NewLiteral(bt.OID, pgtypes.Oid), NewLiteral(typmod, pgtypes.Int32))
	} else if t.TypType == pgtypes.TypeType_Domain {
		oid := t.DomainUnderlyingBaseType().OID
		cf, ok, err = GetFunction(t.InputFunc, receivedVal, NewLiteral(oid, pgtypes.Oid), NewLiteral(t.TypMod, pgtypes.Int32))
	} else if t.ModInFunc != "-" {
		// TODO: there should be better way to check for typmod used
		typmod := t.AttTypMod
		cf, ok, err = GetFunction(t.InputFunc, receivedVal, NewLiteral(t.OID, pgtypes.Oid), NewLiteral(typmod, pgtypes.Int32))
	} else {
		cf, ok, err = GetFunction(t.InputFunc, receivedVal)
	}
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(t.InputFunc)
	}
	return cf.Eval(ctx, nil)
}

// IoOutput converts given type value to output string.
func IoOutput(ctx *sql.Context, t pgtypes.DoltgresType, val any) (string, error) {
	// calling `out` function
	outputVal, ok, err := GetFunction(t.OutputFunc, NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.OutputFunc)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if !ok {
		return "", fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// IoReceive converts external binary format (which is a byte array) to given type value.
func IoReceive(ctx *sql.Context, t pgtypes.DoltgresType, val any) (any, error) {
	rf := t.ReceiveFunc
	if rf == "-" {
		return nil, fmt.Errorf("receive function for type '%s' doesn't exist", t.Name)
	}

	receivedVal := NewLiteral(val, pgtypes.NewInternalTypeWithBaseType(t.OID))

	var cf *CompiledFunction
	var ok bool
	var err error
	if t.ModInFunc != "-" {
		// TODO: there should be better way to check for typmod used
		typmod := t.AttTypMod
		cf, ok, err = GetFunction(t.ReceiveFunc, receivedVal, NewLiteral(t.OID, pgtypes.Oid), NewLiteral(typmod, pgtypes.Int32))
	} else if t.TypType == pgtypes.TypeType_Domain {
		// TODO: if domain type, send underlyting base type OID
		cf, ok, err = GetFunction(t.ReceiveFunc, receivedVal, NewLiteral(t.OID, pgtypes.Oid), NewLiteral(t.TypMod, pgtypes.Int32))
	} else if bt, isArray := t.ArrayBaseType(); isArray {
		typmod := int32(0)
		if bt.ModInFunc != "-" {
			typmod = t.AttTypMod
		}
		cf, ok, err = GetFunction(t.ReceiveFunc, receivedVal, NewLiteral(bt.OID, pgtypes.Oid), NewLiteral(typmod, pgtypes.Int32))
	} else {
		cf, ok, err = GetFunction(t.ReceiveFunc, receivedVal)
	}
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.ReceiveFunc)
	}
	o, err := cf.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	return o, nil
}

// IoSend converts given type value to a byte array.
func IoSend(ctx *sql.Context, t pgtypes.DoltgresType, val any) ([]byte, error) {
	rf := t.SendFunc
	if rf == "-" {
		return nil, fmt.Errorf("send function for type '%s' doesn't exist", t.Name)
	}

	outputVal, ok, err := GetFunction(t.SendFunc, NewLiteral(val, t))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(t.SendFunc)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}
	output, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf(`expected []byte, got %T`, output)
	}
	return output, nil
}

// TypModIn encodes given text array value to type modifier in int32 format.
func TypModIn(ctx *sql.Context, t pgtypes.DoltgresType, val []any) (any, error) {
	// takes []string and return int32
	if t.ModInFunc == "-" {
		return nil, fmt.Errorf("typmodin function for type '%s' doesn't exist", t.Name)
	}
	v, ok, err := GetFunction(t.ModInFunc, NewLiteral(val, pgtypes.TextArray))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(t.ModInFunc)
	}
	return v.Eval(ctx, nil)
}

// TypModOut decodes type modifier in int32 format to string representation of it.
func TypModOut(ctx *sql.Context, t pgtypes.DoltgresType, val int32) (any, error) {
	// takes int32 and returns string
	if t.ModOutFunc != "-" {
		return nil, fmt.Errorf("typmodout function for type '%s' doesn't exist", t.Name)
	}
	v, ok, err := GetFunction(t.ModOutFunc, NewLiteral(val, pgtypes.Int32))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(t.ModOutFunc)
	}
	o, err := v.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	if o == nil {
		return nil, nil
	}
	output, ok := o.(string)
	if !ok {
		return nil, fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// IoCompare compares given two values using the given type. // TODO: both values should have types. e.g. compare between float32 and float64
func IoCompare(ctx *sql.Context, t pgtypes.DoltgresType, v1, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	// TODO: get base type
	f, ok := temporaryTypeToCompareFunctionMapping[t.OID]
	if !ok {
		return 0, fmt.Errorf("compare function does not exist for %s type", t.Name)
	}

	v, ok, err := GetFunction(f, NewLiteral(v1, t), NewLiteral(v2, t))
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrFunctionDoesNotExist.New(f)
	}

	i, err := v.Eval(ctx, nil)
	if err != nil {
		return 0, err
	}
	return int(i.(int32)), nil
}

var temporaryTypeToCompareFunctionMapping = map[uint32]string{
	pgtypes.Bool.OID:         "btboolcmp",
	pgtypes.AnyArray.OID:     "btarraycmp",
	pgtypes.BpChar.OID:       "bpcharcmp",
	pgtypes.Bytea.OID:        "byteacmp",
	pgtypes.Date.OID:         "date_cmp",
	pgtypes.Float32.OID:      "btfloat4cmp", // TODO: btfloat48cmp is for float32 vs float64
	pgtypes.Float64.OID:      "btfloat8cmp", // TODO
	pgtypes.Int16.OID:        "btint2cmp",   // TODO
	pgtypes.Int32.OID:        "btint4cmp",   // TODO
	pgtypes.Int64.OID:        "btint8cmp",   // TODO
	pgtypes.InternalChar.OID: "btcharcmp",
	pgtypes.Interval.OID:     "interval_cmp",
	pgtypes.JsonB.OID:        "jsonb_cmp",
	pgtypes.Name.OID:         "btnamecmp", // TODO
	pgtypes.Numeric.OID:      "numeric_cmp",
	pgtypes.Oid.OID:          "btoidcmp",
	pgtypes.Text.OID:         "bttextcmp", // TODO
	pgtypes.Time.OID:         "time_cmp",
	pgtypes.Timestamp.OID:    "timestamp_cmp",
	pgtypes.TimestampTZ.OID:  "timestamptz_cmp",
	pgtypes.TimeTZ.OID:       "timetz_cmp",
	pgtypes.Uuid.OID:         "uuid_cmp",
	pgtypes.VarChar.OID:      "bttextcmp", // TODO: if there is no cmp function for the type, use preferred type's cmp function?
}

// SQL converts given type value to output string.
func SQL(ctx *sql.Context, t pgtypes.DoltgresType, val any) (string, error) {
	if bt, isArray := t.ArrayBaseType(); isArray {
		if bt.ModInFunc != "-" {
			bt.AttTypMod = t.AttTypMod
		}
		return ArrToString(ctx, val.([]any), bt, true)
	}
	// calling `out` function
	outputVal, ok, err := GetFunction(t.OutputFunc, NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.OutputFunc)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if t.OID == uint32(oid.T_bool) {
		output = string(output[0])
	}
	if !ok {
		return "", fmt.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

func ArrToString(ctx *sql.Context, arr []any, baseType pgtypes.DoltgresType, trimBool bool) (string, error) {
	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(",")
		}
		if v != nil {
			str, err := IoOutput(ctx, baseType, v)
			if err != nil {
				return "", err
			}
			if baseType.OID == uint32(oid.T_bool) && trimBool {
				str = string(str[0])
			}
			shouldQuote := false
			for _, r := range str {
				switch r {
				case ' ', ',', '{', '}', '\\', '"':
					shouldQuote = true
				}
			}
			if shouldQuote || strings.EqualFold(str, "NULL") {
				sb.WriteRune('"')
				sb.WriteString(strings.ReplaceAll(str, `"`, `\"`))
				sb.WriteRune('"')
			} else {
				sb.WriteString(str)
			}
		} else {
			sb.WriteString("NULL")
		}
	}
	sb.WriteRune('}')
	return sb.String(), nil
}
