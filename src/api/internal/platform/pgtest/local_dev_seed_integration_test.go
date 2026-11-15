//go:build integration

package pgtest

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// Local-dev seeds (db/seeds/local-dev/, loaded by `make seed` for APP_ENV=local-dev only, after db/seeds/*.sql):
// real public GitHub developers from committers.top for the Connect directory. They must load on top of the
// shared seed, be demo accounts with no credentials, be listable by the directory rule, and never fail on a
// username or GitHub login that a live account already holds.

func localDevSeedBatch(t *testing.T) string {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repoRoot(t), "db", "seeds", "local-dev", "*.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatal("no local-dev seed files")
	}
	sort.Strings(paths)
	var sb strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		sb.Write(b)
		sb.WriteString("\n")
	}
	return sb.String()
}

func TestLocalDevSeedConnectCommitters(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	batch := localDevSeedBatch(t)
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("local-dev seed: %v", err)
	}
	const block = `u.id::text BETWEEN '01900000-0000-7000-8000-000000000a01' AND '01900000-0000-7000-8000-000000000aff'`
	n := scalarInt(ctx, t, conn, `SELECT count(*) FROM users u WHERE `+block)
	if n < 100 {
		t.Fatalf("committers users = %d, want at least 100", n)
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM users u WHERE `+block+` AND u.provenance <> 'demo'`); got != 0 {
		t.Fatalf("%d committers users are not demo", got)
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM users u JOIN user_credentials c ON c.user_id = u.id WHERE `+block); got != 0 {
		t.Fatalf("%d committers users have credentials", got)
	}
	// Every seeded user has its profile, open to collaboration and public: the directory rule lists them in local-dev.
	listable := scalarInt(ctx, t, conn, `SELECT count(*) FROM users u JOIN profiles p ON p.user_id = u.id WHERE `+block+`
	  AND u.status = 'active' AND p.open_to_collab AND p.profile_visibility = 'public' AND p.open_to_collab_visibility = 'public'
	  AND p.github_username IS NOT NULL AND p.country_code IS NOT NULL`)
	if listable != n {
		t.Fatalf("listable committers profiles = %d, want %d", listable, n)
	}

	// Idempotent: a second run changes nothing.
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("second local-dev seed: %v", err)
	}
	if again := scalarInt(ctx, t, conn, `SELECT count(*) FROM users u WHERE `+block); again != n {
		t.Fatalf("second run: %d users, want %d", again, n)
	}
}

// A live account that already holds one of the seeded usernames or GitHub logins: that row is skipped, the rest
// loads, and the live profile is untouched.
func TestLocalDevSeedSkipsTakenUsernameAndLogin(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	stmts := []string{
		`INSERT INTO users (id, email) VALUES ('01900000-0000-7000-8000-00000000f001', 'live1@example.org'), ('01900000-0000-7000-8000-00000000f002', 'live2@example.org')`,
		`INSERT INTO profiles (user_id, username, display_name) VALUES ('01900000-0000-7000-8000-00000000f001', 'spwoodcock', 'Live one')`,
		`INSERT INTO profiles (user_id, username, display_name, github_username) VALUES ('01900000-0000-7000-8000-00000000f002', 'live-two', 'Live two', 'TKY0065')`,
	}
	for _, s := range stmts {
		if _, err := conn.Exec(ctx, s); err != nil {
			t.Fatalf("%s: %v", s, err)
		}
	}
	if err := runSeed(ctx, t, conn, localDevSeedBatch(t)); err != nil {
		t.Fatalf("local-dev seed over live accounts: %v", err)
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM profiles WHERE username = 'spwoodcock'`); got != 1 {
		t.Fatalf("spwoodcock profiles = %d, want the live one only", got)
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM profiles WHERE display_name = 'Live one' AND username = 'spwoodcock'`); got != 1 {
		t.Fatal("the live spwoodcock profile was changed")
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM profiles WHERE lower(github_username) = 'tky0065'`); got != 1 {
		t.Fatalf("tky0065 profiles = %d, want the live one only", got)
	}
	if got := scalarInt(ctx, t, conn, `SELECT count(*) FROM users WHERE provenance = 'demo' AND email IN ('spwoodcock@demo.226os.invalid', 'tky0065@demo.226os.invalid')`); got != 0 {
		t.Fatalf("%d orphan demo users for taken names", got)
	}
}
