//go:build integration

package pgtest

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"
)

// p0Tables is the authoritative P0 inventory of DATABASE_DESIGN.md 2.1: 41
// tables minus the 3 deferred to P1 (learn_progress, resource_reviews,
// content_reports).
var p0Tables = []string{
	"answer_sources", "answers", "applications", "audit_log", "contribution_needs",
	"content_relations", "countries", "email_verifications", "error_reports",
	"idempotency_keys", "learn_article_tags", "learn_articles", "learn_topics",
	"need_skills", "password_resets", "profiles", "project_maintainers",
	"project_tags", "projects", "providers", "question_sources", "question_tags",
	"questions", "rate_limit_buckets", "resource_categories", "resource_countries",
	"resource_sdk_languages", "resource_sources", "resource_tags", "resources",
	"sessions", "tags", "training_lessons", "training_series", "user_credentials",
	"user_skills", "users", "votes",
}

func TestFreshMigrationToHead(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()

	database := createDB(ctx, t, "")
	db, err := openSQL(pg.dsn(database))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	p, err := newProvider(db, migrationsFS(t))
	if err != nil {
		t.Fatalf("goose provider (expected migrations 00001..00010 in db/migrations): %v", err)
	}

	results, err := p.Up(ctx)
	if err != nil {
		t.Fatalf("goose up on an empty database: %v", err)
	}
	if len(results) != 10 {
		t.Fatalf("applied %d migrations, want 10", len(results))
	}
	for i, r := range results {
		if r.Source.Version != int64(i+1) {
			t.Errorf("result %d has version %d, want %d", i, r.Source.Version, i+1)
		}
		if r.Error != nil {
			t.Errorf("migration %d: %v", r.Source.Version, r.Error)
		}
	}
	if v, err := p.GetDBVersion(ctx); err != nil || v != 10 {
		t.Fatalf("GetDBVersion = %d, %v; want 10", v, err)
	}

	// Second run is a no-op.
	again, err := p.Up(ctx)
	if err != nil || len(again) != 0 {
		t.Fatalf("second Up = %d results, %v; want no-op", len(again), err)
	}

	// goose version table: baseline row 0 plus versions 1..10, all applied.
	rows, err := db.QueryContext(ctx, `SELECT version_id FROM goose_db_version WHERE is_applied ORDER BY version_id`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var got []int64
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			t.Fatal(err)
		}
		got = append(got, v)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if want := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10}; !slices.Equal(got, want) {
		t.Fatalf("goose_db_version applied = %v, want %v", got, want)
	}
}

func TestSchemaHasExactlyTheP0Tables(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	rows, err := conn.Query(ctx, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = 'public' AND table_type = 'BASE TABLE' AND table_name <> 'goose_db_version'
		ORDER BY table_name COLLATE "C"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var got []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			t.Fatal(err)
		}
		got = append(got, n)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(p0Tables) != 38 {
		t.Fatalf("test inventory has %d tables, want 38", len(p0Tables))
	}
	want := slices.Sorted(slices.Values(p0Tables)) // byte order, same as COLLATE "C"
	if !slices.Equal(got, want) {
		t.Fatalf("public tables = %v\nwant %v", got, want)
	}
	for _, deferred := range []string{"learn_progress", "resource_reviews", "content_reports"} {
		if slices.Contains(got, deferred) {
			t.Errorf("P1 table %s must not exist yet", deferred)
		}
	}
}

func TestReferenceRowsAreLoaded(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	// countries: see TestCountriesAreTheFiftyFourAfricanStates.
	var fr, en string
	if err := conn.QueryRow(ctx, `SELECT name_fr, name_en FROM countries WHERE code = 'CI'`).Scan(&fr, &en); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(fr, "Côte") || !strings.Contains(en, "Côte") {
		t.Errorf("CI labels = %q / %q", fr, en)
	}
}

// Owner-approved resource categories (decision of 2026-11-11), in sort order.
var resourceCategoryCodes = []string{
	"mobile-money", "online-checkout", "cross-border", "open-finance", "e-invoicing",
	"identity-kyc", "sms-ussd", "chat-messaging", "maps-addressing", "logistics",
	"open-gov", "energy-climate", "ai-language", "cloud-hosting", "dev-tools",
}

// Owner-approved learn topics (decision of 2026-11-11).
var learnTopicCodes = []string{
	"ai", "payments", "ussd-messaging", "mobile", "api-integration",
	"security-privacy", "cloud-devops", "geo-data", "open-source",
}

func TestResourceCategoriesAreTheOwnerApprovedList(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	rows, err := conn.Query(ctx, `SELECT code, name_fr, name_en, sort_order FROM resource_categories ORDER BY sort_order, code`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var code, nameFR, nameEN string
		var order int16
		if err := rows.Scan(&code, &nameFR, &nameEN, &order); err != nil {
			t.Fatal(err)
		}
		codes = append(codes, code)
		if want := int16(10 * len(codes)); order != want {
			t.Errorf("%s sort_order = %d, want %d (step 10 in list order)", code, order, want)
		}
		if strings.TrimSpace(nameFR) == "" || strings.TrimSpace(nameEN) == "" {
			t.Errorf("%s has an empty name: %q / %q", code, nameFR, nameEN)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(codes) != 15 {
		t.Errorf("resource_categories has %d rows, want 15", len(codes))
	}
	if !slices.Equal(codes, resourceCategoryCodes) {
		t.Errorf("resource_categories codes in sort order = %v\nwant %v", codes, resourceCategoryCodes)
	}

	var nameFR, nameEN string
	if err := conn.QueryRow(ctx, `SELECT name_fr, name_en FROM resource_categories WHERE code = 'cross-border'`).Scan(&nameFR, &nameEN); err != nil {
		t.Fatal(err)
	}
	if nameFR != "Transferts transfrontaliers et change" || nameEN != "Cross-border payouts & FX" {
		t.Errorf("cross-border names = %q / %q", nameFR, nameEN)
	}
}

func TestLearnTopicsAreTheOwnerApprovedList(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	rows, err := conn.Query(ctx, `SELECT code, name_fr, name_en FROM learn_topics ORDER BY code COLLATE "C"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var codes []string
	for rows.Next() {
		var code, nameFR, nameEN string
		if err := rows.Scan(&code, &nameFR, &nameEN); err != nil {
			t.Fatal(err)
		}
		codes = append(codes, code)
		if strings.TrimSpace(nameFR) == "" || strings.TrimSpace(nameEN) == "" {
			t.Errorf("%s has an empty name: %q / %q", code, nameFR, nameEN)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(codes) != 9 {
		t.Errorf("learn_topics has %d rows, want 9", len(codes))
	}
	if want := slices.Sorted(slices.Values(learnTopicCodes)); !slices.Equal(codes, want) {
		t.Errorf("learn_topics codes = %v\nwant %v", codes, want)
	}

	// The typographic apostrophe is part of the approved French label.
	var nameFR, nameEN string
	if err := conn.QueryRow(ctx, `SELECT name_fr, name_en FROM learn_topics WHERE code = 'api-integration'`).Scan(&nameFR, &nameEN); err != nil {
		t.Fatal(err)
	}
	if nameFR != "Intégration d’API et webhooks" || nameEN != "API integration & webhooks" {
		t.Errorf("api-integration names = %q / %q", nameFR, nameEN)
	}
}

func TestIDsDefaultToUUIDv7(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	rows, err := conn.Query(ctx, `
		SELECT table_name, column_default FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'id' AND data_type = 'uuid'
		ORDER BY table_name COLLATE "C"`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var table string
		var def *string
		if err := rows.Scan(&table, &def); err != nil {
			t.Fatal(err)
		}
		n++
		if def == nil || *def != "uuidv7()" {
			t.Errorf("%s.id default = %v, want uuidv7()", table, def)
		}
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if n != 21 {
		t.Errorf("found %d uuid id columns, want 21", n)
	}

	var version int
	err = conn.QueryRow(ctx, `
		INSERT INTO users (email) VALUES ('v7@example.org') RETURNING uuid_extract_version(id)`).Scan(&version)
	if err != nil || version != 7 {
		t.Fatalf("users.id version = %d, %v; want 7", version, err)
	}
}

func TestGeneratedColumnsAreStored(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	// PostgreSQL 18 makes a bare GENERATED ALWAYS AS column VIRTUAL; keys and
	// GIN indexes need STORED (attgenerated = 's').
	rows, err := conn.Query(ctx, `
		SELECT c.relname || '.' || a.attname, a.attgenerated::text
		FROM pg_attribute a JOIN pg_class c ON c.oid = a.attrelid
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind = 'r' AND a.attgenerated <> '' AND NOT a.attisdropped
		ORDER BY 1`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	got := map[string]string{}
	for rows.Next() {
		var col, kind string
		if err := rows.Scan(&col, &kind); err != nil {
			t.Fatal(err)
		}
		got[col] = kind
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"answers.search_document", "content_relations.source_id", "content_relations.target_id",
		"learn_articles.search_document", "projects.search_document", "questions.search_document",
		"resources.search_document", "votes.target_id",
	}
	if len(got) != len(want) {
		t.Errorf("generated columns = %v, want exactly %v", got, want)
	}
	for _, col := range want {
		if got[col] != "s" {
			t.Errorf("%s attgenerated = %q, want \"s\" (STORED)", col, got[col])
		}
	}
}

func TestRateLimitBucketsIsUnlogged(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	var persistence string
	if err := conn.QueryRow(ctx, `SELECT relpersistence::text FROM pg_class WHERE oid = 'public.rate_limit_buckets'::regclass`).Scan(&persistence); err != nil {
		t.Fatal(err)
	}
	if persistence != "u" {
		t.Fatalf("rate_limit_buckets relpersistence = %q, want \"u\" (UNLOGGED)", persistence)
	}
}

// Owner decision: countries holds the complete list of the 54 African states
// (ISO 3166-1 alpha-2), not a subset. Western Sahara (EH) is not a UN member
// state and is not included.
func TestCountriesAreTheFiftyFourAfricanStates(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	var total, distinct, badCode, emptyNames int
	err := conn.QueryRow(ctx, `
		SELECT count(*), count(DISTINCT code),
		       count(*) FILTER (WHERE code !~ '^[A-Z]{2}$'),
		       count(*) FILTER (WHERE btrim(name_fr) = '' OR btrim(name_en) = '')
		FROM countries`).Scan(&total, &distinct, &badCode, &emptyNames)
	if err != nil {
		t.Fatal(err)
	}
	if total != 54 || distinct != 54 {
		t.Errorf("countries = %d rows, %d distinct codes; want 54 and 54", total, distinct)
	}
	if badCode != 0 || emptyNames != 0 {
		t.Errorf("countries with a code outside ^[A-Z]{2}$ = %d, with an empty name = %d; want 0 and 0", badCode, emptyNames)
	}

	// Spot checks: accurate French and English names, accents and special cases.
	want := map[string][2]string{
		"CI": {"Côte d’Ivoire", "Côte d’Ivoire"},
		"CD": {"République démocratique du Congo", "Democratic Republic of the Congo"},
		"CG": {"République du Congo", "Republic of the Congo"},
		"EG": {"Égypte", "Egypt"},
		"ZA": {"Afrique du Sud", "South Africa"},
		"SS": {"Soudan du Sud", "South Sudan"},
		"SD": {"Soudan", "Sudan"},
		"ST": {"Sao Tomé-et-Principe", "São Tomé and Príncipe"},
		"CV": {"Cap-Vert", "Cabo Verde"},
		"SZ": {"Eswatini", "Eswatini"},
		"BF": {"Burkina Faso", "Burkina Faso"},
	}
	for code, names := range want {
		var fr, en string
		if err := conn.QueryRow(ctx, `SELECT name_fr, name_en FROM countries WHERE code = $1`, code).Scan(&fr, &en); err != nil {
			t.Errorf("country %s: %v", code, err)
			continue
		}
		if fr != names[0] || en != names[1] {
			t.Errorf("country %s = %q / %q, want %q / %q", code, fr, en, names[0], names[1])
		}
	}
	for _, notAfrican := range []string{"EH", "FR", "US", "RE"} {
		var n int
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM countries WHERE code = $1`, notAfrican).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s must not be in the African country list: %d, %v", notAfrican, n, err)
		}
	}
}
