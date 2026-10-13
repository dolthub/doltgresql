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

package functions

import (
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
	"github.com/dolthub/doltgresql/postgres/parser/utils"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	wire "github.com/dolthub/doltgresql/utils"
)

const (
	// inetWireFamilyIPv4 is the address family used for IPv4 in the binary wire format (PGSQL_AF_INET).
	inetWireFamilyIPv4 = 2
	// inetWireFamilyIPv6 is the address family used for IPv6 in the binary wire format (PGSQL_AF_INET6).
	inetWireFamilyIPv6 = 3
)

// initInet registers the functions to the catalog.
func initInet() {
	framework.RegisterFunction(inet_in)
	framework.RegisterFunction(inet_out)
	framework.RegisterFunction(inet_recv)
	framework.RegisterFunction(inet_send)
	framework.RegisterFunction(network_cmp)
	framework.RegisterFunction(abbrev_inet)
	framework.RegisterFunction(broadcast_inet)
	framework.RegisterFunction(family_inet)
	framework.RegisterFunction(host_inet)
	framework.RegisterFunction(hostmask_inet)
	framework.RegisterFunction(inet_same_family)
	framework.RegisterFunction(inetnot)
	framework.RegisterFunction(masklen_inet)
	framework.RegisterFunction(netmask_inet)
	framework.RegisterFunction(set_masklen_inet)
	framework.RegisterFunction(text_inet)
}

// inetMaxMask returns the maximum mask length for the family of the given inet value.
func inetMaxMask(val ipaddr.IPAddr) byte {
	if val.Family == ipaddr.IPv4family {
		return 32
	}
	return 128
}

// inetHost returns the address of the given inet value, without the mask.
func inetHost(val ipaddr.IPAddr) string {
	val.Mask = inetMaxMask(val)
	return val.String()
}

// inet_in represents the PostgreSQL function of inet type IO input.
var inet_in = framework.Function1{
	Name:       "inet_in",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input, err := framework.UnwrapString(ctx, val)
		if err != nil {
			return nil, err
		}
		var ip ipaddr.IPAddr
		if err = ipaddr.ParseINet(input, &ip); err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("inet", input)
		}
		return ip, nil
	},
}

// inet_out represents the PostgreSQL function of inet type IO output.
var inet_out = framework.Function1{
	Name:       "inet_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(ipaddr.IPAddr).String(), nil
	},
}

// inet_recv represents the PostgreSQL function of inet type IO receive.
var inet_recv = framework.Function1{
	Name:       "inet_recv",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data, err := framework.UnwrapBytes(ctx, val)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}
		if len(data) < 4 {
			return nil, errors.Errorf(`invalid length in external "inet" value`)
		}
		var ip ipaddr.IPAddr
		var addrLen int
		switch data[0] {
		case inetWireFamilyIPv4:
			ip.Family = ipaddr.IPv4family
			addrLen = 4
		case inetWireFamilyIPv6:
			ip.Family = ipaddr.IPv6family
			addrLen = 16
		default:
			return nil, errors.Errorf(`invalid address family in external "inet" value`)
		}
		ip.Mask = data[1]
		if ip.Mask > inetMaxMask(ip) {
			return nil, errors.Errorf(`invalid bits in external "inet" value`)
		}
		// data[2] is the is_cidr flag, which we ignore since this is always an inet value
		if int(data[3]) != addrLen || len(data) != 4+addrLen {
			return nil, errors.Errorf(`invalid length in external "inet" value`)
		}
		addrData := data[4:]
		if ip.Family == ipaddr.IPv4family {
			addrData = append([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0xFF, 0xFF}, addrData...)
		}
		ip.Addr = ipaddr.Addr(utils.FromBytes(addrData))
		return ip, nil
	},
}

// inet_send represents the PostgreSQL function of inet type IO send.
var inet_send = framework.Function1{
	Name:       "inet_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		family := byte(inetWireFamilyIPv4)
		if ip.Family == ipaddr.IPv6family {
			family = inetWireFamilyIPv6
		}
		addrData := pgtypes.InetAddrBytes(ip)
		writer := wire.NewWireWriter()
		writer.WriteBytes([]byte{family, ip.Mask, 0, byte(len(addrData))})
		writer.WriteBytes(addrData)
		return writer.BufferData(), nil
	},
}

// network_cmp represents the PostgreSQL function of inet type compare.
var network_cmp = framework.Function2{
	Name:       "network_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return int32(pgtypes.InetCompare(val1.(ipaddr.IPAddr), val2.(ipaddr.IPAddr))), nil
	},
}

// abbrev_inet represents the PostgreSQL function of the same name, taking the same parameters.
var abbrev_inet = framework.Function1{
	Name:       "abbrev",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(ipaddr.IPAddr).String(), nil
	},
}

// broadcast_inet represents the PostgreSQL function of the same name, taking the same parameters.
var broadcast_inet = framework.Function1{
	Name:       "broadcast",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		return ip.Broadcast(), nil
	},
}

// family_inet represents the PostgreSQL function of the same name, taking the same parameters.
var family_inet = framework.Function1{
	Name:       "family",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if val.(ipaddr.IPAddr).Family == ipaddr.IPv4family {
			return int32(4), nil
		}
		return int32(6), nil
	},
}

// host_inet represents the PostgreSQL function of the same name, taking the same parameters.
var host_inet = framework.Function1{
	Name:       "host",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return inetHost(val.(ipaddr.IPAddr)), nil
	},
}

// hostmask_inet represents the PostgreSQL function of the same name, taking the same parameters.
var hostmask_inet = framework.Function1{
	Name:       "hostmask",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		return ip.Hostmask(), nil
	},
}

// inet_same_family represents the PostgreSQL function of the same name, taking the same parameters.
var inet_same_family = framework.Function2{
	Name:       "inet_same_family",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		return val1.(ipaddr.IPAddr).Family == val2.(ipaddr.IPAddr).Family, nil
	},
}

// inetnot represents the PostgreSQL function of the same name, taking the same parameters.
var inetnot = framework.Function1{
	Name:       "inetnot",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		return ip.Complement(), nil
	},
}

// masklen_inet represents the PostgreSQL function of the same name, taking the same parameters.
var masklen_inet = framework.Function1{
	Name:       "masklen",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return int32(val.(ipaddr.IPAddr).Mask), nil
	},
}

// netmask_inet represents the PostgreSQL function of the same name, taking the same parameters.
var netmask_inet = framework.Function1{
	Name:       "netmask",
	Return:     pgtypes.Inet,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		return ip.Netmask(), nil
	},
}

// set_masklen_inet represents the PostgreSQL function of the same name, taking the same parameters.
var set_masklen_inet = framework.Function2{
	Name:       "set_masklen",
	Return:     pgtypes.Inet,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Inet, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ip := val1.(ipaddr.IPAddr)
		bits := val2.(int32)
		if bits == -1 {
			bits = int32(inetMaxMask(ip))
		}
		if bits < 0 || bits > int32(inetMaxMask(ip)) {
			return nil, errors.Errorf("invalid mask length: %d", val2.(int32))
		}
		ip.Mask = byte(bits)
		return ip, nil
	},
}

// text_inet represents the PostgreSQL function of the same name, taking the same parameters.
var text_inet = framework.Function1{
	Name:       "text",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Inet},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		ip := val.(ipaddr.IPAddr)
		return inetHost(ip) + "/" + strconv.Itoa(int(ip.Mask)), nil
	},
}
