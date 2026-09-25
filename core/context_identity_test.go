package core

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/sessionstate"
)

func TestExplicitIdentityInitialization(t *testing.T) {
	newContext := func() *sql.Context {
		return sql.NewContext(context.Background(), sql.WithSession(&dsess.DoltSession{Session: sql.NewBaseSession()}))
	}
	first, second := newContext(), newContext()
	if _, err := Identity(first); err == nil {
		t.Fatal("uninitialized session acquired a role")
	}
	if err := InitializeIdentity(first, 101, true); err != nil {
		t.Fatal(err)
	}
	if err := InitializeIdentity(second, 202, false); err != nil {
		t.Fatal(err)
	}
	if err := InitializeIdentity(first, 303, false); err == nil {
		t.Fatal("session was initialized twice")
	}
	a, err := Identity(first)
	if err != nil {
		t.Fatal(err)
	}
	b, err := Identity(second)
	if err != nil {
		t.Fatal(err)
	}
	if a.CurrentRole() != 101 || b.CurrentRole() != 202 {
		t.Fatal("identities leaked between sessions")
	}
	firstValues, err := getContextValues(first)
	if err != nil {
		t.Fatal(err)
	}
	firstValues.identity.SelectRole(303)
	current, err := Identity(first)
	if err != nil || current.CurrentRole() != 303 || a.CurrentRole() != 101 || b.CurrentRole() != 202 {
		t.Fatal("identity snapshots did not isolate session state")
	}
}

func TestCacheClearPreservesDurableSessionState(t *testing.T) {
	cv := &contextValues{
		identity:                  sessionstate.NewIdentity(17, true),
		colls:                     make(map[string]*databaseCollections),
		pgCatalogCache:            struct{}{},
		sessionAdvisoryLockCounts: map[string]int{"lock": 2},
	}
	cv.identity.SelectRole(18)
	cv.DoltgresSessionCacheClear()
	if cv.colls != nil || cv.pgCatalogCache != nil {
		t.Fatal("disposable caches survived")
	}
	if cv.identity.AuthenticatedRole() != 17 || cv.identity.CurrentRole() != 18 {
		t.Fatal("identity was cleared")
	}
	if cv.sessionAdvisoryLockCounts["lock"] != 2 {
		t.Fatal("session advisory lock count was cleared")
	}
}
