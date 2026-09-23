package parser

import (
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// makeCreateDatabase preserves option presence independently of the value, so
// repeated empty strings are rejected just like any other duplicate option.
func makeCreateDatabase(name tree.Name, ifNotExists bool, options []tree.KVOption) (*tree.CreateDatabase, error) {
	db := &tree.CreateDatabase{Name: name, IfNotExists: ifNotExists}
	seen := make(map[tree.Name]bool, len(options))
	for _, option := range options {
		if seen[option.Key] {
			return nil, pgerror.New(pgcode.Syntax, "conflicting or redundant options")
		}
		seen[option.Key] = true
		switch option.Key {
		case "owner":
			db.Owner = option.Value.(*tree.StrVal).RawString()
		case "template":
			db.Template = option.Value.(*tree.StrVal).RawString()
		case "encoding":
			db.Encoding = option.Value.(*tree.StrVal).RawString()
		case "strategy":
			db.Strategy = option.Value.(*tree.StrVal).RawString()
		case "locale":
			db.Locale = option.Value.(*tree.StrVal).RawString()
		case "lc_collate":
			db.Collate = option.Value.(*tree.StrVal).RawString()
		case "lc_ctype":
			db.CType = option.Value.(*tree.StrVal).RawString()
		case "icu_locale":
			db.IcuLocale = option.Value.(*tree.StrVal).RawString()
		case "icu_rules":
			db.IcuRules = option.Value.(*tree.StrVal).RawString()
		case "locale_provider":
			db.LocaleProvider = option.Value.(*tree.StrVal).RawString()
		case "collation_version":
			db.CollationVersion = option.Value.(*tree.StrVal).RawString()
		case "tablespace":
			db.Tablespace = option.Value.(*tree.StrVal).RawString()
		case "allow_connections":
			db.AllowConnections = option.Value
		case "connection limit":
			db.ConnectionLimit = option.Value
		case "is_template":
			db.IsTemplate = option.Value
		case "oid":
			db.Oid = option.Value
		}
	}
	return db, nil
}
