package functions

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
	"github.com/dolthub/doltgresql/server/auth"
)

func TestIdentityExpressionsUseExecutionState(t *testing.T) {
	auth.Init(nil, nil)
	loginName, _ := auth.GetSuperUserAndPassword()
	var login auth.Role
	auth.LockWrite(func() {
		login, _ = auth.LookupRole(loginName)
	})
	ctx := sql.NewContext(context.Background(), sql.WithSession(&dsess.DoltSession{Session: sql.NewBaseSession()}))
	if err := core.InitializeIdentity(ctx, sessionstate.RoleID(login.ID()), login.IsSuperUser); err != nil {
		t.Fatal(err)
	}
	current, err := identityName(ctx, "current_user")
	if err != nil || current != loginName {
		t.Fatalf("login identity: %q, %v", current, err)
	}
	for _, name := range []string{"current_user", "current_role", "user"} {
		got, err := identityName(ctx, name)
		if err != nil || got != loginName {
			t.Fatalf("%s: %q, %v", name, got, err)
		}
	}
	got, err := identityName(ctx, "session_user")
	if err != nil || got != loginName {
		t.Fatalf("session_user: %q, %v", got, err)
	}
	auth.LockWrite(func() { auth.RenameRole(loginName, "renamed") })
	got, err = identityName(ctx, "current_user")
	if err != nil || got != "renamed" {
		t.Fatalf("renamed role: %q, %v", got, err)
	}
	auth.LockWrite(func() {
		auth.DropRole("renamed")
		auth.SetRole(auth.CreateDefaultRole("renamed"))
	})
	if _, err = identityName(ctx, "current_user"); err == nil {
		t.Fatal("deleted ID resolved to replacement role")
	}
}
