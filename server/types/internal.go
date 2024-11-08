package types

import "github.com/lib/pq/oid"

// Internal is an internal type, which means `external binary` type.
var Internal = DoltgresType{
	OID:           uint32(oid.T_internal),
	Name:          "internal",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Pseudo,
	TypCategory:   TypeCategory_PseudoTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         0,
	InputFunc:     "internal_in",
	OutputFunc:    "internal_out",
	ReceiveFunc:   "-",
	SendFunc:      "-",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  0,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
}

func NewInternalTypeWithBaseType(t uint32) DoltgresType {
	it := Internal
	it.BaseTypeForInternal = t
	return it
}
