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

package binary

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInet registers the inet operators with the function framework.
func initInet() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryEqual, network_eq)
	framework.RegisterBinaryFunction(framework.Operator_BinaryNotEqual, network_ne)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessThan, network_lt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryLessOrEqual, network_le)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterThan, network_gt)
	framework.RegisterBinaryFunction(framework.Operator_BinaryGreaterOrEqual, network_ge)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftLeft, network_sub)
	framework.RegisterBinaryFunction(framework.Operator_BinaryShiftRight, network_sup)
	framework.RegisterBinaryFunction(framework.Operator_BinaryArrayOverlap, network_overlap)
	framework.RegisterBinaryFunction(framework.Operator_BinaryBitAnd, inetand)
	framework.RegisterBinaryFunction(framework.Operator_BinaryBitOr, inetor)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, inetpl)
	framework.RegisterBinaryFunction(framework.Operator_BinaryPlus, int8pl_inet)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, inetmi_int8)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, inetmi)
	// These operators (<<= and >>=) are parsed as function calls, so they're registered as regular functions
	framework.RegisterFunction(network_subeq)
	framework.RegisterFunction(network_supeq)
}

// inetCompareOperator returns a comparison operator function for inet values, using the given function to interpret
// the comparison result.
func inetCompareOperator(name string, cmp func(res int) bool) framework.Function2 {
	return framework.Function2{
		Name:       name,
		Return:     pgtypes.Bool,
		Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
		Strict:     true,
		Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
			return cmp(pgtypes.InetCompare(val1.(ipaddr.IPAddr), val2.(ipaddr.IPAddr))), nil
		},
	}
}

// inetContainmentOperator returns a containment operator function for inet values, using the given function to
// determine containment.
func inetContainmentOperator(name string, contains func(left ipaddr.IPAddr, right *ipaddr.IPAddr) bool) framework.Function2 {
	return framework.Function2{
		Name:       name,
		Return:     pgtypes.Bool,
		Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
		Strict:     true,
		Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
			right := val2.(ipaddr.IPAddr)
			return contains(val1.(ipaddr.IPAddr), &right), nil
		},
	}
}

// network_eq represents the PostgreSQL function of the same name, taking the same parameters.
var network_eq = inetCompareOperator("network_eq", func(res int) bool { return res == 0 })

// network_ne represents the PostgreSQL function of the same name, taking the same parameters.
var network_ne = inetCompareOperator("network_ne", func(res int) bool { return res != 0 })

// network_lt represents the PostgreSQL function of the same name, taking the same parameters.
var network_lt = inetCompareOperator("network_lt", func(res int) bool { return res < 0 })

// network_le represents the PostgreSQL function of the same name, taking the same parameters.
var network_le = inetCompareOperator("network_le", func(res int) bool { return res <= 0 })

// network_gt represents the PostgreSQL function of the same name, taking the same parameters.
var network_gt = inetCompareOperator("network_gt", func(res int) bool { return res > 0 })

// network_ge represents the PostgreSQL function of the same name, taking the same parameters.
var network_ge = inetCompareOperator("network_ge", func(res int) bool { return res >= 0 })

// network_sub represents the PostgreSQL function of the same name, taking the same parameters.
var network_sub = inetContainmentOperator("network_sub", ipaddr.IPAddr.ContainedBy)

// network_subeq represents the PostgreSQL function of the same name, taking the same parameters.
var network_subeq = inetContainmentOperator("network_subeq", ipaddr.IPAddr.ContainedByOrEquals)

// network_sup represents the PostgreSQL function of the same name, taking the same parameters.
var network_sup = inetContainmentOperator("network_sup", ipaddr.IPAddr.Contains)

// network_supeq represents the PostgreSQL function of the same name, taking the same parameters.
var network_supeq = inetContainmentOperator("network_supeq", ipaddr.IPAddr.ContainsOrEquals)

// network_overlap represents the PostgreSQL function of the same name, taking the same parameters.
var network_overlap = inetContainmentOperator("network_overlap", ipaddr.IPAddr.ContainsOrContainedBy)

// inetand represents the PostgreSQL function of the same name, taking the same parameters.
var inetand = framework.Function2{
	Name:       "inetand",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		left, right := val1.(ipaddr.IPAddr), val2.(ipaddr.IPAddr)
		return left.And(&right)
	},
}

// inetor represents the PostgreSQL function of the same name, taking the same parameters.
var inetor = framework.Function2{
	Name:       "inetor",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		left, right := val1.(ipaddr.IPAddr), val2.(ipaddr.IPAddr)
		return left.Or(&right)
	},
}

// inetpl represents the PostgreSQL function of the same name, taking the same parameters.
var inetpl = framework.Function2{
	Name:       "inetpl",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		ip := val1.(ipaddr.IPAddr)
		return ip.Add(val2.(int64))
	},
}

// int8pl_inet represents the PostgreSQL function of the same name, taking the same parameters.
var int8pl_inet = framework.Function2{
	Name:       "int8pl_inet",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		ip := val2.(ipaddr.IPAddr)
		return ip.Add(val1.(int64))
	},
}

// inetmi_int8 represents the PostgreSQL function of the same name, taking the same parameters.
var inetmi_int8 = framework.Function2{
	Name:       "inetmi_int8",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		ip := val1.(ipaddr.IPAddr)
		return ip.Sub(val2.(int64))
	},
}

// inetmi represents the PostgreSQL function of the same name, taking the same parameters.
var inetmi = framework.Function2{
	Name:       "inetmi",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		left, right := val1.(ipaddr.IPAddr), val2.(ipaddr.IPAddr)
		return left.SubIPAddr(&right)
	},
}
