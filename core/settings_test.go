package core

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
)

func TestSettingsJournalScopes(t *testing.T) {
	sess := &dsess.DoltSession{Session: sql.NewBaseSession()}
	ctx := sql.NewContext(context.Background(), sql.WithSession(sess))
	cv, err := getContextValues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	check := func(name string, want any, exists bool) {
		t.Helper()
		got, ok, err := Setting(ctx, name)
		if err != nil || ok != exists || got != want {
			t.Fatalf("%s: got %v, %v, %v; want %v, %v", name, got, ok, err, want, exists)
		}
	}
	check("app.tenant", nil, false)
	if err := SetSetting(ctx, "app.tenant", "ignored", true); err != nil {
		t.Fatal(err)
	}
	check("app.tenant", nil, false)
	cv.DoltgresTransactionStarted()
	if err := SetSetting(ctx, "app.tenant", "a", false); err != nil {
		t.Fatal(err)
	}
	cv.DoltgresSavepointCreated("s")
	if err := SetSetting(ctx, "app.tenant", "b", true); err != nil {
		t.Fatal(err)
	}
	check("app.tenant", "b", true)
	cv.DoltgresSavepointRolledBack("s")
	check("app.tenant", "a", true)
	if err := SetSetting(ctx, "app.tenant", "c", true); err != nil {
		t.Fatal(err)
	}
	cv.DoltgresTransactionCommitted()
	check("app.tenant", "a", true)
	cv.DoltgresTransactionStarted()
	if err := SetSetting(ctx, "app.tenant", "", false); err != nil {
		t.Fatal(err)
	}
	check("app.tenant", "", true)
	cv.DoltgresTransactionRolledBack()
	check("app.tenant", "a", true)
	cv.DoltgresTransactionStarted()
	if err := SetSetting(ctx, "app.tenant", "local", true); err != nil {
		t.Fatal(err)
	}
	cv.DoltgresTransactionCommitted()
	check("app.tenant", "a", true)
}

// TestSettingsJournalIndependentParameters prevents a session SET from
// preserving another parameter's LOCAL value when the transaction commits.
func TestSettingsJournalIndependentParameters(t *testing.T) {
	sess := &dsess.DoltSession{Session: sql.NewBaseSession()}
	ctx := sql.NewContext(context.Background(), sql.WithSession(sess))
	cv, err := getContextValues(ctx)
	if err != nil {
		t.Fatal(err)
	}
	set := func(name, value string, local bool) {
		t.Helper()
		if err := SetSetting(ctx, name, value, local); err != nil {
			t.Fatal(err)
		}
	}
	check := func(name, want string, exists bool) {
		t.Helper()
		got, ok, err := Setting(ctx, name)
		if err != nil || ok != exists || (exists && got != want) {
			t.Fatalf("%s = %v, %t, %v; want %q, %t", name, got, ok, err, want, exists)
		}
	}
	set("app.first", "session", false)
	cv.DoltgresTransactionStarted()
	set("app.first", "local", true)
	set("app.second", "persistent", false)
	cv.DoltgresTransactionCommitted()
	check("app.first", "session", true)
	check("app.second", "persistent", true)
	cv.DoltgresTransactionStarted()
	cv.DoltgresSavepointCreated("before_new_setting")
	set("app.third", "temporary", false)
	cv.DoltgresSavepointRolledBack("before_new_setting")
	check("app.third", "", false)
	cv.DoltgresSavepointCreated("before_new_setting")
	set("app.third", "local", true)
	cv.DoltgresSavepointReleased("before_new_setting")
	cv.DoltgresSavepointRolledBack("before_new_setting")
	check("app.third", "", false)
	cv.DoltgresTransactionCommitted()
	check("app.first", "session", true)
	check("app.second", "persistent", true)
	check("app.third", "", false)
}
