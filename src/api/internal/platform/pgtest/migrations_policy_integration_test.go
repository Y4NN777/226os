//go:build integration

package pgtest

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const validHeader = `-- Migration: %s
-- Purpose: test
-- Tables/indexes: t
-- Lock risk: none
-- Data volume: empty
-- Transaction: single transaction
-- Compatibility window: none
-- Restore plan: restore backup
-- Owner: Y4NN77
`

func writeMigration(t *testing.T, dir, name, header, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(header+body), 0o600); err != nil {
		t.Fatal(err)
	}
}

// runCheck runs scripts/check-migrations.sh against dir and returns its exit code.
func runCheck(t *testing.T, dir string, env ...string) (int, string) {
	t.Helper()
	script := filepath.Join(repoRoot(t), "scripts", "check-migrations.sh")
	cmd := exec.CommandContext(t.Context(), "bash", script, dir)
	// MIGRATIONS_BASE_REF is set by CI for the real tree; a scratch directory
	// must only see the value a test passes explicitly.
	var inherited []string
	for _, kv := range os.Environ() {
		if !strings.HasPrefix(kv, "MIGRATIONS_BASE_REF=") {
			inherited = append(inherited, kv)
		}
	}
	cmd.Env = append(inherited, env...)
	out, err := cmd.CombinedOutput()
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return ee.ExitCode(), string(out)
	}
	if err != nil {
		t.Fatalf("cannot run %s: %v\n%s", script, err, out)
	}
	return 0, string(out)
}

func TestCheckMigrationsScript(t *testing.T) {
	const up = "-- +goose Up\nSELECT 1;\n"
	hdr := func(name string) string { return fmt.Sprintf(validHeader, name) }

	cases := []struct {
		name  string
		files map[string]string // file name -> full content
		want  int
	}{
		{"valid sequence", map[string]string{
			"00001_a.sql": hdr("00001_a") + up, "00002_b.sql": hdr("00002_b") + up}, 0},
		{"down section present", map[string]string{
			"00001_a.sql": hdr("00001_a") + up + "-- +goose Down\nSELECT 2;\n"}, 1},
		{"down marker with extra spaces", map[string]string{
			"00001_a.sql": hdr("00001_a") + up + "--  +goose   Down\n"}, 1},
		{"numbering gap", map[string]string{
			"00001_a.sql": hdr("00001_a") + up, "00003_c.sql": hdr("00003_c") + up}, 1},
		{"does not start at 1", map[string]string{
			"00002_b.sql": hdr("00002_b") + up}, 1},
		{"duplicate number", map[string]string{
			"00001_a.sql": hdr("00001_a") + up, "00001_b.sql": hdr("00001_b") + up}, 1},
		{"bad file name", map[string]string{
			"1_a.sql": hdr("1_a") + up}, 1},
		{"uppercase slug", map[string]string{
			"00001_Alpha.sql": hdr("00001_Alpha") + up}, 1},
		{"stray non-sql file", map[string]string{
			"00001_a.sql": hdr("00001_a") + up, "notes.txt": "x"}, 1},
		{"goose token inside a prose comment", map[string]string{
			"00001_a.sql": hdr("00001_a") + "-- note: see '-- +goose NO TRANSACTION' above\n" + up}, 1},
		{"unknown goose annotation", map[string]string{
			"00001_a.sql": hdr("00001_a") + up + "-- +goose Upp\n"}, 1},
		{"valid statement annotations", map[string]string{
			"00001_a.sql": hdr("00001_a") + up + "-- +goose StatementBegin\nSELECT 2;\n-- +goose StatementEnd\n"}, 0},
		{"missing goose up", map[string]string{
			"00001_a.sql": hdr("00001_a") + "SELECT 1;\n"}, 1},
		{"missing header field", map[string]string{
			"00001_a.sql": "-- Purpose: only this\n" + up}, 1},
		{"empty directory", map[string]string{}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			for name, content := range tc.files {
				writeMigration(t, dir, name, "", content)
			}
			got, out := runCheck(t, dir)
			if (got == 0) != (tc.want == 0) {
				t.Fatalf("exit = %d, want %d\n%s", got, tc.want, out)
			}
		})
	}

	t.Run("the repository migrations pass", func(t *testing.T) {
		if code, out := runCheck(t, filepath.Join(repoRoot(t), "db", "migrations")); code != 0 {
			t.Fatalf("exit = %d\n%s", code, out)
		}
	})
}

func TestCheckMigrationsImmutabilityAgainstBaseRef(t *testing.T) {
	repo := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", repo, "-c", "user.name=t", "-c", "user.email=t@example.org", "-c", "commit.gpgsign=false"}, args...)...)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	dir := filepath.Join(repo, "db", "migrations")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		t.Fatal(err)
	}
	const up = "-- +goose Up\nSELECT 1;\n"
	writeMigration(t, dir, "00001_a.sql", fmt.Sprintf(validHeader, "00001_a"), up)
	git("init", "-q", "-b", "main")
	git("add", ".")
	git("commit", "-q", "-m", "base")
	git("checkout", "-q", "-b", "feature")

	// Adding a new migration is allowed.
	writeMigration(t, dir, "00002_b.sql", fmt.Sprintf(validHeader, "00002_b"), up)
	git("add", ".")
	git("commit", "-q", "-m", "add 2")
	if code, out := runCheck(t, dir, "MIGRATIONS_BASE_REF=main"); code != 0 {
		t.Fatalf("adding a migration rejected: exit %d\n%s", code, out)
	}

	// A base ref that cannot be resolved (typo, shallow CI clone) must fail
	// closed, never pass silently.
	if code, out := runCheck(t, dir, "MIGRATIONS_BASE_REF=does-not-exist"); code == 0 {
		t.Fatalf("an unresolvable MIGRATIONS_BASE_REF was accepted (fail-open)\n%s", out)
	}

	// Editing an existing (merged) migration is rejected.
	writeMigration(t, dir, "00001_a.sql", fmt.Sprintf(validHeader, "00001_a"), up+"SELECT 2;\n")
	git("add", ".")
	git("commit", "-q", "-m", "edit 1")
	if code, out := runCheck(t, dir, "MIGRATIONS_BASE_REF=main"); code == 0 {
		t.Fatalf("editing an applied migration was accepted\n%s", out)
	}
}
