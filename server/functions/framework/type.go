package framework

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func IoInput(ctx *sql.Context, t pgtypes.DoltgresType, input string) (any, error) {
	// TODO: ideally, should use NewTextLiteral() -- import cycle issue
	// TODO: not all ioInput function takes 1 argument of text/cstring, some takes 3 arguments
	inputVal, ok, err := GetFunction(t.InputFunc, expression.NewLiteral(input, pgtypes.Text))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(`function "nextval" could not be found for SERIAL default`)
	}
	return inputVal.Eval(ctx, nil)
}

func IoOutput(ctx *sql.Context, t pgtypes.DoltgresType, val any) (string, error) {
	// this is kind of converting before calling `out` function // TODO: check
	v, err := IoReceive(ctx, t, expression.NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	// calling `out` function
	outputVal, ok, err := GetFunction(t.OutputFunc, expression.NewLiteral(v, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf(`function "nextval" could not be found for SERIAL default`)
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
		return nil, fmt.Errorf("recv function for type '%s' doesn't exist", t.Name)
	}

	outputVal, ok, err := GetFunction(t.ReceiveFunc, expression.NewLiteral(val, t))
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf(`function "nextval" could not be found for SERIAL default`)
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
		return nil, fmt.Errorf("recv function for type '%s' doesn't exist", t.Name)
	}

	outputVal, ok, err := GetFunction(t.SendFunc, expression.NewLiteral(val, t))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(`function "nextval" could not be found for SERIAL default`)
	}
	o, err := outputVal.Eval(ctx, nil)
	if err != nil {
		return nil, err
	}
	return o.([]byte), nil
}

//func Compare(ctx *sql.Context, t pgtypes.DoltgresType, v1, v2 any) (int, error) {
//	if v1 == nil && v2 == nil {
//		return 0, nil
//	} else if v1 != nil && v2 == nil {
//		return 1, nil
//	} else if v1 == nil && v2 != nil {
//		return -1, nil
//	}
//
//	ac, _, err := t.Convert(v1)
//	if err != nil {
//		return 0, err
//	}
//	bc, _, err := t.Convert(v2)
//	if err != nil {
//		return 0, err
//	}
//	return t.compareFunc(ac, bc)
//}
