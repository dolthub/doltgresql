package parser

import (
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"reflect"
	"strings"
	"testing"
)

func TestCreateDatabaseOptionPermutations(t *testing.T) {
	options := []string{"TEMPLATE template0", "ENCODING 'UTF8'", "LC_COLLATE 'C'", "OWNER postgres"}
	var visit func([]string, []string)
	visit = func(prefix, rest []string) {
		if len(rest) != 0 {
			for i, option := range rest {
				next := append([]string{}, rest[:i]...)
				next = append(next, rest[i+1:]...)
				visit(append(append([]string{}, prefix...), option), next)
			}
			return
		}
		for _, head := range []string{"CREATE DATABASE db ", "CREATE DATABASE IF NOT EXISTS db WITH "} {
			sql := head + strings.Join(prefix, " ")
			t.Run(sql, func(t *testing.T) {
				stmt, err := ParseOne(sql)
				if err != nil {
					t.Fatal(err)
				}
				got := stmt.AST.(*tree.CreateDatabase)
				want := &tree.CreateDatabase{Name: "db", IfNotExists: strings.Contains(head, "IF NOT EXISTS"), Owner: "postgres", Template: "template0", Encoding: "UTF8", Collate: "C"}
				if !reflect.DeepEqual(got, want) {
					t.Fatalf("got %#v, want %#v", got, want)
				}
			})
		}
	}
	visit(nil, options)
}

func TestCreateDatabaseAllOptions(t *testing.T) {
	// Reverse the formerly required order and exercise optional equals signs.
	stmt, err := ParseOne(`CREATE DATABASE db WITH OID = 17000 IS_TEMPLATE false CONNECTION LIMIT = -1 ALLOW_CONNECTIONS true TABLESPACE pg_default COLLATION_VERSION '1' LOCALE_PROVIDER libc ICU_RULES '' ICU_LOCALE 'en' LC_CTYPE 'C' LC_COLLATE 'C' LOCALE 'C' STRATEGY wal_log ENCODING 'UTF8' TEMPLATE template0 OWNER postgres`)
	if err != nil {
		t.Fatal(err)
	}
	got := stmt.AST.(*tree.CreateDatabase)
	if got.Oid.String() != "17000" || got.IsTemplate.String() != "false" || got.ConnectionLimit.String() != "-1" || got.AllowConnections.String() != "true" {
		t.Fatalf("lost expression option: %#v", got)
	}
	if got.Owner != "postgres" || got.Template != "template0" || got.Encoding != "UTF8" || got.Strategy != "wal_log" || got.Locale != "C" || got.Collate != "C" || got.CType != "C" || got.IcuLocale != "en" || got.IcuRules != "" || got.LocaleProvider != "libc" || got.CollationVersion != "1" || got.Tablespace != "pg_default" {
		t.Fatalf("lost string option: %#v", got)
	}
	for _, sql := range []string{"CREATE DATABASE db", "CREATE DATABASE IF NOT EXISTS db", "CREATE DATABASE db WITH ENCODING = 'UTF8' TEMPLATE = 'template0'"} {
		if _, err := ParseOne(sql); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
}

func TestCreateDatabaseDuplicateOptions(t *testing.T) {
	options := []string{"OWNER postgres", "TEMPLATE template0", "ENCODING 'UTF8'", "STRATEGY wal_log", "LOCALE 'C'", "LC_COLLATE 'C'", "LC_CTYPE 'C'", "ICU_LOCALE 'en'", "ICU_RULES ''", "LOCALE_PROVIDER libc", "COLLATION_VERSION ''", "TABLESPACE pg_default", "ALLOW_CONNECTIONS true", "CONNECTION LIMIT -1", "IS_TEMPLATE false", "OID 17000"}
	for _, option := range options {
		for _, middle := range []string{" ", " LC_COLLATE 'C' "} {
			sql := "CREATE DATABASE db " + option + middle + option
			t.Run(sql, func(t *testing.T) {
				_, err := ParseOne(sql)
				if err == nil || pgerror.GetPGCode(err) != pgcode.Syntax || !strings.Contains(err.Error(), "conflicting or redundant options") {
					t.Fatalf("expected duplicate option syntax error, got %v", err)
				}
			})
		}
	}
}

func TestCreateDatabaseInvalidOptions(t *testing.T) {
	for _, sql := range []string{"CREATE DATABASE db BOGUS 'x'", "CREATE DATABASE db ENCODING", "CREATE DATABASE db TEMPLATE =", "CREATE DATABASE db TEMPLATE template0, ENCODING 'UTF8'"} {
		if _, err := ParseOne(sql); err == nil {
			t.Fatalf("accepted %s", sql)
		}
	}
}
