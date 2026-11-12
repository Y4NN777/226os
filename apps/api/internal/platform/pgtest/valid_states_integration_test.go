//go:build integration

package pgtest

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestFixtureGraphInsertsInEveryTable(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	tx := begin(t, conn)
	mustExec(t, tx, fixtureGraph)

	for _, table := range p0Tables {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+pgx.Identifier{table}.Sanitize()).Scan(&n); err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			t.Errorf("table %s is empty after the fixture graph", table)
		}
	}
}

// Positive controls: the valid counterpart of the violations above must pass,
// otherwise a constraint could be failing for the wrong reason.
func TestValidStatesAreAccepted(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)

	cases := []struct {
		name  string
		stmts []string
	}{
		{"accepted answer of the same question", []string{
			`UPDATE questions SET accepted_answer_id = '` + aBob + `' WHERE id = '` + qAlice + `'`}},
		{"accepted answer can be cleared again", []string{
			`UPDATE questions SET accepted_answer_id = '` + aBob + `' WHERE id = '` + qAlice + `'`,
			`UPDATE questions SET accepted_answer_id = NULL WHERE id = '` + qAlice + `'`}},
		{"verified resource with its own source", []string{
			`UPDATE resources SET verification_status = 'verified', verified_at = now(), verified_source_id = '` + rsKal + `' WHERE id = '` + rKal + `'`}},
		{"provider claim source without check date", []string{
			`INSERT INTO resource_sources (resource_id, kind, title, url) VALUES ('` + rKal + `', 'claim', 'Provider claim', 'https://example.org/claim')`}},
		{"same user votes on a question and an answer", []string{
			`INSERT INTO votes (user_id, target_type, answer_id, value) VALUES ('` + uBob + `', 'answer', '` + aAlice + `', -1)`}},
		{"vote flip through ON CONFLICT", []string{
			`INSERT INTO votes (user_id, target_type, question_id, value) VALUES ('` + uBob + `', 'question', '` + qAlice + `', -1)
			   ON CONFLICT (user_id, target_type, target_id) DO UPDATE SET value = EXCLUDED.value`}},
		{"removed project frees its github url", []string{
			`UPDATE projects SET status = 'removed', removed_at = now() WHERE id = '` + prAlice + `'`,
			`INSERT INTO projects (owner_id, name, summary, language, github_url, stage) VALUES ('` + uBob + `', 'Fork', 's', 'en', 'https://github.com/example/webhook-kit', 'active')`}},
		{"project without url while idea", []string{
			`INSERT INTO projects (owner_id, name, summary, language, stage) VALUES ('` + uBob + `', 'Another idea', 's', 'en', 'idea')`}},
		{"resolved error report frees the open slot", []string{
			`UPDATE error_reports SET status = 'resolved', handled_by = '` + uMod + `', handled_at = now() WHERE id = '` + erKal + `'`,
			`INSERT INTO error_reports (resource_id, reporter_id, kind) VALUES ('` + rKal + `', '` + uBob + `', 'broken_link')`}},
		{"other kind of report on the same resource", []string{
			`INSERT INTO error_reports (resource_id, reporter_id, kind) VALUES ('` + rKal + `', '` + uBob + `', 'outdated')`}},
		{"withdrawn application can be reopened", []string{
			`UPDATE applications SET status = 'withdrawn', withdrawn_at = now() WHERE id = '` + apBob + `'`,
			`UPDATE applications SET status = 'pending', withdrawn_at = NULL, message = 'back again' WHERE id = '` + apBob + `'`}},
		{"accepted application with decision", []string{
			`UPDATE applications SET status = 'accepted', decided_by = '` + uAlice + `', decided_at = now() WHERE id = '` + apBob + `'`}},
		{"relation to a project and from an article", []string{
			`INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_project_id, created_by) VALUES ('question', '` + qAlice + `', 'project', '` + prAlice + `', '` + uAlice + `')`,
			`INSERT INTO content_relations (source_type, src_resource_id, target_type, tgt_article_id, relation_kind, created_by) VALUES ('resource', '` + rKal + `', 'article', '` + arAlice + `', 'explains', '` + uAlice + `')`}},
		{"same pair with another relation kind", []string{
			`INSERT INTO content_relations (source_type, src_question_id, target_type, tgt_resource_id, relation_kind, created_by) VALUES ('question', '` + qAlice + `', 'resource', '` + rKal + `', 'uses', '` + uAlice + `')`}},
		{"hidden relation", []string{`UPDATE content_relations SET status = 'hidden' WHERE id = '` + crOne + `'`}},
		{"moderator removal of a question", []string{
			`UPDATE questions SET status = 'removed', removed_at = now() WHERE id = '` + qAlice + `'`}},
		{"deleted user", []string{
			`UPDATE users SET status = 'deleted', deleted_at = now(), email = 'deleted-x@invalid' WHERE id = '` + uMod + `'`}},
		{"retracting a vote (hard delete)", []string{
			`DELETE FROM votes WHERE user_id = '` + uBob + `'`}},
		{"completed idempotency key", []string{
			`UPDATE idempotency_keys SET state = 'completed', response_status = 201, completed_at = now(), response_body = '{"id":"x"}'::jsonb`}},
		{"audit details object", []string{
			`INSERT INTO audit_log (action, details) VALUES ('question.hide', '{"reason":"spam"}')`}},
		{"revoked session", []string{
			`UPDATE sessions SET revoked_at = now(), revoked_reason = 'logout' WHERE id = '` + sAlice + `'`}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tx := begin(t, conn)
			mustExec(t, tx, fixtureGraph)
			for _, s := range tc.stmts {
				mustExec(t, tx, s)
			}
		})
	}
}

func TestSearchDocumentIsSetByTheDatabase(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	tx := begin(t, conn)
	mustExec(t, tx, fixtureGraph)
	var n int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM questions WHERE search_document <> ''::tsvector`).Scan(&n); err != nil || n != 2 {
		t.Fatalf("questions with a search document = %d, %v; want 2", n, err)
	}
	// Weighted: the title (3 words) is A, the body B, positions run on.
	var weights string
	if err := tx.QueryRow(ctx, `SELECT search_document::text FROM questions WHERE id = $1`, qAlice).Scan(&weights); err != nil {
		t.Fatal(err)
	}
	if want := `'webhook':1A`; !strings.Contains(weights, want) {
		t.Errorf("search_document %q does not contain %q", weights, want)
	}
	if want := `'hmac':4B`; !strings.Contains(weights, want) {
		t.Errorf("search_document %q does not contain %q", weights, want)
	}
	// Cannot be written by the application.
	if _, err := tx.Exec(ctx, `UPDATE questions SET search_document = ''::tsvector`); pgErr(err) == nil {
		t.Error("search_document must not be writable")
	}
}

func TestCascadeDeletesAuthArtefactsAndJoinRows(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	tx := begin(t, conn)
	mustExec(t, tx, fixtureGraph)
	mustExec(t, tx, `DELETE FROM idempotency_keys`)

	// uMod has no content: deleting the user is allowed.
	mustExec(t, tx, `INSERT INTO user_credentials (user_id, password_hash) VALUES ('`+uMod+`', '$argon2id$v=19$m=1,t=1,p=1$c$h')`)
	mustExec(t, tx, `INSERT INTO sessions (user_id, token_hash, csrf_token_hash, idle_expires_at, absolute_expires_at)
		VALUES ('`+uMod+`', decode(repeat('0a', 32), 'hex'), decode(repeat('0b', 32), 'hex'), now() + interval '1 hour', now() + interval '1 day')`)
	mustExec(t, tx, `INSERT INTO user_skills (user_id, tag_id) VALUES ('`+uMod+`', '`+tGo+`')`)
	mustExec(t, tx, `INSERT INTO password_resets (user_id, token_hash, expires_at) VALUES ('`+uMod+`', decode(repeat('0c', 32), 'hex'), now() + interval '1 hour')`)
	// the fixture references uMod as audit actor and source checker: detach them.
	mustExec(t, tx, `UPDATE audit_log SET actor_id = NULL`)
	mustExec(t, tx, `UPDATE resource_sources SET checked_by = NULL`)
	mustExec(t, tx, `DELETE FROM users WHERE id = '`+uMod+`'`)
	for _, table := range []string{"user_credentials", "sessions", "user_skills", "password_resets"} {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, uMod).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s rows left for the deleted user = %d, %v; want 0", table, n, err)
		}
	}

	// Join rows follow their parent question.
	mustExec(t, tx, `DELETE FROM votes`)
	mustExec(t, tx, `DELETE FROM answer_sources`)
	mustExec(t, tx, `DELETE FROM answers WHERE question_id = '`+qAlice+`'`)
	mustExec(t, tx, `DELETE FROM content_relations`)
	mustExec(t, tx, `DELETE FROM questions WHERE id = '`+qAlice+`'`)
	for _, table := range []string{"question_tags", "question_sources"} {
		var n int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM `+table+` WHERE question_id = $1`, qAlice).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s rows left for the deleted question = %d, %v; want 0", table, n, err)
		}
	}
}

// training_lessons_position_uq is DEFERRABLE INITIALLY DEFERRED: a swap within
// one committed transaction must succeed, while a real duplicate fails at COMMIT.
func TestLessonReorderSwapCommitsAndDuplicateFailsAtCommit(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	if _, err := conn.Exec(ctx, fixtureGraph); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, `INSERT INTO training_lessons (id, series_id, position, title, body_md, language) VALUES ('`+leTwo+`', '`+seAlice+`', 2, 'Leçon 2', 'c', 'fr')`); err != nil {
		t.Fatal(err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE training_lessons SET position = 2 WHERE id = '`+leOne+`'`); err != nil {
		t.Fatalf("swap step 1 must not fail immediately: %v", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE training_lessons SET position = 1 WHERE id = '`+leTwo+`'`); err != nil {
		t.Fatalf("swap step 2: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit of a position swap: %v", err)
	}

	tx, err = conn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE training_lessons SET position = 1 WHERE id = '`+leOne+`'`); err != nil {
		t.Fatalf("duplicate position must be deferred: %v", err)
	}
	err = tx.Commit(ctx)
	e := pgErr(err)
	if e == nil || e.Code != uniqueViol || e.ConstraintName != "training_lessons_position_uq" {
		t.Fatalf("commit error = %v; want 23505 training_lessons_position_uq", err)
	}
}

// Concurrent duplicate writes: the database, not the application, guarantees
// exactly one row (applications_user_need_uq, votes_pk).
func TestConcurrentDuplicateWritesYieldOneRow(t *testing.T) {
	ctx := t.Context()
	conn := migratedDB(ctx, t)
	if _, err := conn.Exec(ctx, fixtureGraph); err != nil {
		t.Fatal(err)
	}
	if _, err := conn.Exec(ctx, `DELETE FROM applications; DELETE FROM votes`); err != nil {
		t.Fatal(err)
	}
	cfg := conn.Config()

	run := func(t *testing.T, table, stmt string) {
		const workers = 8
		var ok, dup atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		for range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				c, err := pgx.ConnectConfig(ctx, cfg)
				if err != nil {
					t.Error(err)
					return
				}
				defer func() { _ = c.Close(context.Background()) }()
				<-start
				_, err = c.Exec(ctx, stmt)
				switch e := pgErr(err); {
				case err == nil:
					ok.Add(1)
				case e != nil && e.Code == uniqueViol:
					dup.Add(1)
				default:
					t.Errorf("unexpected error: %v", err)
				}
			}()
		}
		close(start)
		wg.Wait()
		if ok.Load() != 1 || dup.Load() != workers-1 {
			t.Fatalf("%s: %d successes, %d duplicates; want 1 and %d", table, ok.Load(), dup.Load(), workers-1)
		}
		var n int
		if err := conn.QueryRow(ctx, `SELECT count(*) FROM `+table).Scan(&n); err != nil || n != 1 {
			t.Fatalf("%s rows = %d, %v; want 1", table, n, err)
		}
	}
	t.Run("applications", func(t *testing.T) {
		run(t, "applications", `INSERT INTO applications (contribution_need_id, user_id, message, language) VALUES ('`+nAlice+`', '`+uBob+`', 'me', 'en')`)
	})
	t.Run("votes", func(t *testing.T) {
		run(t, "votes", `INSERT INTO votes (user_id, target_type, question_id) VALUES ('`+uBob+`', 'question', '`+qAlice+`')`)
	})
}
