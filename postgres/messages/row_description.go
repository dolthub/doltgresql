// Copyright 2023 Dolthub, Inc.
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

package messages

import (
	"fmt"

	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/postgres/connection"
)

const (
	OidBool              = uint32(16)
	OidBytea             = uint32(17)
	OidChar              = uint32(18)
	OidName              = uint32(19)
	OidInt8              = uint32(20)
	OidInt2              = uint32(21)
	OidInt2Vector        = uint32(22)
	OidInt4              = uint32(23)
	OidRegproc           = uint32(24)
	OidText              = uint32(25)
	OidOid               = uint32(26)
	OidTid               = uint32(27)
	OidXid               = uint32(28)
	OidCid               = uint32(29)
	OidOidVector         = uint32(30)
	OidPgType            = uint32(71)
	OidPgAttribute       = uint32(75)
	OidPgProc            = uint32(81)
	OidPgClass           = uint32(83)
	OidJson              = uint32(114)
	OidXml               = uint32(142)
	OidXmlArray          = uint32(143)
	OidPgNodeTree        = uint32(194)
	OidPgNodeTreeArray   = uint32(195)
	OidJsonArray         = uint32(199)
	OidSmgr              = uint32(210)
	OidIndexAm           = uint32(261)
	OidPoint             = uint32(600)
	OidLseg              = uint32(601)
	OidPath              = uint32(602)
	OidBox               = uint32(603)
	OidPolygon           = uint32(604)
	OidLine              = uint32(628)
	OidCidr              = uint32(650)
	OidCidrArray         = uint32(651)
	OidFloat4            = uint32(700)
	OidFloat8            = uint32(701)
	OidAbstime           = uint32(702)
	OidReltime           = uint32(703)
	OidTinterval         = uint32(704)
	OidUnknown           = uint32(705)
	OidCircle            = uint32(718)
	OidCash              = uint32(790)
	OidMacaddr           = uint32(829)
	OidInet              = uint32(869)
	OidByteaArray        = uint32(1001)
	OidInt2Array         = uint32(1005)
	OidInt4Array         = uint32(1007)
	OidTextArray         = uint32(1009)
	OidVarcharArray      = uint32(1015)
	OidInt8Array         = uint32(1016)
	OidPointArray        = uint32(1017)
	OidFloat4Array       = uint32(1021)
	OidFloat8Array       = uint32(1022)
	OidAclitem           = uint32(1033)
	OidAclitemArray      = uint32(1034)
	OidInetArray         = uint32(1041)
	OidVarchar           = uint32(1043)
	OidDate              = uint32(1082)
	OidTime              = uint32(1083)
	OidTimestamp         = uint32(1114)
	OidTimestampArray    = uint32(1115)
	OidDateArray         = uint32(1182)
	OidTimeArray         = uint32(1183)
	OidInterval          = uint32(1186)
	OidIntervalArray     = uint32(1187)
	OidNumeric           = uint32(1700)
	OidRefcursor         = uint32(1790)
	OidRegprocedure      = uint32(2202)
	OidRegoper           = uint32(2203)
	OidRegoperator       = uint32(2204)
	OidRegclass          = uint32(2205)
	OidRegtype           = uint32(2206)
	OidRegrole           = uint32(4096)
	OidRegnamespace      = uint32(4097)
	OidRegnamespaceArray = uint32(4098)
	OidRegclassArray     = uint32(4099)
	OidRegRoleArray      = uint32(4090)
)

func init() {
	connection.InitializeDefaultMessage(RowDescription{})
}

// RowDescription represents a RowDescription message intended for the client.
type RowDescription struct {
	Fields []pgproto3.FieldDescription
}

var rowDescriptionDefault = connection.MessageFormat{
	Name: "RowDescription",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('T'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Fields",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ColumnName",
						Type: connection.String,
						Data: "",
					},
					{
						Name: "TableObjectID",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "ColumnAttributeNumber",
						Type: connection.Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeObjectID",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "DataTypeSize",
						Type: connection.Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeModifier",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "FormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ connection.Message = RowDescription{}

// Encode implements the interface connection.Message.
func (m RowDescription) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i := 0; i < len(m.Fields); i++ {
		field := m.Fields[i]
		outputMessage.Field("Fields").Child("ColumnName", i).MustWrite(string(field.Name))
		outputMessage.Field("Fields").Child("DataTypeObjectID", i).MustWrite(field.DataTypeOID)
		outputMessage.Field("Fields").Child("DataTypeSize", i).MustWrite(field.DataTypeSize)
		outputMessage.Field("Fields").Child("DataTypeModifier", i).MustWrite(field.TypeModifier)
		outputMessage.Field("Fields").Child("FormatCode", i).MustWrite(field.Format)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m RowDescription) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("RowDescription messages do not support decoding, as they're only sent from the server.")
}

// DefaultMessage implements the interface connection.Message.
func (m RowDescription) DefaultMessage() *connection.MessageFormat {
	return &rowDescriptionDefault
}

// VitessTypeToObjectID returns a type, as defined by Vitess, into a type as defined by Postgres.
// OIDs can be obtained with the following query: `SELECT oid, typname FROM pg_type ORDER BY 1;`
func VitessTypeToObjectID(typ query.Type) (uint32, error) {
	switch typ {
	case query.Type_INT8:
		// Postgres doesn't make use of a small integer type for integer returns, which presents a bit of a conundrum.
		// GMS defines boolean operations as the smallest integer type, while Postgres has an explicit bool type.
		// We can't always assume that `INT8` means bool, since it could just be a small integer. As a result, we'll
		// always return this as though it's an `INT16`, which also means that we can't support bools right now.
		// OIDs 16 (bool) and 18 (char, ASCII only?) are the only single-byte types as far as I'm aware.
		return OidInt2, nil
	case query.Type_INT16:
		// The technically correct OID is 21 (2-byte integer), however it seems like some clients don't actually expect
		// this, so I'm not sure when it's actually used by Postgres. Because of this, we'll just pretend it's an `INT32`.
		return OidInt2, nil
	case query.Type_INT24:
		// Postgres doesn't have a 3-byte integer type, so just pretend it's `INT32`.
		return OidInt4, nil
	case query.Type_INT32:
		return OidInt4, nil
	case query.Type_INT64:
		return OidInt8, nil
	case query.Type_UINT8:
		return OidInt4, nil
	case query.Type_UINT16:
		return OidInt4, nil
	case query.Type_UINT24:
		return OidInt4, nil
	case query.Type_UINT32:
		// Since this has an upperbound greater than `INT32`, we'll treat it as `INT64`
		return OidOid, nil
	case query.Type_UINT64:
		// Since this has an upperbound greater than `INT64`, we'll treat it as `NUMERIC`
		return OidNumeric, nil
	case query.Type_FLOAT32:
		return OidFloat4, nil
	case query.Type_FLOAT64:
		return OidFloat8, nil
	case query.Type_DECIMAL:
		return OidNumeric, nil
	case query.Type_CHAR:
		return OidChar, nil
	case query.Type_VARCHAR:
		return OidVarchar, nil
	case query.Type_TEXT:
		return OidText, nil
	case query.Type_BLOB:
		return OidBytea, nil
	case query.Type_JSON:
		return OidJson, nil
	case query.Type_TIMESTAMP, query.Type_DATETIME:
		return OidTimestamp, nil
	case query.Type_DATE:
		return OidDate, nil
	case query.Type_NULL_TYPE:
		return OidText, nil // NULL is treated as TEXT on the wire
	case query.Type_ENUM:
		return OidText, nil // TODO: temporary solution until we support CREATE TYPE
	default:
		return 0, fmt.Errorf("unsupported type: %s", typ)
	}
}
