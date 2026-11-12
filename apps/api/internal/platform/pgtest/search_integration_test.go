//go:build integration

package pgtest

import (
	"testing"
)

func TestExtensionsAndImmutableUnaccent(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	for _, ext := range []string{"pg_trgm", "unaccent"} {
		var n int
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM pg_extension WHERE extname = $1`, ext).Scan(&n); err != nil || n != 1 {
			t.Errorf("extension %s installed = %d, %v; want 1", ext, n, err)
		}
	}

	var volatility string
	if err := conn.QueryRow(ctx, `SELECT provolatile::text FROM pg_proc WHERE oid = 'public.f_unaccent(text)'::regprocedure`).Scan(&volatility); err != nil {
		t.Fatalf("f_unaccent(text) missing: %v", err)
	}
	if volatility != "i" {
		t.Errorf("f_unaccent provolatile = %q, want \"i\" (IMMUTABLE)", volatility)
	}

	var got string
	if err := conn.QueryRow(ctx, `SELECT f_unaccent('éléphant')`).Scan(&got); err != nil || got != "elephant" {
		t.Errorf("f_unaccent('éléphant') = %q, %v; want elephant", got, err)
	}
	if err := conn.QueryRow(ctx, `SELECT f_unaccent('Côte d’Ivoire à Ça')`).Scan(&got); err != nil || got != "Cote d'Ivoire a Ca" {
		t.Errorf("f_unaccent folding (the typographic apostrophe folds to ASCII) = %q, %v", got, err)
	}
	var isNull bool
	if err := conn.QueryRow(ctx, `SELECT f_unaccent(NULL) IS NULL`).Scan(&isNull); err != nil || !isNull {
		t.Errorf("f_unaccent(NULL) must be NULL (STRICT): %v %v", isNull, err)
	}

	// An immutable wrapper must be usable in a generated column and an index.
	for _, stmt := range []string{
		`CREATE TEMP TABLE unaccent_probe (t text, d text GENERATED ALWAYS AS (f_unaccent(t)) STORED)`,
		`CREATE INDEX unaccent_probe_idx ON unaccent_probe (f_unaccent(t))`,
	} {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Errorf("%s: %v", stmt, err)
		}
	}
}

func TestFrenchAccentMatchingBothDirections(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, fixtureGraph); err != nil {
		t.Fatalf("fixture graph: %v", err)
	}
	// Accented document, unaccented query, and the reverse.
	if _, err := tx.Exec(ctx, `
		INSERT INTO questions (id, author_id, title, body_md, language) VALUES
		 ('`+qAccent+`', '`+uAlice+`', 'Éléphant et café', 'Le déploiement échoue', 'fr'),
		 ('`+qPlain+`', '`+uAlice+`', 'elephant and cafe', 'deployment fails', 'en')`); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name  string
		query string
		want  []string
	}{
		{"unaccented query hits accented title", "elephant", []string{qAccent, qPlain}},
		{"accented query hits unaccented title", "éléphant", []string{qAccent, qPlain}},
		{"uppercase accented query", "ÉLÉPHANT", []string{qAccent, qPlain}},
		{"accented body term", "deploiement", []string{qAccent}},
		{"phrase from the english row only", "deployment", []string{qPlain}},
		{"no match", "kubernetes", nil},
		{"sql-looking input is data", `'; DROP TABLE questions; --`, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, err := tx.Query(ctx, `
				SELECT id::text FROM questions
				WHERE id IN ($2, $3) AND search_document @@ websearch_to_tsquery('simple', f_unaccent($1))
				ORDER BY id::text`, tc.query, qAccent, qPlain)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var got []string
			for rows.Next() {
				var id string
				if err := rows.Scan(&id); err != nil {
					t.Fatal(err)
				}
				got = append(got, id)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("query %q matched %v, want %v", tc.query, got, tc.want)
			}
			for _, w := range tc.want {
				found := false
				for _, g := range got {
					found = found || g == w
				}
				if !found {
					t.Fatalf("query %q matched %v, want %v", tc.query, got, tc.want)
				}
			}
		})
	}

	// The trigram index expression uses the same wrapper.
	var n int
	if err := tx.QueryRow(ctx, `
		SELECT count(*) FROM questions WHERE f_unaccent(title) ILIKE '%' || f_unaccent($1) || '%'`, "cafe").Scan(&n); err != nil || n < 2 {
		t.Errorf("trigram-style ILIKE over f_unaccent(title) = %d, %v; want >= 2", n, err)
	}
}
