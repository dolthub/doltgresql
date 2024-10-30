package framework

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// NewTextLiteral is the implementation for NewTextLiteral function
// that is being set from expression package to avoid circular dependencies.
var NewTextLiteral func(input string) sql.Expression

// NewLiteral is the implementation for NewLiteral function
// that is being set from expression package to avoid circular dependencies.
var NewLiteral func(input any, t pgtypes.DoltgresType) sql.Expression

func IoInput(ctx *sql.Context, t pgtypes.DoltgresType, input string) (any, error) {
	// TODO: not all ioInput function takes 1 argument of text/cstring, some takes 3 arguments
	inputVal, ok, err := GetFunction(t.InputFunc, NewTextLiteral(input))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(t.InputFunc)
	}
	return inputVal.Eval(ctx, nil)
}

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

func IoReceive(ctx *sql.Context, t pgtypes.DoltgresType, val any) (any, error) {
	rf := t.ReceiveFunc
	if rf == "-" {
		return nil, fmt.Errorf("receive function for type '%s' doesn't exist", t.Name)
	}

	outputVal, ok, err := GetFunction(t.ReceiveFunc, NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", ErrFunctionDoesNotExist.New(t.ReceiveFunc)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return "", err
	}
	return o, nil
}

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
	output, ok := o.([]byte)
	if !ok {
		return nil, fmt.Errorf(`expected byte[], got %T`, output)
	}
	return output, nil
}

// IoCompare might not be the correct name for it? TODO: it seems byte compare?
func IoCompare(ctx *sql.Context, t pgtypes.DoltgresType, v1, v2 any) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	//ac, _, err := t.Convert(v1)
	//if err != nil {
	//	return 0, err
	//}
	//bc, _, err := t.Convert(v2)
	//if err != nil {
	//	return 0, err
	//}
	// TODO: get function name from somewhere?
	return 1, nil
}
