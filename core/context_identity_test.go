package core

import (
	"context"
	"errors"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/dolthub/go-mysql-server/sql"
	sqltypes "github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/core/sessionstate"
)

type identityTestProvider struct{ dsess.DoltDatabaseProvider }

func (identityTestProvider) DoltDatabases() []dsess.SqlDatabase { return nil }
func (identityTestProvider) FileSystem() filesys.Filesys        { return nil }

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
	if err := ApplyAuthorizedIdentityChange(first, false, func(id *sessionstate.Identity) error {
		id.SelectRole(303)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	current, err := Identity(first)
	if err != nil || current.CurrentRole() != 303 || a.CurrentRole() != 101 || b.CurrentRole() != 202 {
		t.Fatal("identity snapshots did not isolate session state")
	}
}

func TestCacheClearPreservesDurableSessionState(t *testing.T) {
	cv := &contextValues{
		identity:                  sessionstate.NewJournal(sessionstate.NewIdentity(17, true)),
		colls:                     make(map[string]*databaseCollections),
		pgCatalogCache:            struct{}{},
		sessionAdvisoryLockCounts: map[string]int{"lock": 2},
	}
	selected := *cv.identity.Current()
	selected.SelectRole(18)
	cv.identity.SetSession(selected)
	cv.DoltgresSessionCacheClear()
	if cv.colls != nil || cv.pgCatalogCache != nil {
		t.Fatal("disposable caches survived")
	}
	if cv.identity.Current().AuthenticatedRole() != 17 || cv.identity.Current().CurrentRole() != 18 {
		t.Fatal("identity was cleared")
	}
	if cv.sessionAdvisoryLockCounts["lock"] != 2 {
		t.Fatal("session advisory lock count was cleared")
	}
}

func TestIdentityJournalLifecycleAndCleanup(t *testing.T) {
	sess := &dsess.DoltSession{Session: sql.NewBaseSession()}
	ctx := sql.NewContext(context.Background(), sql.WithSession(sess))
	if err := InitializeIdentity(ctx, 17, true); err != nil {
		t.Fatal(err)
	}
	cv, err := getContextValues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	ends := 0
	cv.transactionEndCallbacks = append(cv.transactionEndCallbacks, func() { ends++ })
	cv.DoltgresTransactionStarted()
	if err := ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error { id.SelectRole(18); return nil }); err != nil {
		t.Fatal(err)
	}
	cv.DoltgresSavepointCreated("sp")
	if err := ApplyAuthorizedIdentityChange(ctx, true, func(id *sessionstate.Identity) error { id.SelectRole(19); return nil }); err != nil {
		t.Fatal(err)
	}
	if cv.identity.Current().CurrentRole() != 19 {
		t.Fatal("local selection was not visible")
	}
	cv.DoltgresSavepointRolledBack("sp")
	if cv.identity.Current().CurrentRole() != 18 {
		t.Fatal("savepoint rollback did not restore selection")
	}
	cv.DoltgresTransactionCommitted()
	cv.DoltgresTransactionCommitted()
	cv.DoltgresTransactionEnd()
	cv.DoltgresTransactionEnd()
	if cv.identity.Current().CurrentRole() != 18 || ends != 1 {
		t.Fatal("commit lost session selection or ran cleanup twice")
	}
	cv.DoltgresTransactionStarted()
	if err := ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error { id.SetSessionRole(20); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error { id.SelectRole(21); return errors.New("denied") }); err == nil {
		t.Fatal("failed transition succeeded")
	}
	if cv.identity.Current().CurrentRole() != 20 {
		t.Fatal("failed transition changed role")
	}
	cv.DoltgresTransactionRolledBack()
	if cv.identity.Current().SessionRole() != 17 || cv.identity.Current().CurrentRole() != 18 {
		t.Fatal("rollback did not restore both identity fields")
	}
}

func TestIdentityJournalThroughDoltTransactions(t *testing.T) {
	if _, _, ok := sql.SystemVariables.GetGlobal(dsess.TransactionsDisabledSysVar); !ok {
		sql.SystemVariables.AddSystemVariables([]sql.SystemVariable{&sql.MysqlSystemVariable{
			Name: dsess.TransactionsDisabledSysVar, Scope: sql.GetMysqlScope(sql.SystemVariableScope_Session),
			Dynamic: true, Type: sqltypes.NewSystemBoolType(dsess.TransactionsDisabledSysVar), Default: int8(0),
		}})
	}
	sess := dsess.DefaultSession(identityTestProvider{}, nil)
	ctx := sql.NewContext(context.Background(), sql.WithSession(sess))
	if err := InitializeIdentity(ctx, 17, true); err != nil {
		t.Fatal(err)
	}
	tx, err := sess.StartTransaction(ctx, sql.ReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error { id.SelectRole(18); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := sess.CreateSavepoint(ctx, tx, "s"); err != nil {
		t.Fatal(err)
	}
	if err := ApplyAuthorizedIdentityChange(ctx, true, func(id *sessionstate.Identity) error { id.SelectRole(19); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := sess.RollbackToSavepoint(ctx, tx, "s"); err != nil {
		t.Fatal(err)
	}
	if id, err := Identity(ctx); err != nil || id.CurrentRole() != 18 {
		t.Fatalf("after savepoint rollback: %v, %v", id, err)
	}
	if err := sess.CommitTransaction(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if id, err := Identity(ctx); err != nil || id.CurrentRole() != 18 {
		t.Fatalf("after commit: %v, %v", id, err)
	}
	tx, err = sess.StartTransaction(ctx, sql.ReadWrite)
	if err != nil {
		t.Fatal(err)
	}
	if err := ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error { id.SetSessionRole(20); return nil }); err != nil {
		t.Fatal(err)
	}
	if err := sess.Rollback(ctx, tx); err != nil {
		t.Fatal(err)
	}
	if id, err := Identity(ctx); err != nil || id.SessionRole() != 17 || id.CurrentRole() != 18 {
		t.Fatalf("after rollback: %v, %v", id, err)
	}
}
