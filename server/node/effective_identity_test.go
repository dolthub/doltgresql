package node

import (
	"context"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
	"github.com/dolthub/doltgresql/server/auth"
)

func TestRoleAdministrationUsesEffectiveRole(t *testing.T) {
	auth.Init(nil, nil)
	loginName, _ := auth.GetSuperUserAndPassword()
	var actor auth.Role
	auth.LockWrite(func() {
		actor = auth.CreateDefaultRole("actor")
		auth.SetRole(actor)
		auth.SetRole(auth.CreateDefaultRole("recipient"))
		auth.SetRole(auth.CreateDefaultRole("target"))
	})
	ctx := sql.NewContext(context.Background(), sql.WithSession(&dsess.DoltSession{Session: sql.NewBaseSession()}))
	if err := auth.InitializeSessionIdentity(ctx.Session, loginName); err != nil {
		t.Fatal(err)
	}
	if err := core.ApplyAuthorizedIdentityChange(ctx, false, func(next *sessionstate.Identity) error {
		next.SelectRole(sessionstate.RoleID(actor.ID()))
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := (&CreateRole{Name: "created"}).RowIter(ctx, nil); err == nil {
		t.Fatal("ordinary selected role created a role using login superuser authority")
	}
	if _, err := (&Grant{ToRoles: []string{"recipient"}, GrantRole: &GrantRole{Groups: []string{"target"}}}).RowIter(ctx, nil); err == nil {
		t.Fatal("ordinary selected role granted membership using login superuser authority")
	}
	if _, err := (&Revoke{FromRoles: []string{"recipient"}, RevokeRole: &RevokeRole{Groups: []string{"target"}}}).RowIter(ctx, nil); err == nil {
		t.Fatal("ordinary selected role revoked membership using login superuser authority")
	}
}
