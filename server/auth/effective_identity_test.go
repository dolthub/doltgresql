package auth

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
)

func TestEffectiveRoleAuthorizationAndDoltPrincipal(t *testing.T) {
	Init(nil, nil)
	loginName, _ := GetSuperUserAndPassword()
	var login, ordinary, privileged Role
	LockWrite(func() {
		login, _ = LookupRole(loginName)
		ordinary = CreateDefaultRole("ordinary")
		privileged = CreateDefaultRole("privileged")
		privileged.IsSuperUser = true
		SetRole(ordinary)
		SetRole(privileged)
		AddMemberToGroup(ordinary.ID(), privileged.ID(), false, login.ID())
	})

	sess := &dsess.DoltSession{Session: sql.NewBaseSession()}
	if err := InitializeSessionIdentity(sess, loginName); err != nil {
		t.Fatal(err)
	}
	ctx := sql.NewContext(context.Background(), sql.WithSession(sess))
	selectRole := func(target sessionstate.RoleID) {
		t.Helper()
		if err := core.ApplyAuthorizedIdentityChange(ctx, false, func(next *sessionstate.Identity) error {
			next.SelectRole(target)
			return nil
		}); err != nil {
			t.Fatal(err)
		}
	}
	selectRole(sessionstate.RoleID(ordinary.ID()))
	if got := sess.GetUser(); got != ordinary.Name {
		t.Fatalf("Dolt principal = %q, want %q", got, ordinary.Name)
	}
	state := (&AuthorizationHandler{}).NewQueryState(ctx).(AuthorizationQueryState)
	if state.err != nil || state.role.ID() != ordinary.ID() {
		t.Fatalf("query role = %v, error = %v", state.role, state.err)
	}
	privileges, err := NewPrivilegeSetLayer(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if privileges.Has(sql.PrivilegeType_Super) {
		t.Fatal("selected ordinary role retained login superuser privilege")
	}
	databasePrivileges := privileges.Database("test")
	selectRole(sessionstate.RoleID(login.ID()))
	if !privileges.Has(sql.PrivilegeType_Super) || !databasePrivileges.Has(sql.PrivilegeType_Super) {
		t.Fatal("nested privilege sets did not observe the effective role change")
	}
	selectRole(sessionstate.RoleID(ordinary.ID()))
	if privileges.Has(sql.PrivilegeType_Super) || databasePrivileges.Has(sql.PrivilegeType_Super) {
		t.Fatal("nested privilege sets retained superuser power after role change")
	}
	LockWrite(func() {
		ordinary.IsSuperUser = true
		SetRole(ordinary)
	})
	if !privileges.Has(sql.PrivilegeType_Super) {
		t.Fatal("privilege set did not observe changed role attributes")
	}
	LockWrite(func() {
		ordinary.IsSuperUser = false
		SetRole(ordinary)
	})

	key := TablePrivilegeKey{Role: ordinary.ID(), Table: doltdb.TableName{Name: "records", Schema: "public"}}
	LockRead(func() {
		if HasTablePrivilege(key, Privilege_SELECT) {
			t.Error("ordinary role inherited superuser bypass through membership")
		}
	})
	LockWrite(func() {
		AddTablePrivilege(TablePrivilegeKey{Role: privileged.ID(), Table: key.Table}, GrantedPrivilege{Privilege: Privilege_SELECT, GrantedBy: login.ID()}, false)
	})
	LockRead(func() {
		if !HasTablePrivilege(key, Privilege_SELECT) {
			t.Error("ordinary role did not inherit explicit SELECT privilege")
		}
	})
	LockWrite(func() { RemoveMemberFromGroup(ordinary.ID(), privileged.ID(), false) })
	LockRead(func() {
		if HasTablePrivilege(key, Privilege_SELECT) {
			t.Error("revoked membership still grants SELECT")
		}
	})
	LockWrite(func() { RenameRole("ordinary", "renamed") })
	if got := sess.GetUser(); got != "renamed" {
		t.Fatalf("Dolt principal after rename = %q", got)
	}
	LockWrite(func() { DropRole("renamed") })
	if _, err := CurrentRole(ctx); err == nil {
		t.Fatal("deleted effective role was resolved")
	}
	if _, err := NewPrivilegeSetLayer(ctx); err != nil {
		t.Fatalf("deleted selected role prevented context privilege setup: %v", err)
	}
	if privileges.Has(sql.PrivilegeType_Super) {
		t.Fatal("deleted selected role retained superuser privilege")
	}
	authHandler := &AuthorizationHandler{}
	queryState := authHandler.NewQueryState(ctx)
	if err := queryState.Error(); err != nil {
		t.Fatalf("deleted selected role blocked statement preparation: %v", err)
	}
	if err := authHandler.HandleAuth(ctx, queryState, vitess.AuthInformation{AuthType: AuthType_SELECT}); err == nil {
		t.Fatal("deleted selected role passed an authorization check")
	}
	if got := sess.GetUser(); got != "" {
		t.Fatalf("deleted effective role used Dolt principal %q", got)
	}
	selectRole(0)
	if role, err := CurrentRole(ctx); err != nil || role.ID() != login.ID() {
		t.Fatalf("clearing deleted selection did not restore session role: %v, %v", role, err)
	}
}

func TestMembershipCapabilitiesFollowDistinctPaths(t *testing.T) {
	Init(nil, nil)
	var actor, inherited, alternate, target Role
	LockWrite(func() {
		actor = CreateDefaultRole("actor")
		inherited = CreateDefaultRole("inherited")
		alternate = CreateDefaultRole("alternate")
		target = CreateDefaultRole("target")
		for _, role := range []Role{actor, inherited, alternate, target} {
			SetRole(role)
		}
		AddMemberToGroup(actor.ID(), inherited.ID(), false, actor.ID())
		AddMemberToGroup(actor.ID(), alternate.ID(), false, actor.ID())
		AddMemberToGroup(inherited.ID(), target.ID(), false, actor.ID())
		AddMemberToGroup(alternate.ID(), target.ID(), true, actor.ID())
	})
	LockRead(func() {
		if !CanSetRole(actor.ID(), target.ID()) || !InheritsPrivileges(actor.ID(), target.ID()) || !CanAdministerRole(actor.ID(), target.ID()) {
			t.Error("alternate membership paths did not preserve distinct capabilities")
		}
	})
	LockWrite(func() {
		RemoveMemberFromGroup(alternate.ID(), target.ID(), false)
		actor.InheritPrivileges = false
		SetRole(actor)
	})
	LockRead(func() {
		if !CanSetRole(actor.ID(), target.ID()) {
			t.Error("NOINHERIT incorrectly blocked SET")
		}
		if InheritsPrivileges(actor.ID(), target.ID()) || CanAdministerRole(actor.ID(), target.ID()) {
			t.Error("NOINHERIT granted inherited or ADMIN authority")
		}
	})
}

func TestMembershipCycleCheckUsesStoredEdges(t *testing.T) {
	Init(nil, nil)
	loginName, _ := GetSuperUserAndPassword()
	LockWrite(func() {
		login, _ := LookupRole(loginName)
		ordinary := CreateDefaultRole("ordinary")
		SetRole(ordinary)
		if HasRoleMembership(login.ID(), ordinary.ID()) {
			t.Error("superuser authority was mistaken for a stored grant")
		}
		AddMemberToGroup(login.ID(), ordinary.ID(), false, login.ID())
		func() {
			defer func() {
				if recover() == nil {
					t.Error("cycle through a superuser role was accepted")
				}
			}()
			AddMemberToGroup(ordinary.ID(), login.ID(), false, login.ID())
		}()
	})
}
