package functions

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

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

func TestIdentityExpressionsUseSelectedRole(t *testing.T) {
	auth.Init(nil, nil)
	loginName, _ := auth.GetSuperUserAndPassword()
	var login, selected auth.Role
	auth.LockWrite(func() {
		login, _ = auth.LookupRole(loginName)
		selected = auth.CreateDefaultRole("selected")
		auth.SetRole(selected)
	})
	ctx := sql.NewContext(context.Background(), sql.WithSession(&dsess.DoltSession{Session: sql.NewBaseSession()}))
	if err := core.InitializeIdentity(ctx, sessionstate.RoleID(login.ID()), login.IsSuperUser); err != nil {
		t.Fatal(err)
	}
	if err := core.ApplyAuthorizedIdentityChange(ctx, false, func(id *sessionstate.Identity) error {
		id.SelectRole(sessionstate.RoleID(selected.ID()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"current_user", "current_role", "user"} {
		got, err := identityName(ctx, name)
		if err != nil || got != "selected" {
			t.Fatalf("%s: %q, %v", name, got, err)
		}
	}
	if got, err := identityName(ctx, "session_user"); err != nil || got != loginName {
		t.Fatalf("session_user: %q, %v", got, err)
	}
	auth.LockWrite(func() { auth.RenameRole("selected", "renamed") })
	if got, err := identityName(ctx, "current_user"); err != nil || got != "renamed" {
		t.Fatalf("renamed role: %q, %v", got, err)
	}
	auth.LockWrite(func() {
		auth.DropRole("renamed")
		auth.SetRole(auth.CreateDefaultRole("renamed"))
	})
	if _, err := identityName(ctx, "current_user"); err == nil {
		t.Fatal("deleted ID resolved to replacement role")
	}
}

func TestDoltProcedureAdminGateUsesEffectiveRole(t *testing.T) {
	auth.Init(nil, nil)
	loginName, _ := auth.GetSuperUserAndPassword()
	var actor auth.Role
	auth.LockWrite(func() {
		actor = auth.CreateDefaultRole("actor")
		auth.SetRole(actor)
	})
	ctx := sql.NewContext(context.Background(), sql.WithSession(&dsess.DoltSession{Session: sql.NewBaseSession()}))
	if err := auth.InitializeSessionIdentity(ctx.Session, loginName); err != nil {
		t.Fatal(err)
	}
	selectRole := func(target sessionstate.RoleID) {
		t.Helper()
		if err := core.ApplyAuthorizedIdentityChange(ctx, false, func(next *sessionstate.Identity) error {
			next.SelectRole(target)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	procedure := &plan.ExternalProcedure{ExternalStoredProcedureDetails: sql.ExternalStoredProcedureDetails{ReadOnly: true, AdminOnly: true}}
	selectRole(sessionstate.RoleID(actor.ID()))
	if err := checkDoltProcedureAccess(ctx, procedure); err != ErrDoltProcedurePermissionDenied {
		t.Fatalf("ordinary role access = %v", err)
	}
	selectRole(0)
	if err := checkDoltProcedureAccess(ctx, procedure); err != nil {
		t.Fatalf("login superuser access = %v", err)
	}
}
