//go:build integration

package pgtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/platform/dbgen"
)

// PR #83 review (2026-11-13): the row budget of the batched catalogue queries must be per resource, so one
// resource with many links cannot starve the others on the same page. (The review's other point, a related
// resource leaking through an unpublished provider, cannot happen: content_relations_pair_ck allows no
// resource-to-resource link, only resource to project and resource to article.)

func pgUUID(s string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true} }

func TestCatalogueBatchedQueriesCapPerResource(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	conn := migratedDB(ctx, t)
	seedCatalogue(ctx, t, conn)
	stmts := []string{
		// Alpha: 25 more projects it links to (related), 25 more questions that link to it (discussions), 25 more SDKs.
		`INSERT INTO projects (owner_id, name, summary, language)
		   SELECT '` + catUser + `', 'P ' || g, 'Un projet.', 'fr' FROM generate_series(1, 25) g`,
		`INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_project_id, created_by)
		   SELECT 'resource', '` + catAlpha + `', 'project', id, '` + catUser + `' FROM projects WHERE name LIKE 'P %'`,
		`INSERT INTO questions (author_id, title, body_md, language)
		   SELECT '` + catUser + `', 'Q ' || g, 'body', 'fr' FROM generate_series(1, 25) g`,
		`INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by)
		   SELECT 'question', id, 'resource', '` + catAlpha + `', '` + catUser + `' FROM questions WHERE title LIKE 'Q %'`,
		`INSERT INTO resource_sdk_languages (resource_id, language, name, package_url)
		   SELECT '` + catAlpha + `', 'go', 'SDK ' || g, 'https://pkg.example/' || g FROM generate_series(1, 25) g`,
		// Beta has one of each: it must not be starved by Alpha.
		`INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_project_id, created_by)
		   VALUES ('resource', '` + catBeta + `', 'project', '` + catProject + `', '` + catUser + `')`,
		`INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, created_by)
		   VALUES ('question', '` + catQuestion + `', 'resource', '` + catBeta + `', '` + catUser + `')`,
		`INSERT INTO resource_sdk_languages (resource_id, language, name, package_url)
		   VALUES ('` + catBeta + `', 'js', 'Beta SDK', 'https://pkg.example/beta')`,
	}
	for _, stmt := range stmts {
		if _, err := conn.Exec(ctx, stmt); err != nil {
			t.Fatalf("seed: %v\n%s", err, stmt)
		}
	}
	q := dbgen.New(conn)
	ids := []pgtype.UUID{pgUUID(catAlpha), pgUUID(catBeta)}
	alpha, beta := pgUUID(catAlpha).Bytes, pgUUID(catBeta).Bytes

	related, err := q.CatalogueRelated(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	count := map[[16]byte]int{}
	for _, r := range related {
		count[r.SrcResourceID.Bytes]++
	}
	if count[alpha] != 20 || count[beta] != 1 {
		t.Fatalf("related per resource = alpha %d beta %d, want 20 and 1", count[alpha], count[beta])
	}

	sdks, err := q.CatalogueSDKs(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	count = map[[16]byte]int{}
	for _, s := range sdks {
		count[s.ResourceID.Bytes]++
	}
	if count[alpha] != 20 || count[beta] != 1 {
		t.Fatalf("sdks per resource = alpha %d beta %d, want 20 and 1", count[alpha], count[beta])
	}

	discussions, err := q.CatalogueDiscussions(ctx, ids)
	if err != nil {
		t.Fatal(err)
	}
	count = map[[16]byte]int{}
	for _, d := range discussions {
		count[d.ResourceID.Bytes]++
	}
	if count[alpha] != 20 || count[beta] != 1 {
		t.Fatalf("discussions per resource = alpha %d beta %d, want 20 and 1", count[alpha], count[beta])
	}
}
