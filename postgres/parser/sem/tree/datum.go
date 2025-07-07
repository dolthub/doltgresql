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

// Copyright 2015 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package tree

import (
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/cockroachdb/apd/v2"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/geo"
	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
	"github.com/dolthub/doltgresql/postgres/parser/json"
	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgdate"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/stringencoding"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/postgres/parser/utils"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
)

var (
	constDBoolTrue  DBool = true
	constDBoolFalse DBool = false

	// DBoolTrue is a pointer to the DBool(true) value and can be used in
	// comparisons against Datum types.
	DBoolTrue = &constDBoolTrue
	// DBoolFalse is a pointer to the DBool(false) value and can be used in
	// comparisons against Datum types.
	DBoolFalse = &constDBoolFalse

	// DNull is the NULL Datum.
	DNull Datum = NullLiteral{}

	// DTimeMaxTimeRegex is a compiled regex for parsing the 24:00 time value.
	DTimeMaxTimeRegex = regexp.MustCompile(`^([0-9-]*(\s|T))?\s*24:00(:00(.0+)?)?\s*$`)

	// The maximum timestamp Golang can represents is represented as UNIX
	// time time.Unix(-9223372028715321601, 0).
	// However, this causes errors as we cannot reliably sort as we use
	// UNIX time in the key encoding, and 9223372036854775807 > -9223372028715321601
	// but time.Unix(9223372036854775807, 0) < time.Unix(-9223372028715321601, 0).
	//
	// To be compatible with pgwire, we only support the published min/max for
	// postgres 4714 BC (JULIAN = 0) - 4713 in their docs - and 294276 AD.

	// MaxSupportedTime is the maximum time we support parsing.
	MaxSupportedTime = time.Unix(9224318016000-1, 999999000).UTC() // 294276-12-31 23:59:59.999999
	// MinSupportedTime is the minimum time we support parsing.
	MinSupportedTime = time.Unix(-210866803200, 0).UTC() // 4714-11-24 00:00:00+00 BC

	// LibPQTimePrefix is the prefix lib/pq prints time-type datatypes with.
	LibPQTimePrefix = "0000-01-01"
)

// Datum represents a SQL value.
type Datum interface {
	TypedExpr

	// AmbiguousFormat indicates whether the result of formatting this Datum can
	// be interpreted into more than one type. Used with
	// fmtFlags.disambiguateDatumTypes.
	AmbiguousFormat() bool

	// Size returns a lower bound on the total size of the receiver in bytes,
	// including memory that is pointed at (even if shared between Datum
	// instances) but excluding allocation overhead.
	//
	// It holds for every Datum d that d.Size().
	Size() uintptr
}

// Datums is a slice of Datum values.
type Datums []Datum

// Len returns the number of Datum values.
func (d Datums) Len() int { return len(d) }

// Format implements the NodeFormatter interface.
func (d *Datums) Format(ctx *FmtCtx) {
	ctx.WriteByte('(')
	for i, v := range *d {
		if i > 0 {
			ctx.WriteString(", ")
		}
		ctx.FormatNode(v)
	}
	ctx.WriteByte(')')
}

// CompositeDatum is a Datum that may require composite encoding in
// indexes. Any Datum implementing this interface must also add itself to
// colinfo.HasCompositeKeyEncoding.
type CompositeDatum interface {
	Datum
	// IsComposite returns true if this datum is not round-tripable in a key
	// encoding.
	IsComposite() bool
}

// DBool is the boolean Datum.
type DBool bool

// MakeDBool converts its argument to a *DBool, returning either DBoolTrue or
// DBoolFalse.
func MakeDBool(d DBool) *DBool {
	if d {
		return DBoolTrue
	}
	return DBoolFalse
}

// makeParseError returns a parse error using the provided string and type. An
// optional error can be provided, which will be appended to the end of the
// error string.
func makeParseError(s string, typ *types.T, err error) error {
	if err != nil {
		return pgerror.Wrapf(err, pgcode.InvalidTextRepresentation,
			"could not parse %q as type %s", s, typ)
	}
	return pgerror.Newf(pgcode.InvalidTextRepresentation,
		"could not parse %q as type %s", s, typ)
}

func isCaseInsensitivePrefix(prefix, s string) bool {
	if len(prefix) > len(s) {
		return false
	}
	return strings.EqualFold(prefix, s[:len(prefix)])
}

// ParseDBool parses and returns the *DBool Datum value represented by the provided
// string, or an error if parsing is unsuccessful.
// See https://github.com/postgres/postgres/blob/90627cf98a8e7d0531789391fd798c9bfcc3bc1a/src/backend/utils/adt/bool.c#L36
func ParseDBool(s string) (*DBool, error) {
	s = strings.TrimSpace(s)
	if len(s) >= 1 {
		switch s[0] {
		case 't', 'T':
			if isCaseInsensitivePrefix(s, "true") {
				return DBoolTrue, nil
			}
		case 'f', 'F':
			if isCaseInsensitivePrefix(s, "false") {
				return DBoolFalse, nil
			}
		case 'y', 'Y':
			if isCaseInsensitivePrefix(s, "yes") {
				return DBoolTrue, nil
			}
		case 'n', 'N':
			if isCaseInsensitivePrefix(s, "no") {
				return DBoolFalse, nil
			}
		case '1':
			if s == "1" {
				return DBoolTrue, nil
			}
		case '0':
			if s == "0" {
				return DBoolFalse, nil
			}
		case 'o', 'O':
			// Just 'o' is ambiguous between 'on' and 'off'.
			if len(s) > 1 {
				if isCaseInsensitivePrefix(s, "on") {
					return DBoolTrue, nil
				}
				if isCaseInsensitivePrefix(s, "off") {
					return DBoolFalse, nil
				}
			}
		}
	}
	return nil, makeParseError(s, types.Bool, pgerror.New(pgcode.InvalidTextRepresentation, "invalid bool value"))
}

// ParseDByte parses a string representation of hex encoded binary
// data. It supports both the hex format, with "\x" followed by a
// string of hexadecimal digits (the "\x" prefix occurs just once at
// the beginning), and the escaped format, which supports "\\" and
// octal escapes.
func ParseDByte(s string) (*DBytes, error) {
	res, err := lex.DecodeRawBytesToByteArrayAuto([]byte(s))
	if err != nil {
		return nil, makeParseError(s, types.Bytes, err)
	}
	return NewDBytes(DBytes(res)), nil
}

// ParseDUuidFromString parses and returns the *DUuid Datum value represented
// by the provided input string, or an error.
func ParseDUuidFromString(s string) (*DUuid, error) {
	uv, err := uuid.FromString(s)
	if err != nil {
		return nil, makeParseError(s, types.Uuid, err)
	}
	return NewDUuid(DUuid{uv}), nil
}

// ParseDUuidFromBytes parses and returns the *DUuid Datum value represented
// by the provided input bytes, or an error.
func ParseDUuidFromBytes(b []byte) (*DUuid, error) {
	uv, err := uuid.FromBytes(b)
	if err != nil {
		return nil, makeParseError(string(b), types.Uuid, err)
	}
	return NewDUuid(DUuid{uv}), nil
}

// ParseDIPAddrFromINetString parses and returns the *DIPAddr Datum value
// represented by the provided input INet string, or an error.
func ParseDIPAddrFromINetString(s string) (*DIPAddr, error) {
	var d DIPAddr
	err := ipaddr.ParseINet(s, &d.IPAddr)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DBool) ResolvedType() *types.T {
	return types.Bool
}

// AmbiguousFormat implements the Datum interface.
func (*DBool) AmbiguousFormat() bool { return false }

// Format implements the NodeFormatter interface.
func (d *DBool) Format(ctx *FmtCtx) {
	if ctx.HasFlags(fmtPgwireFormat) {
		if bool(*d) {
			ctx.WriteByte('t')
		} else {
			ctx.WriteByte('f')
		}
		return
	}
	ctx.WriteString(strconv.FormatBool(bool(*d)))
}

// Size implements the Datum interface.
func (d *DBool) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DBitArray is the BIT/VARBIT Datum.
type DBitArray struct {
	utils.BitArray
}

// ParseDBitArray parses a string representation of binary digits.
func ParseDBitArray(s string) (*DBitArray, error) {
	var a DBitArray
	var err error
	a.BitArray, err = utils.Parse(s)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

var errCannotCastNegativeIntToBitArray = pgerror.Newf(pgcode.CannotCoerce,
	"cannot cast negative integer to bit varying with unbounded width")

// NewDBitArrayFromInt creates a bit array from the specified integer
// at the specified width.
// If the width is zero, only positive integers can be converted.
// If the width is nonzero, the value is truncated to that width.
// Negative values are encoded using two's complement.
func NewDBitArrayFromInt(i int64, width uint) (*DBitArray, error) {
	if width == 0 && i < 0 {
		return nil, errCannotCastNegativeIntToBitArray
	}
	return &DBitArray{
		BitArray: utils.MakeBitArrayFromInt64(width, i, 64),
	}, nil
}

// AsDInt computes the integer value of the given bit array.
// The value is assumed to be encoded using two's complement.
// The result is truncated to the given integer number of bits,
// if specified.
// The given width must be 64 or smaller. The results are undefined
// if n is greater than 64.
func (d *DBitArray) AsDInt(n uint) *DInt {
	if n == 0 {
		n = 64
	}
	return NewDInt(DInt(d.BitArray.AsInt64(n)))
}

// ResolvedType implements the TypedExpr interface.
func (*DBitArray) ResolvedType() *types.T {
	return types.VarBit
}

// AmbiguousFormat implements the Datum interface.
func (*DBitArray) AmbiguousFormat() bool { return false }

// Format implements the NodeFormatter interface.
func (d *DBitArray) Format(ctx *FmtCtx) {
	f := ctx.flags
	if f.HasFlags(fmtPgwireFormat) {
		d.BitArray.Format(&ctx.Buffer)
	} else {
		withQuotes := !f.HasFlags(FmtFlags(lex.EncBareStrings))
		if withQuotes {
			ctx.WriteString("B'")
		}
		d.BitArray.Format(&ctx.Buffer)
		if withQuotes {
			ctx.WriteByte('\'')
		}
	}
}

// Size implements the Datum interface.
func (d *DBitArray) Size() uintptr {
	return d.BitArray.Sizeof()
}

// DInt is the int Datum.
type DInt int64

// NewDInt is a helper routine to create a *DInt initialized from its argument.
func NewDInt(d DInt) *DInt {
	return &d
}

// ParseDInt parses and returns the *DInt Datum value represented by the provided
// string, or an error if parsing is unsuccessful.
func ParseDInt(s string) (*DInt, error) {
	i, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return nil, makeParseError(s, types.Int, err)
	}
	return NewDInt(DInt(i)), nil
}

// ResolvedType implements the TypedExpr interface.
func (*DInt) ResolvedType() *types.T {
	return types.Int
}

// AmbiguousFormat implements the Datum interface.
func (*DInt) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DInt) Format(ctx *FmtCtx) {
	// If the number is negative, we need to use parens or the `:::INT` type hint
	// will take precedence over the negation sign.
	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	needParens := (disambiguate || parsable) && *d < 0
	if needParens {
		ctx.WriteByte('(')
	}
	ctx.WriteString(strconv.FormatInt(int64(*d), 10))
	if needParens {
		ctx.WriteByte(')')
	}
}

// Size implements the Datum interface.
func (d *DInt) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DFloat is the float Datum.
type DFloat float64

// NewDFloat is a helper routine to create a *DFloat initialized from its
// argument.
func NewDFloat(d DFloat) *DFloat {
	return &d
}

// ParseDFloat parses and returns the *DFloat Datum value represented by the provided
// string, or an error if parsing is unsuccessful.
func ParseDFloat(s string) (*DFloat, error) {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil, makeParseError(s, types.Float, err)
	}
	return NewDFloat(DFloat(f)), nil
}

// ResolvedType implements the TypedExpr interface.
func (*DFloat) ResolvedType() *types.T {
	return types.Float
}

// AmbiguousFormat implements the Datum interface.
func (*DFloat) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DFloat) Format(ctx *FmtCtx) {
	fl := float64(*d)

	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	quote := parsable && (math.IsNaN(fl) || math.IsInf(fl, 0))
	// We need to use Signbit here and not just fl < 0 because of -0.
	needParens := !quote && (disambiguate || parsable) && math.Signbit(fl)
	// If the number is negative, we need to use parens or the `:::INT` type hint
	// will take precedence over the negation sign.
	if quote {
		ctx.WriteByte('\'')
	} else if needParens {
		ctx.WriteByte('(')
	}
	if _, frac := math.Modf(fl); frac == 0 && -1000000 < *d && *d < 1000000 {
		// d is a small whole number. Ensure it is printed using a decimal point.
		ctx.Printf("%.1f", fl)
	} else {
		ctx.Printf("%g", fl)
	}
	if quote {
		ctx.WriteByte('\'')
	} else if needParens {
		ctx.WriteByte(')')
	}
}

// Size implements the Datum interface.
func (d *DFloat) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// IsComposite implements the CompositeDatum interface.
func (d *DFloat) IsComposite() bool {
	// -0 is composite.
	return math.Float64bits(float64(*d)) == 1<<63
}

// DDecimal is the decimal Datum.
type DDecimal struct {
	apd.Decimal
}

// ParseDDecimal parses and returns the *DDecimal Datum value represented by the
// provided string, or an error if parsing is unsuccessful.
func ParseDDecimal(s string) (*DDecimal, error) {
	dd := &DDecimal{}
	err := dd.SetString(s)
	return dd, err
}

// SetString sets d to s. Any non-standard NaN values are converted to a
// normal NaN. Any negative zero is converted to positive.
func (d *DDecimal) SetString(s string) error {
	// ExactCtx should be able to handle any decimal, but if there is any rounding
	// or other inexact conversion, it will result in an error.
	//_, res, err := HighPrecisionCtx.SetString(&d.Decimal, s)
	_, res, err := ExactCtx.SetString(&d.Decimal, s)
	if res != 0 || err != nil {
		return makeParseError(s, types.Decimal, nil)
	}
	switch d.Form {
	case apd.NaNSignaling:
		d.Form = apd.NaN
		d.Negative = false
	case apd.NaN:
		d.Negative = false
	case apd.Finite:
		if d.IsZero() && d.Negative {
			d.Negative = false
		}
	}
	return nil
}

// ResolvedType implements the TypedExpr interface.
func (*DDecimal) ResolvedType() *types.T {
	return types.Decimal
}

// AmbiguousFormat implements the Datum interface.
func (*DDecimal) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DDecimal) Format(ctx *FmtCtx) {
	// If the number is negative, we need to use parens or the `:::INT` type hint
	// will take precedence over the negation sign.
	disambiguate := ctx.flags.HasFlags(fmtDisambiguateDatumTypes)
	parsable := ctx.flags.HasFlags(FmtParsableNumerics)
	quote := parsable && d.Decimal.Form != apd.Finite
	needParens := !quote && (disambiguate || parsable) && d.Negative
	if needParens {
		ctx.WriteByte('(')
	}
	if quote {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.Decimal.String())
	if quote {
		ctx.WriteByte('\'')
	}
	if needParens {
		ctx.WriteByte(')')
	}
}

// SizeOfDecimal returns the size in bytes of an apd.Decimal.
func SizeOfDecimal(d apd.Decimal) uintptr {
	return uintptr(cap(d.Coeff.Bits())) * unsafe.Sizeof(big.Word(0))
}

// Size implements the Datum interface.
func (d *DDecimal) Size() uintptr {
	return unsafe.Sizeof(*d) + SizeOfDecimal(d.Decimal)
}

var (
	decimalNegativeZero = &apd.Decimal{Negative: true}
	bigTen              = big.NewInt(10)
)

// IsComposite implements the CompositeDatum interface.
func (d *DDecimal) IsComposite() bool {
	// -0 is composite.
	if d.Decimal.CmpTotal(decimalNegativeZero) == 0 {
		return true
	}

	// Check if d is divisible by 10.
	var r big.Int
	r.Rem(&d.Decimal.Coeff, bigTen)
	return r.Sign() == 0
}

// DString is the string Datum.
type DString string

// NewDString is a helper routine to create a *DString initialized from its
// argument.
func NewDString(d string) *DString {
	r := DString(d)
	return &r
}

// ResolvedType implements the TypedExpr interface.
func (*DString) ResolvedType() *types.T {
	return types.String
}

// AmbiguousFormat implements the Datum interface.
func (*DString) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DString) Format(ctx *FmtCtx) {
	buf, f := &ctx.Buffer, ctx.flags
	if f.HasFlags(fmtRawStrings) {
		buf.WriteString(string(*d))
	} else {
		lex.EncodeSQLStringWithFlags(buf, string(*d), f.EncodeFlags())
	}
}

// Size implements the Datum interface.
func (d *DString) Size() uintptr {
	return unsafe.Sizeof(*d) + uintptr(len(*d))
}

// DCollatedString is the Datum for strings with a locale. The struct members
// are intended to be immutable.
type DCollatedString struct {
	Contents string
	Locale   string
	// Key is the collation key.
	Key []byte
}

// AmbiguousFormat implements the Datum interface.
func (*DCollatedString) AmbiguousFormat() bool { return false }

// Format implements the NodeFormatter interface.
func (d *DCollatedString) Format(ctx *FmtCtx) {
	lex.EncodeSQLString(&ctx.Buffer, d.Contents)
	ctx.WriteString(" COLLATE ")
	lex.EncodeLocaleName(&ctx.Buffer, d.Locale)
}

// ResolvedType implements the TypedExpr interface.
func (d *DCollatedString) ResolvedType() *types.T {
	return types.MakeCollatedString(types.String, d.Locale)
}

// Size implements the Datum interface.
func (d *DCollatedString) Size() uintptr {
	return unsafe.Sizeof(*d) + uintptr(len(d.Contents)) + uintptr(len(d.Locale)) + uintptr(len(d.Key))
}

// IsComposite implements the CompositeDatum interface.
func (d *DCollatedString) IsComposite() bool {
	return true
}

// DBytes is the bytes Datum. The underlying type is a string because we want
// the immutability, but this may contain arbitrary bytes.
type DBytes string

// NewDBytes is a helper routine to create a *DBytes initialized from its
// argument.
func NewDBytes(d DBytes) *DBytes {
	return &d
}

// ResolvedType implements the TypedExpr interface.
func (*DBytes) ResolvedType() *types.T {
	return types.Bytes
}

// AmbiguousFormat implements the Datum interface.
func (*DBytes) AmbiguousFormat() bool { return true }

func writeAsHexString(ctx *FmtCtx, d *DBytes) {
	b := string(*d)
	for i := 0; i < len(b); i++ {
		ctx.Write(stringencoding.RawHexMap[b[i]])
	}
}

// Format implements the NodeFormatter interface.
func (d *DBytes) Format(ctx *FmtCtx) {
	f := ctx.flags
	if f.HasFlags(fmtPgwireFormat) {
		ctx.WriteString(`"\\x`)
		writeAsHexString(ctx, d)
		ctx.WriteString(`"`)
	} else {
		withQuotes := !f.HasFlags(FmtFlags(lex.EncBareStrings))
		if withQuotes {
			if f.HasFlags(fmtFormatByteLiterals) {
				ctx.WriteByte('b')
			}
			ctx.WriteByte('\'')
		}
		ctx.WriteString("\\x")
		writeAsHexString(ctx, d)
		if withQuotes {
			ctx.WriteByte('\'')
		}
	}
}

// Size implements the Datum interface.
func (d *DBytes) Size() uintptr {
	return unsafe.Sizeof(*d) + uintptr(len(*d))
}

// DUuid is the UUID Datum.
type DUuid struct {
	uuid.UUID
}

// NewDUuid is a helper routine to create a *DUuid initialized from its
// argument.
func NewDUuid(d DUuid) *DUuid {
	return &d
}

// ResolvedType implements the TypedExpr interface.
func (*DUuid) ResolvedType() *types.T {
	return types.Uuid
}

// AmbiguousFormat implements the Datum interface.
func (*DUuid) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DUuid) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.UUID.String())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DUuid) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DIPAddr is the IPAddr Datum.
type DIPAddr struct {
	ipaddr.IPAddr
}

// ResolvedType implements the TypedExpr interface.
func (*DIPAddr) ResolvedType() *types.T {
	return types.INet
}

// AmbiguousFormat implements the Datum interface.
func (*DIPAddr) AmbiguousFormat() bool {
	return true
}

// Format implements the NodeFormatter interface.
func (d *DIPAddr) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.IPAddr.String())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DIPAddr) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DDate is the date Datum represented as the number of days after
// the Unix epoch.
type DDate struct {
	pgdate.Date
}

// NewDDate is a helper routine to create a *DDate initialized from its
// argument.
func NewDDate(d pgdate.Date) *DDate {
	return &DDate{Date: d}
}

// ParseTimeContext provides the information necessary for
// parsing dates, times, and timestamps. A nil value is generally
// acceptable and will result in reasonable defaults being applied.
type ParseTimeContext interface {
	// GetRelativeParseTime returns the transaction time in the session's
	// timezone (i.e. now()). This is used to calculate relative dates,
	// like "tomorrow", and also provides a default time.Location for
	// parsed times.
	GetRelativeParseTime() time.Time
}

var _ ParseTimeContext = &simpleParseTimeContext{}

type simpleParseTimeContext struct {
	RelativeParseTime time.Time
}

// GetRelativeParseTime implements ParseTimeContext.
func (ctx simpleParseTimeContext) GetRelativeParseTime() time.Time {
	return ctx.RelativeParseTime
}

// relativeParseTime chooses a reasonable "now" value for
// performing date parsing.
func relativeParseTime(ctx ParseTimeContext) time.Time {
	if ctx == nil {
		return time.Now()
	}
	return ctx.GetRelativeParseTime()
}

// ParseDDate parses and returns the *DDate Datum value represented by the provided
// string in the provided location, or an error if parsing is unsuccessful.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseDDate(ctx ParseTimeContext, s string) (_ *DDate, dependsOnContext bool, _ error) {
	now := relativeParseTime(ctx)
	t, dependsOnContext, err := pgdate.ParseDate(now, 0 /* mode */, s)
	return NewDDate(t), dependsOnContext, err
}

// ResolvedType implements the TypedExpr interface.
func (*DDate) ResolvedType() *types.T {
	return types.Date
}

// AmbiguousFormat implements the Datum interface.
func (*DDate) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DDate) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	d.Date.Format(&ctx.Buffer)
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DDate) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DTime is the time Datum.
type DTime timeofday.TimeOfDay

// MakeDTime creates a DTime from a TimeOfDay.
func MakeDTime(t timeofday.TimeOfDay) *DTime {
	d := DTime(t)
	return &d
}

// ParseDTime parses and returns the *DTime Datum value represented by the
// provided string, or an error if parsing is unsuccessful.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseDTime(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTime, dependsOnContext bool, _ error) {
	now := relativeParseTime(ctx)

	// Special case on 24:00 and 24:00:00 as the parser
	// does not handle these correctly.
	if DTimeMaxTimeRegex.MatchString(s) {
		return MakeDTime(timeofday.Time2400), false, nil
	}

	if strings.HasPrefix(s, LibPQTimePrefix) {
		s = "1970-01-01" + s[len(LibPQTimePrefix):]
	}

	t, dependsOnContext, err := pgdate.ParseTimeWithoutTimezone(now, pgdate.ParseModeYMD, s)
	if err != nil {
		// Build our own error message to avoid exposing the dummy date.
		return nil, false, makeParseError(s, types.Time, nil)
	}
	return MakeDTime(timeofday.FromTime(t).Round(precision)), dependsOnContext, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DTime) ResolvedType() *types.T {
	return types.Time
}

// Round returns a new DTime to the specified precision.
func (d *DTime) Round(precision time.Duration) *DTime {
	return MakeDTime(timeofday.TimeOfDay(*d).Round(precision))
}

// AmbiguousFormat implements the Datum interface.
func (*DTime) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DTime) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(timeofday.TimeOfDay(*d).String())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DTime) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DTimeTZ is the time with time zone Datum.
type DTimeTZ struct {
	timetz.TimeTZ
}

// NewDTimeTZ creates a DTimeTZ from a timetz.TimeTZ.
func NewDTimeTZ(t timetz.TimeTZ) *DTimeTZ {
	return &DTimeTZ{t}
}

// NewDTimeTZFromOffset creates a DTimeTZ from a TimeOfDay and offset.
func NewDTimeTZFromOffset(t timeofday.TimeOfDay, offsetSecs int32) *DTimeTZ {
	return &DTimeTZ{timetz.MakeTimeTZ(t, offsetSecs)}
}

// ParseDTimeTZ parses and returns the *DTime Datum value represented by the
// provided string, or an error if parsing is unsuccessful. The |loc| used
// to set the timezone used to parse the time value.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseDTimeTZ(ctx ParseTimeContext, s string, precision time.Duration) (_ *DTimeTZ, dependsOnContext bool, _ error) {
	now := relativeParseTime(ctx)
	d, dependsOnContext, err := timetz.ParseTimeTZ(now, s, precision)
	if err != nil {
		return nil, false, err
	}
	return NewDTimeTZ(d), dependsOnContext, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DTimeTZ) ResolvedType() *types.T {
	return types.TimeTZ
}

// Round returns a new DTimeTZ to the specified precision.
func (d *DTimeTZ) Round(precision time.Duration) *DTimeTZ {
	return NewDTimeTZ(d.TimeTZ.Round(precision))
}

// AmbiguousFormat implements the Datum interface.
func (*DTimeTZ) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DTimeTZ) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.TimeTZ.String())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DTimeTZ) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DTimestamp is the timestamp Datum.
type DTimestamp struct {
	// Time always has UTC location.
	time.Time
}

// MakeDTimestamp creates a DTimestamp with specified precision.
func MakeDTimestamp(t time.Time, precision time.Duration) (*DTimestamp, error) {
	ret := t.Round(precision)
	if ret.After(MaxSupportedTime) || ret.Before(MinSupportedTime) {
		return nil, errors.Newf("timestamp out of range: %q", ret.Format(time.RFC3339))
	}
	return &DTimestamp{Time: ret}, nil
}

// MustMakeDTimestamp wraps MakeDTimestamp but panics if there is an error.
// This is intended for testing applications only.
func MustMakeDTimestamp(t time.Time, precision time.Duration) *DTimestamp {
	ret, err := MakeDTimestamp(t, precision)
	if err != nil {
		panic(err)
	}
	return ret
}

// time.Time formats.
const (
	// TimestampTZOutputFormat is used to output all TimestampTZs.
	TimestampTZOutputFormat = "2006-01-02 15:04:05.999999-07:00"
	// TimestampOutputFormat is used to output all Timestamps.
	TimestampOutputFormat = "2006-01-02 15:04:05.999999"
)

// ParseDTimestamp parses and returns the *DTimestamp Datum value represented by
// the provided string in UTC, or an error if parsing is unsuccessful.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseDTimestamp(
	ctx ParseTimeContext, s string, precision time.Duration,
) (_ *DTimestamp, dependsOnContext bool, _ error) {
	now := relativeParseTime(ctx)
	t, dependsOnContext, err := pgdate.ParseTimestampWithoutTimezone(now, pgdate.ParseModeYMD, s)
	if err != nil {
		t, dependsOnContext, err = pgdate.ParseTimestampWithoutTimezone(now, pgdate.ParseModeDMY, s)
		if err != nil {
			t, dependsOnContext, err = pgdate.ParseTimestampWithoutTimezone(now, pgdate.ParseModeMDY, s)
			if err != nil {
				return nil, false, err
			}
		}
	}
	d, err := MakeDTimestamp(t, precision)
	return d, dependsOnContext, err
}

// Round returns a new DTimestamp to the specified precision.
func (d *DTimestamp) Round(precision time.Duration) (*DTimestamp, error) {
	return MakeDTimestamp(d.Time, precision)
}

// ResolvedType implements the TypedExpr interface.
func (*DTimestamp) ResolvedType() *types.T {
	return types.Timestamp
}

// AmbiguousFormat implements the Datum interface.
func (*DTimestamp) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DTimestamp) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.UTC().Format(TimestampOutputFormat))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DTimestamp) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DTimestampTZ is the timestamp Datum that is rendered with session offset.
type DTimestampTZ struct {
	time.Time
}

// MakeDTimestampTZ creates a DTimestampTZ with specified precision.
func MakeDTimestampTZ(t time.Time, precision time.Duration) (*DTimestampTZ, error) {
	ret := t.Round(precision)
	if ret.After(MaxSupportedTime) || ret.Before(MinSupportedTime) {
		return nil, errors.Newf("timestamp %q exceeds supported timestamp bounds", ret.Format(time.RFC3339))
	}
	return &DTimestampTZ{Time: ret}, nil
}

// MustMakeDTimestampTZ wraps MakeDTimestampTZ but panics if there is an error.
// This is intended for testing applications only.
func MustMakeDTimestampTZ(t time.Time, precision time.Duration) *DTimestampTZ {
	ret, err := MakeDTimestampTZ(t, precision)
	if err != nil {
		panic(err)
	}
	return ret
}

// ParseDTimestampTZ parses and returns the *DTimestampTZ Datum value represented by
// the provided string in the provided location, or an error if parsing is unsuccessful.
//
// The dependsOnContext return value indicates if we had to consult the
// ParseTimeContext (either for the time or the local timezone).
func ParseDTimestampTZ(ctx ParseTimeContext, s string, precision time.Duration, loc *time.Location) (_ *DTimestampTZ, dependsOnContext bool, _ error) {
	now := relativeParseTime(ctx).In(loc)
	t, dependsOnContext, err := pgdate.ParseTimestamp(now, pgdate.ParseModeYMD, s)
	if err != nil {
		t, dependsOnContext, err = pgdate.ParseTimestamp(now, pgdate.ParseModeDMY, s)
		if err != nil {
			t, dependsOnContext, err = pgdate.ParseTimestamp(now, pgdate.ParseModeMDY, s)
			if err != nil {
				return nil, false, err
			}
		}
	}
	// Always normalize time to the current location.
	d, err := MakeDTimestampTZ(t, precision)
	return d, dependsOnContext, err
}

// Round returns a new DTimestampTZ to the specified precision.
func (d *DTimestampTZ) Round(precision time.Duration) (*DTimestampTZ, error) {
	return MakeDTimestampTZ(d.Time, precision)
}

// ResolvedType implements the TypedExpr interface.
func (*DTimestampTZ) ResolvedType() *types.T {
	return types.TimestampTZ
}

// AmbiguousFormat implements the Datum interface.
func (*DTimestampTZ) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DTimestampTZ) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.Time.Format(TimestampTZOutputFormat))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DTimestampTZ) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DInterval is the interval Datum.
type DInterval struct {
	duration.Duration
}

// ParseDInterval parses and returns the *DInterval Datum value represented by the provided
// string, or an error if parsing is unsuccessful.
func ParseDInterval(s string) (*DInterval, error) {
	return ParseDIntervalWithTypeMetadata(s, types.DefaultIntervalTypeMetadata)
}

// truncateDInterval truncates the input DInterval downward to the nearest
// interval quantity specified by the DurationField input.
// If precision is set for seconds, this will instead round at the second layer.
func truncateDInterval(d *DInterval, itm types.IntervalTypeMetadata) {
	switch itm.DurationField.DurationType {
	case types.IntervalDurationType_YEAR:
		d.Duration.Months = d.Duration.Months - d.Duration.Months%12
		d.Duration.Days = 0
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_MONTH:
		d.Duration.Days = 0
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_DAY:
		d.Duration.SetNanos(0)
	case types.IntervalDurationType_HOUR:
		d.Duration.SetNanos(d.Duration.Nanos() - d.Duration.Nanos()%time.Hour.Nanoseconds())
	case types.IntervalDurationType_MINUTE:
		d.Duration.SetNanos(d.Duration.Nanos() - d.Duration.Nanos()%time.Minute.Nanoseconds())
	case types.IntervalDurationType_SECOND, types.IntervalDurationType_UNSET:
		if itm.PrecisionIsSet || itm.Precision > 0 {
			prec := TimeFamilyPrecisionToRoundDuration(itm.Precision)
			d.Duration.SetNanos(time.Duration(d.Duration.Nanos()).Round(prec).Nanoseconds())
		}
	}
}

// ParseDIntervalWithTypeMetadata is like ParseDInterval, but it also takes a
// types.IntervalTypeMetadata that both specifies the units for unitless, numeric intervals
// and also specifies the precision of the interval.
func ParseDIntervalWithTypeMetadata(s string, itm types.IntervalTypeMetadata) (*DInterval, error) {
	d, err := parseDInterval(s, itm)
	if err != nil {
		return nil, err
	}
	truncateDInterval(d, itm)
	return d, nil
}

func parseDInterval(s string, itm types.IntervalTypeMetadata) (*DInterval, error) {
	// At this time the only supported interval formats are:
	// - SQL standard.
	// - Postgres compatible.
	// - iso8601 format (with designators only), see interval.go for
	//   sources of documentation.
	// - Golang time.parseDuration compatible.

	// If it's a blank string, exit early.
	if len(s) == 0 {
		return nil, makeParseError(s, types.Interval, nil)
	}
	if s[0] == 'P' {
		// If it has a leading P we're most likely working with an iso8601
		// interval.
		dur, err := iso8601ToDuration(s)
		if err != nil {
			return nil, makeParseError(s, types.Interval, err)
		}
		return &DInterval{Duration: dur}, nil
	}
	if strings.IndexFunc(s, unicode.IsLetter) == -1 {
		// If it has no letter, then we're most likely working with a SQL standard
		// interval, as both postgres and golang have letter(s) and iso8601 has been tested.
		dur, err := sqlStdToDuration(s, itm)
		if err != nil {
			return nil, makeParseError(s, types.Interval, err)
		}
		return &DInterval{Duration: dur}, nil
	}

	// We're either a postgres string or a Go duration.
	// Our postgres syntax parser also supports golang, so just use that for both.
	dur, err := parseDuration(s, itm)
	if err != nil {
		return nil, makeParseError(s, types.Interval, err)
	}
	return &DInterval{Duration: dur}, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DInterval) ResolvedType() *types.T {
	return types.Interval
}

// ValueAsString returns the interval as a string (e.g. "1h2m").
func (d *DInterval) ValueAsString() string {
	return d.Duration.String()
}

// AmbiguousFormat implements the Datum interface.
func (*DInterval) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DInterval) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	d.Duration.Format(&ctx.Buffer)
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DInterval) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DGeography is the Geometry Datum.
type DGeography struct {
	geo.Geography
}

// NewDGeography returns a new Geography Datum.
func NewDGeography(g geo.Geography) *DGeography {
	return &DGeography{Geography: g}
}

// ParseDGeography attempts to pass `str` as a Geography type.
func ParseDGeography(str string) (*DGeography, error) {
	g, err := geo.ParseGeography(str)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse geography")
	}
	return &DGeography{Geography: g}, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DGeography) ResolvedType() *types.T {
	return types.Geography
}

// AmbiguousFormat implements the Datum interface.
func (*DGeography) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DGeography) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.Geography.EWKBHex())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DGeography) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DGeometry is the Geometry Datum.
type DGeometry struct {
	geo.Geometry
}

// NewDGeometry returns a new Geometry Datum.
func NewDGeometry(g geo.Geometry) *DGeometry {
	return &DGeometry{Geometry: g}
}

// ParseDGeometry attempts to pass `str` as a Geometry type.
func ParseDGeometry(str string) (*DGeometry, error) {
	g, err := geo.ParseGeometry(str)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse geometry")
	}
	return &DGeometry{Geometry: g}, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DGeometry) ResolvedType() *types.T {
	return types.Geometry
}

// AmbiguousFormat implements the Datum interface.
func (*DGeometry) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DGeometry) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.Geometry.EWKBHex())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DGeometry) Size() uintptr {
	return unsafe.Sizeof(*d)
}

// DBox2D is the Datum representation of the Box2D type.
type DBox2D struct {
	geo.CartesianBoundingBox
}

// NewDBox2D returns a new Box2D Datum.
func NewDBox2D(b geo.CartesianBoundingBox) *DBox2D {
	return &DBox2D{CartesianBoundingBox: b}
}

// ParseDBox2D attempts to pass `str` as a Box2D type.
func ParseDBox2D(str string) (*DBox2D, error) {
	b, err := geo.ParseCartesianBoundingBox(str)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse geometry")
	}
	return &DBox2D{CartesianBoundingBox: b}, nil
}

// ResolvedType implements the TypedExpr interface.
func (*DBox2D) ResolvedType() *types.T {
	return types.Box2D
}

// AmbiguousFormat implements the Datum interface.
func (*DBox2D) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DBox2D) Format(ctx *FmtCtx) {
	f := ctx.flags
	bareStrings := f.HasFlags(FmtFlags(lex.EncBareStrings))
	if !bareStrings {
		ctx.WriteByte('\'')
	}
	ctx.WriteString(d.CartesianBoundingBox.Repr())
	if !bareStrings {
		ctx.WriteByte('\'')
	}
}

// Size implements the Datum interface.
func (d *DBox2D) Size() uintptr {
	return unsafe.Sizeof(*d) + unsafe.Sizeof(d.CartesianBoundingBox)
}

// DJSON is the JSON Datum.
type DJSON struct{ json.JSON }

// NewDJSON is a helper routine to create a DJSON initialized from its argument.
func NewDJSON(j json.JSON) *DJSON {
	return &DJSON{j}
}

// ParseDJSON takes a string of JSON and returns a DJSON value.
func ParseDJSON(s string) (Datum, error) {
	j, err := json.ParseJSON(s)
	if err != nil {
		return nil, pgerror.Wrapf(err, pgcode.Syntax, "could not parse JSON")
	}
	return NewDJSON(j), nil
}

// ResolvedType implements the TypedExpr interface.
func (*DJSON) ResolvedType() *types.T {
	return types.Jsonb
}

// AmbiguousFormat implements the Datum interface.
func (*DJSON) AmbiguousFormat() bool { return true }

// Format implements the NodeFormatter interface.
func (d *DJSON) Format(ctx *FmtCtx) {
	// TODO(justin): ideally the JSON string encoder should know it needs to
	// escape things to be inside SQL strings in order to avoid this allocation.
	s := d.JSON.String()
	if ctx.flags.HasFlags(fmtRawStrings) {
		ctx.WriteString(s)
	} else {
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, s, ctx.flags.EncodeFlags())
	}
}

// Size implements the Datum interface.
// TODO(justin): is this a frequently-called method? Should we be caching the computed size?
func (d *DJSON) Size() uintptr {
	return unsafe.Sizeof(*d) + d.JSON.Size()
}

// DTuple is the tuple Datum.
type DTuple struct {
	D Datums

	// sorted indicates that the values in D are pre-sorted.
	// This is used to accelerate IN comparisons.
	sorted bool

	// typ is the tuple's type.
	//
	// The Types sub-field can be initially uninitialized, and is then
	// populated upon first invocation of ResolvedTypes(). If
	// initialized it must have the same arity as D.
	//
	// The Labels sub-field can be left nil. If populated, it must have
	// the same arity as D.
	typ *types.T
}

// maybePopulateType populates the tuple's type if it hasn't yet been
// populated.
func (d *DTuple) maybePopulateType() {
	if d.typ == nil {
		contents := make([]*types.T, len(d.D))
		for i, v := range d.D {
			contents[i] = v.ResolvedType()
		}
		d.typ = types.MakeTuple(contents)
	}
}

// ResolvedType implements the TypedExpr interface.
func (d *DTuple) ResolvedType() *types.T {
	d.maybePopulateType()
	return d.typ
}

// AmbiguousFormat implements the Datum interface.
func (*DTuple) AmbiguousFormat() bool { return false }

// Format implements the NodeFormatter interface.
func (d *DTuple) Format(ctx *FmtCtx) {
	if ctx.HasFlags(fmtPgwireFormat) {
		d.pgwireFormat(ctx)
		return
	}

	typ := d.ResolvedType()
	showLabels := len(typ.TupleLabels()) > 0
	if showLabels {
		ctx.WriteByte('(')
	}
	ctx.WriteByte('(')
	comma := ""
	parsable := ctx.HasFlags(FmtParsable)
	for i, v := range d.D {
		ctx.WriteString(comma)
		ctx.FormatNode(v)
		if parsable && (v == DNull) && len(typ.TupleContents()) > i {
			// If Tuple has types.Unknown for this slot, then we can't determine
			// the column type to write this annotation. Somebody else will provide
			// an error message in this case, if necessary, so just skip the
			// annotation and continue.
			if typ.TupleContents()[i].Family() != types.UnknownFamily {
				ctx.WriteString("::")
				ctx.WriteString(typ.TupleContents()[i].SQLString())
			}
		}
		comma = ", "
	}
	if len(d.D) == 1 {
		// Ensure the pretty-printed 1-value tuple is not ambiguous with
		// the equivalent value enclosed in grouping parentheses.
		ctx.WriteByte(',')
	}
	ctx.WriteByte(')')
	if showLabels {
		ctx.WriteString(" AS ")
		comma := ""
		for i := range typ.TupleLabels() {
			ctx.WriteString(comma)
			ctx.FormatNode((*Name)(&typ.TupleLabels()[i]))
			comma = ", "
		}
		ctx.WriteByte(')')
	}
}

// Sorted returns true if the tuple is known to be sorted (and contains no
// NULLs).
func (d *DTuple) Sorted() bool {
	return d.sorted
}

// SetSorted sets the sorted flag on the DTuple. This should be used when a
// DTuple is known to be sorted based on the datums added to it.
func (d *DTuple) SetSorted() *DTuple {
	if d.ContainsNull() {
		// A DTuple that contains a NULL (see ContainsNull) cannot be marked as sorted.
		return d
	}
	d.sorted = true
	return d
}

// AssertSorted asserts that the DTuple is sorted.
func (d *DTuple) AssertSorted() {
	if !d.sorted {
		panic(errors.AssertionFailedf("expected sorted tuple, found %#v", d))
	}
}

// Size implements the Datum interface.
func (d *DTuple) Size() uintptr {
	sz := unsafe.Sizeof(*d)
	for _, e := range d.D {
		dsz := e.Size()
		sz += dsz
	}
	return sz
}

// ContainsNull returns true if the tuple contains NULL, possibly nested inside
// other tuples. For example, all the following tuples contain NULL:
//
//	(1, 2, NULL)
//	((1, 1), (2, NULL))
//	(((1, 1), (2, 2)), ((3, 3), (4, NULL)))
func (d *DTuple) ContainsNull() bool {
	for _, r := range d.D {
		if r == DNull {
			return true
		}
		if t, ok := r.(*DTuple); ok {
			if t.ContainsNull() {
				return true
			}
		}
	}
	return false
}

type NullLiteral struct{}

// ResolvedType implements the TypedExpr interface.
func (NullLiteral) ResolvedType() *types.T {
	return types.Unknown
}

// AmbiguousFormat implements the Datum interface.
func (NullLiteral) AmbiguousFormat() bool { return false }

// Format implements the NodeFormatter interface.
func (NullLiteral) Format(ctx *FmtCtx) {
	if ctx.HasFlags(fmtPgwireFormat) {
		// NULL sub-expressions in pgwire text values are represented with
		// the empty string.
		return
	}
	ctx.WriteString("NULL")
}

// Size implements the Datum interface.
func (d NullLiteral) Size() uintptr {
	return unsafe.Sizeof(d)
}

// DArray is the array Datum. Any Datum inserted into a DArray are treated as
// text during serialization.
type DArray struct {
	ParamTyp *types.T
	Array    Datums
	// HasNulls is set to true if any of the datums within the array are null.
	// This is used in the binary array serialization format.
	HasNulls bool
	// HasNonNulls is set to true if any of the datums within the are non-null.
	// This is used in expression serialization (FmtParsable).
	HasNonNulls bool

	// customOid, if non-0, is the oid of this array datum.
	customOid oid.Oid
}

// NewDArray returns a DArray containing elements of the specified type.
func NewDArray(paramTyp *types.T) *DArray {
	return &DArray{ParamTyp: paramTyp}
}

// AsDArray attempts to retrieve a *DArray from an Expr, returning a *DArray and
// a flag signifying whether the assertion was successful. The function should
// be used instead of direct type assertions wherever a *DArray wrapped by a
// *DOidWrapper is possible.
func AsDArray(e Expr) (*DArray, bool) {
	switch t := e.(type) {
	case *DArray:
		return t, true
	case *DOidWrapper:
		return AsDArray(t.Wrapped)
	}
	return nil, false
}

// MustBeDArray attempts to retrieve a *DArray from an Expr, panicking if the
// assertion fails.
func MustBeDArray(e Expr) *DArray {
	i, ok := AsDArray(e)
	if !ok {
		panic(errors.AssertionFailedf("expected *DArray, found %T", e))
	}
	return i
}

// ResolvedType implements the TypedExpr interface.
func (d *DArray) ResolvedType() *types.T {
	switch d.customOid {
	case oid.T_int2vector:
		return types.Int2Vector
	case oid.T_oidvector:
		return types.OidVector
	}
	return types.MakeArray(d.ParamTyp)
}

// IsComposite implements the CompositeDatum interface.
func (d *DArray) IsComposite() bool {
	for _, elem := range d.Array {
		if cdatum, ok := elem.(CompositeDatum); ok && cdatum.IsComposite() {
			return true
		}
	}
	return false
}

// FirstIndex returns the first index of the array. 1 for normal SQL arrays,
// which are 1-indexed, and 0 for the special Postgers vector types which are
// 0-indexed.
func (d *DArray) FirstIndex() int {
	switch d.customOid {
	case oid.T_int2vector, oid.T_oidvector:
		return 0
	}
	return 1
}

// AmbiguousFormat implements the Datum interface.
func (d *DArray) AmbiguousFormat() bool {
	// The type of the array is ambiguous if it is empty or all-null; when
	// serializing we need to annotate it with the type.
	if d.ParamTyp.Family() == types.UnknownFamily {
		// If the array's type is unknown, marking it as ambiguous would cause the
		// expression formatter to try to annotate it with UNKNOWN[], which is not
		// a valid type. So an array of unknown type is (paradoxically) unambiguous.
		return false
	}
	return !d.HasNonNulls
}

// Format implements the NodeFormatter interface.
func (d *DArray) Format(ctx *FmtCtx) {
	if ctx.HasFlags(fmtPgwireFormat) {
		d.pgwireFormat(ctx)
		return
	}

	// If we want to export arrays, we need to ensure that
	// the datums within the arrays are formatted with enclosing quotes etc.
	if ctx.HasFlags(FmtExport) {
		oldFlags := ctx.flags
		ctx.flags = oldFlags & ^FmtExport | FmtParsable
		defer func() { ctx.flags = oldFlags }()
	}

	ctx.WriteString("ARRAY[")
	comma := ""
	for _, v := range d.Array {
		ctx.WriteString(comma)
		ctx.FormatNode(v)
		comma = ","
	}
	ctx.WriteByte(']')
}

const maxArrayLength = math.MaxInt32

var errArrayTooLongError = errors.New("ARRAYs can be at most 2^31-1 elements long")

// Validate checks that the given array is valid,
// for example, that it's not too big.
func (d *DArray) Validate() error {
	if d.Len() > maxArrayLength {
		return errors.WithStack(errArrayTooLongError)
	}
	return nil
}

// Len returns the length of the Datum array.
func (d *DArray) Len() int {
	return len(d.Array)
}

// Size implements the Datum interface.
func (d *DArray) Size() uintptr {
	sz := unsafe.Sizeof(*d)
	for _, e := range d.Array {
		dsz := e.Size()
		sz += dsz
	}
	return sz
}

var errNonHomogeneousArray = pgerror.New(pgcode.ArraySubscript, "multidimensional arrays must have array expressions with matching dimensions")

// Append appends a Datum to the array, whose parameterized type must be
// consistent with the type of the Datum.
func (d *DArray) Append(v Datum) error {
	if v != DNull && !d.ParamTyp.Equivalent(v.ResolvedType()) {
		return errors.AssertionFailedf("cannot append %s to array containing %s", d.ParamTyp,
			v.ResolvedType())
	}
	if d.Len() >= maxArrayLength {
		return errors.WithStack(errArrayTooLongError)
	}
	if d.ParamTyp.Family() == types.ArrayFamily {
		if v == DNull {
			return errNonHomogeneousArray
		}
		if d.Len() > 0 {
			prevItem := d.Array[d.Len()-1]
			if prevItem == DNull {
				return errNonHomogeneousArray
			}
			expectedLen := MustBeDArray(prevItem).Len()
			if MustBeDArray(v).Len() != expectedLen {
				return errNonHomogeneousArray
			}
		}
	}
	if v == DNull {
		d.HasNulls = true
	} else {
		d.HasNonNulls = true
	}
	d.Array = append(d.Array, v)
	return d.Validate()
}

// DEnum represents an ENUM value.
type DEnum struct {
	// EnumType is the hydrated type of this enum.
	EnumTyp *types.T
	// PhysicalRep is a slice containing the encodable and ordered physical
	// representation of this datum. It is used for comparisons and encoding.
	PhysicalRep []byte
	// LogicalRep is a string containing the user visible value of the enum.
	LogicalRep string
}

// Size implements the Datum interface.
func (d *DEnum) Size() uintptr {
	// When creating DEnums, we store pointers back into the type enum
	// metadata, so enums themselves don't pay for the memory of their
	// physical and logical representations.
	return unsafe.Sizeof(d.EnumTyp) +
		unsafe.Sizeof(d.PhysicalRep) +
		unsafe.Sizeof(d.LogicalRep)
}

// GetEnumComponentsFromPhysicalRep returns the physical and logical components
// for an enum of the requested type. It returns an error if it cannot find a
// matching physical representation.
func GetEnumComponentsFromPhysicalRep(typ *types.T, rep []byte) ([]byte, string, error) {
	idx, err := typ.EnumGetIdxOfPhysical(rep)
	if err != nil {
		return nil, "", err
	}
	meta := typ.TypeMeta.EnumData
	// Take a pointer into the enum metadata rather than holding on
	// to a pointer to the input bytes.
	return meta.PhysicalRepresentations[idx], meta.LogicalRepresentations[idx], nil
}

// MakeDEnumFromPhysicalRepresentation creates a DEnum of the input type
// and the input physical representation.
func MakeDEnumFromPhysicalRepresentation(typ *types.T, rep []byte) (*DEnum, error) {
	// Return a nice error if the input requested type is types.AnyEnum.
	if typ.Oid() == oid.T_anyenum {
		return nil, errors.New("cannot create enum of unspecified type")
	}
	phys, log, err := GetEnumComponentsFromPhysicalRep(typ, rep)
	if err != nil {
		return nil, err
	}
	return &DEnum{
		EnumTyp:     typ,
		PhysicalRep: phys,
		LogicalRep:  log,
	}, nil
}

// MakeDEnumFromLogicalRepresentation creates a DEnum of the input type
// and input logical representation. It returns an error if the input
// logical representation is invalid.
func MakeDEnumFromLogicalRepresentation(typ *types.T, rep string) (*DEnum, error) {
	// Return a nice error if the input requested type is types.AnyEnum.
	if typ.Oid() == oid.T_anyenum {
		return nil, errors.New("cannot create enum of unspecified type")
	}
	// Take a pointer into the enum metadata rather than holding on
	// to a pointer to the input string.
	idx, err := typ.EnumGetIdxOfLogical(rep)
	if err != nil {
		return nil, err
	}
	// If this enum member is read only, we cannot construct it from the logical
	// representation. This is to ensure that it will not be written until all
	// nodes in the cluster are able to decode the physical representation.
	if typ.TypeMeta.EnumData.IsMemberReadOnly[idx] {
		return nil, errors.Newf("enum label %q is not yet public", rep)
	}
	return &DEnum{
		EnumTyp:     typ,
		PhysicalRep: typ.TypeMeta.EnumData.PhysicalRepresentations[idx],
		LogicalRep:  typ.TypeMeta.EnumData.LogicalRepresentations[idx],
	}, nil
}

// Format implements the NodeFormatter interface.
func (d *DEnum) Format(ctx *FmtCtx) {
	if ctx.HasFlags(fmtStaticallyFormatUserDefinedTypes) {
		s := DBytes(d.PhysicalRep)
		// We use the fmtFormatByteLiterals flag here so that the bytes
		// get formatted as byte literals. Consider an enum of type t with physical
		// representation \x80. If we don't format this as a bytes literal then
		// it gets emitted as '\x80':::t. '\x80' is scanned as a string, and we try
		// to find a logical representation matching '\x80', which won't exist.
		// Instead, we want to emit b'\x80'::: so that '\x80' is scanned as bytes,
		// triggering the logic to cast the bytes \x80 to t.
		ctx.WithFlags(ctx.flags|fmtFormatByteLiterals, func() {
			s.Format(ctx)
		})
	} else {
		s := DString(d.LogicalRep)
		s.Format(ctx)
	}
}

func (d *DEnum) String() string {
	return AsString(d)
}

// ResolvedType implements the Datum interface.
func (d *DEnum) ResolvedType() *types.T {
	return d.EnumTyp
}

// AmbiguousFormat implements the Datum interface.
func (d *DEnum) AmbiguousFormat() bool {
	return true
}

// MaxWriteable returns the largest member of the enum that is writeable.
func (d *DEnum) MaxWriteable() (Datum, bool) {
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		return nil, false
	}
	for i := len(enumData.PhysicalRepresentations) - 1; i >= 0; i-- {
		if !enumData.IsMemberReadOnly[i] {
			return &DEnum{
				EnumTyp:     d.EnumTyp,
				PhysicalRep: enumData.PhysicalRepresentations[i],
				LogicalRep:  enumData.LogicalRepresentations[i],
			}, true
		}
	}
	return nil, false
}

// MinWriteable returns the smallest member of the enum that is writeable.
func (d *DEnum) MinWriteable() (Datum, bool) {
	enumData := d.EnumTyp.TypeMeta.EnumData
	if len(enumData.PhysicalRepresentations) == 0 {
		return nil, false
	}
	for i := 0; i < len(enumData.PhysicalRepresentations); i++ {
		if !enumData.IsMemberReadOnly[i] {
			return &DEnum{
				EnumTyp:     d.EnumTyp,
				PhysicalRep: enumData.PhysicalRepresentations[i],
				LogicalRep:  enumData.LogicalRepresentations[i],
			}, true
		}
	}
	return nil, false
}

// DOid is the Postgres OID datum. It can represent either an OID type or any
// of the reg* types, such as regproc or regclass.
type DOid struct {
	// A DOid embeds a DInt, the underlying integer OID for this OID datum.
	DInt
	// semanticType indicates the particular variety of OID this datum is, whether raw
	// oid or a reg* type.
	semanticType *types.T
	// name is set to the resolved name of this OID, if available.
	name string
}

// MakeDOid is a helper routine to create a DOid initialized from a DInt.
func MakeDOid(d DInt) DOid {
	return DOid{DInt: d, semanticType: types.Oid, name: ""}
}

// NewDOid is a helper routine to create a *DOid initialized from a DInt.
func NewDOid(d DInt) *DOid {
	oid := MakeDOid(d)
	return &oid
}

// AsRegProc changes the input DOid into a regproc with the given name and
// returns it.
func (d *DOid) AsRegProc(name string) *DOid {
	d.name = name
	d.semanticType = types.RegProc
	return d
}

// AmbiguousFormat implements the Datum interface.
func (*DOid) AmbiguousFormat() bool { return true }

// Format implements the Datum interface.
func (d *DOid) Format(ctx *FmtCtx) {
	if d.semanticType.Oid() == oid.T_oid || d.name == "" {
		// If we call FormatNode directly when the disambiguateDatumTypes flag
		// is set, then we get something like 123:::INT:::OID. This is the
		// important flag set by FmtParsable which is supposed to be
		// roundtrippable. Since in this branch, a DOid is a thin wrapper around
		// a DInt, I _think_ it's correct to just delegate to the DInt's Format.
		d.DInt.Format(ctx)
	} else if ctx.HasFlags(fmtDisambiguateDatumTypes) {
		ctx.WriteString("crdb_internal.create_")
		ctx.WriteString(d.semanticType.SQLStandardName())
		ctx.WriteByte('(')
		d.DInt.Format(ctx)
		ctx.WriteByte(',')
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, d.name, lex.EncNoFlags)
		ctx.WriteByte(')')
	} else {
		// This is used to print the name of pseudo-procedures in e.g.
		// pg_catalog.pg_type.typinput
		lex.EncodeSQLStringWithFlags(&ctx.Buffer, d.name, lex.EncBareStrings)
	}
}

// ResolvedType implements the Datum interface.
func (d *DOid) ResolvedType() *types.T {
	return d.semanticType
}

// Size implements the Datum interface.
func (d *DOid) Size() uintptr { return unsafe.Sizeof(*d) }

// DOidWrapper is a Datum implementation which is a wrapper around a Datum, allowing
// custom Oid values to be attached to the Datum and its types.T.
// The reason the Datum type was introduced was to permit the introduction of Datum
// types with new Object IDs while maintaining identical behavior to current Datum
// types. Specifically, it obviates the need to define a new tree.Datum type for
// each possible Oid value.
//
// Instead, DOidWrapper allows a standard Datum to be wrapped with a new Oid.
// This approach provides two major advantages:
//   - performance of the existing Datum types are not affected because they
//     do not need to have custom oid.Oids added to their structure.
//   - the introduction of new Datum aliases is straightforward and does not require
//     additions to typing rules or type-dependent evaluation behavior.
//
// Types that currently benefit from DOidWrapper are:
// - DName => DOidWrapper(*DString, oid.T_name)
type DOidWrapper struct {
	Wrapped Datum
	Oid     oid.Oid
}

// wrapWithOid wraps a Datum with a custom Oid.
func wrapWithOid(d Datum, oid oid.Oid) Datum {
	switch v := d.(type) {
	case nil:
		return nil
	case *DInt:
	case *DString:
	case *DArray:
	case NullLiteral, *DOidWrapper:
		panic(errors.AssertionFailedf("cannot wrap %T with an Oid", v))
	default:
		// Currently only *DInt, *DString, *DArray are hooked up to work with
		// *DOidWrapper. To support another base Datum type, replace all type
		// assertions to that type with calls to functions like AsDInt and
		// MustBeDInt.
		panic(errors.AssertionFailedf("unsupported Datum type passed to wrapWithOid: %T", d))
	}
	return &DOidWrapper{
		Wrapped: d,
		Oid:     oid,
	}
}

// UnwrapDatum returns the base Datum type for a provided datum, stripping
// an *DOidWrapper if present. This is useful for cases like type switches,
// where type aliases should be ignored.
func UnwrapDatum(d Datum) Datum {
	if w, ok := d.(*DOidWrapper); ok {
		return w.Wrapped
	}
	return d
}

// ResolvedType implements the TypedExpr interface.
func (d *DOidWrapper) ResolvedType() *types.T {
	return types.OidToType[d.Oid]
}

// AmbiguousFormat implements the Datum interface.
func (d *DOidWrapper) AmbiguousFormat() bool {
	return d.Wrapped.AmbiguousFormat()
}

// Format implements the NodeFormatter interface.
func (d *DOidWrapper) Format(ctx *FmtCtx) {
	// Custom formatting based on d.OID could go here.
	ctx.FormatNode(d.Wrapped)
}

// Size implements the Datum interface.
func (d *DOidWrapper) Size() uintptr {
	return unsafe.Sizeof(*d) + d.Wrapped.Size()
}

// AmbiguousFormat implements the Datum interface.
func (d *Placeholder) AmbiguousFormat() bool {
	return true
}

// Size implements the Datum interface.
func (d *Placeholder) Size() uintptr {
	panic(errors.AssertionFailedf("shouldn't get called"))
}

// NewDNameFromDString is a helper routine to create a *DName (implemented as
// a *DOidWrapper) initialized from an existing *DString.
func NewDNameFromDString(d *DString) Datum {
	return wrapWithOid(d, oid.T_name)
}
