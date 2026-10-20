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

package types

import (
	"bytes"
	"encoding/binary"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
)

// Inet is the IPv4 or IPv6 host address type, along with an optional subnet.
var Inet = &DoltgresType{
	ID:                  toInternal("inet"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_NetworkAddressTypes,
	IsPreferred:         true,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                internalNullType,
	Array:               internalNullType,
	InputFunc:           toFuncID("inet_in", toInternal("cstring")),
	OutputFunc:          toFuncID("inet_out", toInternal("inet")),
	ReceiveFunc:         toFuncID("inet_recv", toInternal("internal")),
	SendFunc:            toFuncID("inet_send", toInternal("inet")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Main,
	NotNull:             false,
	BaseTypeType:        internalNullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("network_cmp", toInternal("inet"), toInternal("inet")),
	SerializationFunc:   serializeTypeInet,
	DeserializationFunc: deserializeTypeInet,
}

// InetCompare compares two inet values using the same ordering as Postgres: the family is compared first, then the
// network bits common to both masks, then the mask length, and finally the full address. Returns -1, 0, or 1.
func InetCompare(a ipaddr.IPAddr, b ipaddr.IPAddr) int {
	if a.Family != b.Family {
		if a.Family < b.Family {
			return -1
		}
		return 1
	}
	aBytes := InetAddrBytes(a)
	bBytes := InetAddrBytes(b)
	if res := inetBitCompare(aBytes, bBytes, int(min(a.Mask, b.Mask))); res != 0 {
		return res
	}
	if a.Mask != b.Mask {
		if a.Mask < b.Mask {
			return -1
		}
		return 1
	}
	return inetBitCompare(aBytes, bBytes, len(aBytes)*8)
}

// InetAddrBytes returns the address bytes of the given inet value, which is 4 bytes for IPv4 and 16 bytes for IPv6.
func InetAddrBytes(val ipaddr.IPAddr) []byte {
	if val.Family == ipaddr.IPv4family {
		data := make([]byte, 4)
		binary.BigEndian.PutUint32(data, uint32(val.Addr.Lo))
		return data
	}
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[:8], val.Addr.Hi)
	binary.BigEndian.PutUint64(data[8:], val.Addr.Lo)
	return data
}

// inetBitCompare compares the leading number of bits of the given byte slices. Returns -1, 0, or 1.
func inetBitCompare(a []byte, b []byte, bits int) int {
	fullBytes := bits / 8
	if res := bytes.Compare(a[:fullBytes], b[:fullBytes]); res != 0 {
		return res
	}
	if remainingBits := bits % 8; remainingBits > 0 {
		mask := byte(0xFF << (8 - remainingBits))
		aByte, bByte := a[fullBytes]&mask, b[fullBytes]&mask
		if aByte < bByte {
			return -1
		} else if aByte > bByte {
			return 1
		}
	}
	return 0
}

// serializeTypeInet handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeInet(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	ip := val.(ipaddr.IPAddr)
	return ip.ToBuffer(nil), nil
}

// deserializeTypeInet handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeInet(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	if len(data) != ipaddr.IPv4size && len(data) != ipaddr.IPv6size {
		return nil, errors.Errorf("invalid serialized inet length: %d", len(data))
	}
	var ip ipaddr.IPAddr
	if _, err := ip.FromBuffer(data); err != nil {
		return nil, err
	}
	return ip, nil
}
