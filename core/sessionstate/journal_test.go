package sessionstate

import "testing"

func expectValue(t *testing.T, j *Journal[string], want string) {
	t.Helper()
	if got := *j.Current(); got != want {
		t.Fatalf("value = %q, want %q", got, want)
	}
}

func TestJournalTransactionScopes(t *testing.T) {
	for _, tc := range []struct {
		name               string
		set                func(*Journal[string])
		current, committed string
	}{
		{"session", func(j *Journal[string]) { j.SetSession("session") }, "session", "session"},
		{"local", func(j *Journal[string]) { j.SetLocal("local") }, "local", "initial"},
		{"session then local", func(j *Journal[string]) { j.SetSession("session"); j.SetLocal("local") }, "local", "session"},
		{"local then session", func(j *Journal[string]) { j.SetLocal("local"); j.SetSession("session") }, "session", "session"},
		{"repeated changes", func(j *Journal[string]) { j.SetSession("a"); j.SetLocal("b"); j.SetSession("c"); j.SetLocal("d") }, "d", "c"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := NewJournal("initial")
			j.Begin()
			tc.set(&j)
			expectValue(t, &j, tc.current)
			j.Commit()
			j.Commit()
			expectValue(t, &j, tc.committed)
		})
	}

	j := NewJournal("initial")
	j.SetLocal("ignored")
	expectValue(t, &j, "initial")
	j.Begin()
	j.SetSession("session")
	j.SetLocal("local")
	j.Rollback()
	j.Rollback()
	expectValue(t, &j, "initial")
}

func TestJournalNestedRepeatedSavepoints(t *testing.T) {
	j := NewJournal("initial")
	j.Begin()
	j.SetSession("before")
	j.Savepoint("outer")
	j.SetLocal("outer local")
	j.Savepoint("same")
	j.SetSession("first")
	j.Savepoint("same")
	j.SetLocal("second")
	expectValue(t, &j, "second")
	if !j.RollbackTo("same") {
		t.Fatal("newest savepoint missing")
	}
	expectValue(t, &j, "first")
	j.SetSession("replacement")
	if !j.Release("same") {
		t.Fatal("newest savepoint missing on release")
	}
	if !j.RollbackTo("same") {
		t.Fatal("older repeated savepoint missing")
	}
	expectValue(t, &j, "outer local")
	if !j.Release("outer") {
		t.Fatal("outer savepoint missing")
	}
	if j.RollbackTo("same") {
		t.Fatal("released child savepoint remained")
	}
	j.Commit()
	expectValue(t, &j, "before")
}

func TestJournalRollbackRestoresScopeHistory(t *testing.T) {
	j := NewJournal("initial")
	j.Begin()
	j.SetSession("session")
	j.Savepoint("s")
	j.SetLocal("local")
	j.SetSession("later session")
	if !j.RollbackTo("s") {
		t.Fatal("savepoint missing")
	}
	expectValue(t, &j, "session")
	j.SetLocal("new local")
	j.Commit()
	expectValue(t, &j, "session")
}
