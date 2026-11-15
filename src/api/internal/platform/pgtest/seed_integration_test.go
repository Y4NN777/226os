//go:build integration

package pgtest

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// Demo fixtures contract (db/seeds/README.md, SEED #75): the seed files, run the way `make seed` runs
// them (lexical order, one transaction), must load exactly the documented demo data, stay synthetic
// and be idempotent.

// seededTables are the tables whose rows and content are compared across two seed runs.
var seededTables = []string{
	"users", "profiles", "tags",
	"learn_articles", "learn_article_tags", "training_series", "training_lessons",
	"providers", "resources", "resource_sources", "resource_countries", "resource_tags", "resource_sdk_languages",
	"projects", "project_maintainers", "project_tags", "contribution_needs", "need_skills",
	"questions", "answers", "question_tags",
}

func seedFiles(t testing.TB) (names []string, batch string) {
	t.Helper()
	dir := filepath.Join(repoRoot(t), "db", "seeds")
	paths, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) == 0 {
		t.Fatalf("no seed files in %s", dir)
	}
	sort.Strings(paths)
	var sb strings.Builder
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		names = append(names, filepath.Base(p))
		sb.Write(b)
		sb.WriteString("\n")
	}
	return names, sb.String()
}

// runSeed executes the whole batch in ONE transaction over the simple protocol, like
// `psql -X -v ON_ERROR_STOP=1 -1`; any error rolls everything back.
func runSeed(ctx context.Context, t testing.TB, conn *pgx.Conn, batch string) error {
	t.Helper()
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	if _, err := tx.Conn().PgConn().Exec(ctx, batch).ReadAll(); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func seededDB(t *testing.T) (context.Context, *pgx.Conn, string) {
	t.Helper()
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	_, batch := seedFiles(t)
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return ctx, conn, batch
}

func scalarInt(ctx context.Context, t *testing.T, conn *pgx.Conn, q string) int {
	t.Helper()
	var n int
	if err := conn.QueryRow(ctx, q).Scan(&n); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return n
}

func TestSeedCounts(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	tests := []struct {
		name  string
		query string
		want  int
	}{
		{"users", `SELECT count(*) FROM users`, 3},
		{"profiles", `SELECT count(*) FROM profiles`, 3},
		{"tags", `SELECT count(*) FROM tags`, 15},
		{"learn_articles", `SELECT count(*) FROM learn_articles`, 4},
		{"learn_articles all tutorial", `SELECT count(*) FROM learn_articles WHERE kind = 'tutorial'`, 4},
		{"training_series", `SELECT count(*) FROM training_series`, 1},
		{"training_lessons", `SELECT count(*) FROM training_lessons`, 4},
		{"lessons with tutorial article", `SELECT count(*) FROM training_lessons l JOIN learn_articles a ON a.id = l.article_id WHERE a.kind = 'tutorial'`, 4},
		{"lessons with null article", `SELECT count(*) FROM training_lessons WHERE article_id IS NULL`, 0},
		{"lesson positions 1..4", `SELECT count(DISTINCT position) FROM training_lessons WHERE position BETWEEN 1 AND 4`, 4},
		{"providers", `SELECT count(*) FROM providers`, 3},
		{"resources", `SELECT count(*) FROM resources`, 4},
		{"resource_sources", `SELECT count(*) FROM resource_sources`, 4},
		{"projects", `SELECT count(*) FROM projects`, 2},
		{"projects public published", `SELECT count(*) FROM projects WHERE visibility = 'public' AND status = 'published'`, 2},
		{"contribution_needs", `SELECT count(*) FROM contribution_needs`, 4},
		{"contribution_needs open", `SELECT count(*) FROM contribution_needs WHERE status = 'open'`, 4},
		{"questions", `SELECT count(*) FROM questions`, 2},
		{"answers", `SELECT count(*) FROM answers`, 2},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := scalarInt(ctx, t, conn, tc.query); got != tc.want {
				t.Fatalf("%s = %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

func TestSeedProvenanceIsDemo(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	for _, table := range []string{"users", "tags", "providers", "resources", "questions", "answers", "projects", "learn_articles", "training_series"} {
		t.Run(table, func(t *testing.T) {
			if n := scalarInt(ctx, t, conn, fmt.Sprintf(`SELECT count(*) FROM %s WHERE provenance <> 'demo'`, table)); n != 0 {
				t.Fatalf("%s has %d non-demo rows", table, n)
			}
			if n := scalarInt(ctx, t, conn, fmt.Sprintf(`SELECT count(*) FROM %s`, table)); n == 0 {
				t.Fatalf("%s is empty: provenance check would be vacuous", table)
			}
		})
	}
}

func TestSeedDemoAccountsHaveNoCredentials(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	tests := []struct {
		name  string
		query string
	}{
		{"username not demo-", `SELECT count(*) FROM profiles WHERE username NOT LIKE 'demo-%'`},
		{"email outside demo domain", `SELECT count(*) FROM users WHERE email NOT LIKE '%@demo.226os.invalid'`},
		{"user_credentials", `SELECT count(*) FROM user_credentials`},
		{"sessions", `SELECT count(*) FROM sessions`},
		{"email_verifications", `SELECT count(*) FROM email_verifications`},
		{"password_resets", `SELECT count(*) FROM password_resets`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if n := scalarInt(ctx, t, conn, tc.query); n != 0 {
				t.Fatalf("%s: %d offending rows, want 0", tc.name, n)
			}
		})
	}
}

func TestSeedHasNoFakeActivity(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	tests := []struct {
		name  string
		query string
	}{
		{"votes", `SELECT count(*) FROM votes`},
		{"applications", `SELECT count(*) FROM applications`},
		{"accepted answers", `SELECT count(*) FROM questions WHERE accepted_answer_id IS NOT NULL`},
		{"answer_count mismatch", `SELECT count(*) FROM questions q WHERE q.answer_count <> (SELECT count(*) FROM answers a WHERE a.question_id = q.id)`},
		{"question vote_score", `SELECT count(*) FROM questions WHERE vote_score <> 0`},
		{"answer vote_score", `SELECT count(*) FROM answers WHERE vote_score <> 0`},
		{"resource not unverified", `SELECT count(*) FROM resources WHERE verification_status <> 'unverified'`},
		{"resource verified_at", `SELECT count(*) FROM resources WHERE verified_at IS NOT NULL`},
		{"resource verified_source_id", `SELECT count(*) FROM resources WHERE verified_source_id IS NOT NULL`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if n := scalarInt(ctx, t, conn, tc.query); n != 0 {
				t.Fatalf("%s: %d offending rows, want 0", tc.name, n)
			}
		})
	}
}

func TestSeedURLsUseExampleOrg(t *testing.T) {
	ctx, conn, _ := seededDB(t)
	tests := []struct{ table, column string }{
		{"providers", "website_url"},
		{"resources", "website_url"},
		{"resources", "docs_url"},
		{"resource_sources", "url"},
		{"resource_sdk_languages", "package_url"},
		{"resource_sdk_languages", "repo_url"},
		{"projects", "github_url"},
	}
	for _, tc := range tests {
		t.Run(tc.table+"."+tc.column, func(t *testing.T) {
			q := fmt.Sprintf(`SELECT count(*) FROM %s WHERE %s IS NOT NULL AND %s NOT LIKE 'https://example.org/%%'`, tc.table, tc.column, tc.column)
			if n := scalarInt(ctx, t, conn, q); n != 0 {
				t.Fatalf("%s.%s: %d URLs outside https://example.org/", tc.table, tc.column, n)
			}
		})
	}
}

// tableDigest is the row count and an md5 over the ordered text form of every row.
func tableDigest(ctx context.Context, t *testing.T, conn *pgx.Conn, table string) string {
	t.Helper()
	var n int
	var sum string
	q := fmt.Sprintf(`SELECT count(*), md5(coalesce(string_agg(x::text, E'\n' ORDER BY x::text), '')) FROM %s x`, pgx.Identifier{table}.Sanitize())
	if err := conn.QueryRow(ctx, q).Scan(&n, &sum); err != nil {
		t.Fatalf("digest %s: %v", table, err)
	}
	return fmt.Sprintf("%d/%s", n, sum)
}

func TestSeedIsIdempotent(t *testing.T) {
	ctx, conn, batch := seededDB(t)
	before := map[string]string{}
	for _, table := range seededTables {
		before[table] = tableDigest(ctx, t, conn, table)
	}
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("second seed run: %v", err)
	}
	for _, table := range seededTables {
		if after := tableDigest(ctx, t, conn, table); after != before[table] {
			t.Errorf("%s changed on second run: %s -> %s", table, before[table], after)
		}
	}
}

// staging-test already holds live content when it is seeded: a live tag with a slug the seed also uses
// (here `go` and `api`) must be reused, not collide on tags_slug_uq and roll the whole batch back.
func TestSeedReusesLiveTagsBySlug(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	const liveGo, liveAPI = "01900000-0000-7000-8000-00000000ff01", "01900000-0000-7000-8000-00000000ff02"
	if _, err := conn.Exec(ctx, `INSERT INTO tags (id, slug, label) VALUES ($1, 'go', 'Go'), ($2, 'api', 'API')`, liveGo, liveAPI); err != nil {
		t.Fatal(err)
	}
	_, batch := seedFiles(t)
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("seed over live tags: %v", err)
	}
	for _, tc := range []struct{ slug, id string }{{"go", liveGo}, {"api", liveAPI}} {
		var id, provenance string
		if err := conn.QueryRow(ctx, `SELECT id::text, provenance FROM tags WHERE slug = $1`, tc.slug).Scan(&id, &provenance); err != nil {
			t.Fatal(err)
		}
		if id != tc.id || provenance != "live" {
			t.Errorf("tag %s = (%s, %s), want the live row (%s, live)", tc.slug, id, provenance, tc.id)
		}
	}
	if n := scalarInt(ctx, t, conn, `SELECT count(*) FROM resource_tags WHERE tag_id = '`+liveAPI+`'`); n == 0 {
		t.Error("no demo resource is tagged with the live `api` tag")
	}
	if n := scalarInt(ctx, t, conn, `SELECT count(*) FROM tags WHERE provenance = 'demo' AND slug IN ('go', 'api')`); n != 0 {
		t.Errorf("%d demo duplicates of a live slug", n)
	}
	if err := runSeed(ctx, t, conn, batch); err != nil {
		t.Fatalf("second seed over live tags: %v", err)
	}
}

func TestSeedSQLIsDeterministic(t *testing.T) {
	names, _ := seedFiles(t)
	forbidden := regexp.MustCompile(`(?i)now\s*\(|current_timestamp|uuidv7\s*\(|random\s*\(`)
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join(repoRoot(t), "db", "seeds", name))
			if err != nil {
				t.Fatal(err)
			}
			if m := forbidden.Find(b); m != nil {
				t.Fatalf("%s contains non-deterministic construct %q", name, m)
			}
		})
	}
}
