//go:build integration

package pgtest

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/app"
	"github.com/Y4NN777/226os/src/api/internal/identity/repository"
	"github.com/Y4NN777/226os/src/api/internal/identity/service"
	identitytransport "github.com/Y4NN777/226os/src/api/internal/identity/transport"
	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
	"github.com/Y4NN777/226os/src/api/internal/platform/cursor"
	"github.com/Y4NN777/226os/src/api/internal/platform/httpx"
	"github.com/Y4NN777/226os/src/api/internal/platform/idempotency"
	"github.com/Y4NN777/226os/src/api/internal/platform/ratelimit"
	"github.com/Y4NN777/226os/src/api/internal/solve"
)

// US-SOLVE-001 (questions and answers) on the production router over a migrated database: public
// reads of published questions, member publishing, owner-only edits behind If-Match, idempotent
// replay, links, and answer_count kept exact under concurrency.

const (
	solveArticle = "019a0000-0000-7000-8000-0000000000e1"
	solveDraft   = "019a0000-0000-7000-8000-0000000000e2"
	solvePrivate = "019a0000-0000-7000-8000-0000000000e3"
)

func newSolveStack(t *testing.T) *s2Stack {
	t.Helper()
	s := newS2Stack(t.Context(), t, nil)
	svc := service.New(repository.New(s.pool), s.mail, s2Secret)
	codec, err := cursor.New(s2Secret)
	if err != nil {
		t.Fatal(err)
	}
	deps := httpx.Deps{
		Server: app.NewServer(
			app.WithIdentity(identitytransport.New(svc, false, nil)),
			app.WithSolve(solve.NewHandler(s.pool, codec, ratelimit.New(s.pool, s2Secret, nil), nil)),
		),
		Ready:       s.pool,
		Idempotency: idempotency.NewPgStore(s.pool),
	}
	deps.Subjects = identitytransport.NewResolver(svc, false)
	deps.CSRF = identitytransport.CSRFMiddleware(svc, false, nil)
	s.router = httpx.NewRouter(slog.New(slog.NewTextHandler(io.Discard, nil)), deps)
	for _, stmt := range []string{
		`INSERT INTO learn_articles (id, author_id, kind, title, body_md, language, level, status, published_at) VALUES
		  ('` + solveArticle + `', '` + catUser + `', 'article', 'Publié', 'Corps.', 'fr', 'beginner', 'published', now()),
		  ('` + solveDraft + `', '` + catUser + `', 'article', 'Brouillon', 'Corps.', 'fr', 'beginner', 'draft', NULL)`,
		`INSERT INTO projects (id, owner_id, name, summary, language, visibility) VALUES
		  ('` + solvePrivate + `', '` + catUser + `', 'Privé', 'Un projet privé.', 'fr', 'private')`,
	} {
		if _, err := s.conn.Exec(t.Context(), stmt); err != nil {
			t.Fatal(err)
		}
	}
	return s
}

// solveBody is French text with accents, Markdown, raw HTML, a trailing space and mixed newlines: it must
// come back exactly as sent.
const solveBody = "Comment encaisser un paiement ?\r\n\nÉtape 1 : <b>créer</b> le compte.  \n\n```go\nfmt.Println(\"ça marche\")\n```\n"

func questionBody(title string, extra ...string) string {
	b := `{"title":` + jsonString(title) + `,"body_md":` + jsonString(solveBody) + `,"lang":"fr","tags":["go","mobile-money"]`
	for _, e := range extra {
		b += "," + e
	}
	return b + "}"
}

func jsonString(s string) string {
	out, _ := json.Marshal(s)
	return string(out)
}

func (s *s2Stack) questionCount(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM questions WHERE author_id <> $1`, catUser).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) answerRows(t *testing.T, question string) (rows, counter int) {
	t.Helper()
	if err := s.conn.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM answers WHERE question_id = $1), answer_count FROM questions WHERE id = $1`, question).Scan(&rows, &counter); err != nil {
		t.Fatal(err)
	}
	return rows, counter
}

func (s *s2Stack) ask(t *testing.T, m *member, title string, extra ...string) apigen.Question {
	t.Helper()
	rec := s.write(t, m, http.MethodPost, "/api/v1/questions", questionBody(title, extra...), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("ask %q: %d %s", title, rec.Code, rec.Body.String())
	}
	return decodeAs[apigen.Question](t, rec)
}

func (s *s2Stack) answer(t *testing.T, m *member, question, body string) apigen.Answer {
	t.Helper()
	rec := s.write(t, m, http.MethodPost, "/api/v1/questions/"+question+"/answers", `{"body_md":`+jsonString(body)+`,"lang":"fr"}`, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("answer: %d %s", rec.Code, rec.Body.String())
	}
	return decodeAs[apigen.Answer](t, rec)
}

func TestSolveWriteAuthorizationMatrix(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-alice", true, 71)
	bob := s.signIn(t, "solve-bob", true, 72)
	unverified := s.signIn(t, "solve-unverified", false, 73)
	q := s.ask(t, &alice, "Comment encaisser un paiement")
	qid := q.Id.String()
	rows0, count0 := s.answerRows(t, qid)
	stale := map[string]string{"If-Match": `"stale"`}

	// createQuestion: anonymous 401, unverified 403, wrong CSRF 403, bad body 422; nothing created.
	before := s.questionCount(t)
	assertProblem(t, s.write(t, nil, http.MethodPost, "/api/v1/questions", questionBody("Titre valide"), nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, s.write(t, &unverified, http.MethodPost, "/api/v1/questions", questionBody("Titre valide"), nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	assertProblem(t, s.write(t, &bob, http.MethodPost, "/api/v1/questions", questionBody("Titre valide"), map[string]string{"X-CSRF-Token": "wrong"}), http.StatusForbidden, "CSRF_FAILED")
	if n := s.questionCount(t); n != before {
		t.Fatalf("%d questions after denied writes, want %d", n, before)
	}

	// updateQuestion: anonymous 401; a non-owner gets 403 on a visible question whether the tag is stale
	// or right (403 is decided before the tag); an unknown and a removed question are the same 404 even
	// with a stale tag.
	assertProblem(t, s.write(t, nil, http.MethodPatch, "/api/v1/questions/"+qid, `{"title":"Autre titre"}`, stale), http.StatusUnauthorized, "UNAUTHENTICATED")
	got := s.read(t, nil, "/api/v1/questions/"+qid)
	tag := got.Header().Get("ETag")
	if tag == "" {
		t.Fatal("getQuestion sent no ETag")
	}
	assertProblem(t, s.write(t, &bob, http.MethodPatch, "/api/v1/questions/"+qid, `{"title":"Titre de Bob"}`, stale), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.write(t, &bob, http.MethodPatch, "/api/v1/questions/"+qid, `{"title":"Titre de Bob"}`, map[string]string{"If-Match": tag}), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.write(t, &alice, http.MethodPatch, "/api/v1/questions/"+qid, `{"title":"Autre titre"}`, map[string]string{"X-CSRF-Token": "wrong", "If-Match": tag}), http.StatusForbidden, "CSRF_FAILED")
	if _, err := s.conn.Exec(t.Context(), `UPDATE questions SET status = 'removed', removed_at = now() WHERE id = $1`, catQuestion); err != nil {
		t.Fatal(err)
	}
	unknown := s.write(t, &alice, http.MethodPatch, "/api/v1/questions/019a0000-0000-7000-8000-00000000ffff", `{"title":"Autre titre"}`, stale)
	gone := s.write(t, &alice, http.MethodPatch, "/api/v1/questions/"+catQuestion, `{"title":"Autre titre"}`, stale)
	assertProblem(t, unknown, http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, gone, http.StatusNotFound, "NOT_FOUND")
	if problemWithoutInstance(t, unknown) != problemWithoutInstance(t, gone) {
		t.Fatalf("unknown and removed differ: %s / %s", unknown.Body.String(), gone.Body.String())
	}
	if after := s.read(t, nil, "/api/v1/questions/"+qid); after.Header().Get("ETag") != tag {
		t.Fatal("a denied edit changed the question")
	}

	// createAnswer: anonymous 401, unverified 403, wrong CSRF 403, unknown or removed question 404.
	path := "/api/v1/questions/" + qid + "/answers"
	body := `{"body_md":"Une réponse.","lang":"fr"}`
	assertProblem(t, s.write(t, nil, http.MethodPost, path, body, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
	assertProblem(t, s.write(t, &unverified, http.MethodPost, path, body, nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
	assertProblem(t, s.write(t, &bob, http.MethodPost, path, body, map[string]string{"X-CSRF-Token": "wrong"}), http.StatusForbidden, "CSRF_FAILED")
	assertProblem(t, s.write(t, &bob, http.MethodPost, "/api/v1/questions/"+catQuestion+"/answers", body, nil), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.write(t, &bob, http.MethodPost, "/api/v1/questions/019a0000-0000-7000-8000-00000000ffff/answers", body, nil), http.StatusNotFound, "NOT_FOUND")
	if rows, count := s.answerRows(t, qid); rows != rows0 || count != count0 {
		t.Fatalf("denied answers changed %d/%d to %d/%d", rows0, count0, rows, count)
	}
}

func problemWithoutInstance(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	out, _ := json.Marshal(problem(t, rec))
	return string(out)
}

func TestSolveCreateAndReadKeepTheOriginalTextAndResolveLinks(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-read-alice", true, 74)

	rec := s.write(t, &alice, http.MethodPost, "/api/v1/questions", questionBody("Comment encaisser un paiement",
		`"country":"BF"`,
		`"snippets":[{"language":"go","code":"fmt.Println(\"<script>alert(1)</script>\")"}]`,
		`"links":[{"type":"resource","id":"`+catAlpha+`"},{"type":"project","id":"`+catProject+`"},{"type":"article","id":"`+solveArticle+`"},{"type":"article","id":"`+solveArticle+`"}]`), nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create %d %s", rec.Code, rec.Body.String())
	}
	created := decodeAs[apigen.Question](t, rec)
	if rec.Header().Get("Location") != "/api/v1/questions/"+created.Id.String() {
		t.Fatalf("Location %q", rec.Header().Get("Location"))
	}
	if created.State != "open" || created.AnswerCount != 0 || created.VoteScore != 0 || created.AcceptedAnswerId != nil || created.Author.Id.String() != alice.id {
		t.Fatalf("created %+v", created)
	}

	got := s.read(t, nil, "/api/v1/questions/"+created.Id.String())
	if got.Code != http.StatusOK || got.Header().Get("ETag") == "" {
		t.Fatalf("get %d etag %q", got.Code, got.Header().Get("ETag"))
	}
	q := decodeAs[apigen.Question](t, got)
	if q.BodyMd != solveBody || q.Lang != "fr" || *q.Country != "BF" || len(q.Tags) != 2 || q.Tags[0] != "go" {
		t.Fatalf("question %+v", q)
	}
	if len(q.Snippets) != 1 || q.Snippets[0].Language != "go" || !strings.Contains(q.Snippets[0].Code, "<script>") {
		t.Fatalf("snippet stored as given, never run: %+v", q.Snippets)
	}
	kinds := map[string]string{}
	for _, l := range q.Related {
		kinds[string(l.Type)] = l.Title
	}
	if len(q.Related) != 3 || kinds["resource"] != "Alpha Pay" || kinds["project"] != "Caisse" || kinds["article"] != "Publié" {
		t.Fatalf("related %+v", q.Related)
	}
	// A link target that goes away stops being shown.
	if _, err := s.conn.Exec(t.Context(), `UPDATE learn_articles SET status = 'archived' WHERE id = $1`, solveArticle); err != nil {
		t.Fatal(err)
	}
	if again := decodeAs[apigen.Question](t, s.read(t, nil, "/api/v1/questions/"+created.Id.String())); len(again.Related) != 2 {
		t.Fatalf("archived article still linked: %+v", again.Related)
	}
	// Unknown and removed are the same 404 for a reader.
	assertProblem(t, s.read(t, nil, "/api/v1/questions/"+catGoneQ), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.read(t, nil, "/api/v1/questions/019a0000-0000-7000-8000-00000000ffff"), http.StatusNotFound, "NOT_FOUND")
}

func TestSolveCreateValidation(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-validate-alice", true, 75)
	before := s.questionCount(t)
	for name, extra := range map[string]string{
		"private project":     `"links":[{"type":"project","id":"` + solvePrivate + `"}]`,
		"draft article":       `"links":[{"type":"article","id":"` + solveDraft + `"}]`,
		"removed question":    `"links":[{"type":"question","id":"` + catGoneQ + `"}]`,
		"series":              `"links":[{"type":"series","id":"` + catQuestion + `"}]`,
		"unknown resource":    `"links":[{"type":"resource","id":"019a0000-0000-7000-8000-00000000ffff"}]`,
		"unknown country":     `"country":"ZZ"`,
		"two snippets":        `"snippets":[{"language":"go","code":"a"},{"language":"go","code":"b"}]`,
		"snippet language 31": `"snippets":[{"language":"` + strings.Repeat("a", 31) + `","code":"a"}]`,
	} {
		rec := s.write(t, &alice, http.MethodPost, "/api/v1/questions", questionBody("Titre valide", extra), nil)
		if rec.Code != http.StatusUnprocessableEntity {
			t.Errorf("%s: status %d %s", name, rec.Code, rec.Body.String())
		}
	}
	tooMany := `{"title":"Titre valide","body_md":"x","lang":"fr","tags":["a","b","c","d","e","f"]}`
	assertProblem(t, s.write(t, &alice, http.MethodPost, "/api/v1/questions", tooMany, nil), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if n := s.questionCount(t); n != before {
		t.Fatalf("%d questions after rejected writes, want %d", n, before)
	}
	// A link rejected in the transaction leaves neither the question nor its tags behind.
	var tags int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM question_tags`).Scan(&tags); err != nil || tags != 0 {
		t.Fatalf("%d question_tags rows after rejected writes (%v)", tags, err)
	}
}

func TestSolveIdempotentReplay(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-idem-alice", true, 76)
	bob := s.signIn(t, "solve-idem-bob", true, 77)
	key := map[string]string{"Idempotency-Key": "solve-question-retry-1"}

	first := s.write(t, &alice, http.MethodPost, "/api/v1/questions", questionBody("Titre valide"), key)
	second := s.write(t, &alice, http.MethodPost, "/api/v1/questions", questionBody("Titre valide"), key)
	if first.Code != http.StatusCreated || second.Code != http.StatusCreated || second.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("first %d second %d replayed %q", first.Code, second.Code, second.Header().Get("Idempotency-Replayed"))
	}
	if a, b := decodeAs[apigen.Question](t, first), decodeAs[apigen.Question](t, second); a.Id.String() != b.Id.String() {
		t.Fatalf("replay answered another question: %s / %s", a.Id, b.Id)
	}
	if n := s.questionCount(t); n != 1 {
		t.Fatalf("%d questions after a replay, want 1", n)
	}
	assertProblem(t, s.write(t, &alice, http.MethodPost, "/api/v1/questions", questionBody("Autre titre valide"), key), http.StatusUnprocessableEntity, "IDEMPOTENCY_KEY_REUSED")

	qid := decodeAs[apigen.Question](t, first).Id.String()
	akey := map[string]string{"Idempotency-Key": "solve-answer-retry-1"}
	path := "/api/v1/questions/" + qid + "/answers"
	body := `{"body_md":"Une réponse.","lang":"fr"}`
	a1 := s.write(t, &bob, http.MethodPost, path, body, akey)
	a2 := s.write(t, &bob, http.MethodPost, path, body, akey)
	if a1.Code != http.StatusCreated || a2.Code != http.StatusCreated || a2.Header().Get("Idempotency-Replayed") != "true" {
		t.Fatalf("answer first %d second %d", a1.Code, a2.Code)
	}
	if rows, count := s.answerRows(t, qid); rows != 1 || count != 1 {
		t.Fatalf("after a replayed answer: %d rows, answer_count %d", rows, count)
	}
}

func TestSolveUpdateQuestionPreconditions(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-edit-alice", true, 78)
	bob := s.signIn(t, "solve-edit-bob", true, 79)
	q := s.ask(t, &alice, "Titre de départ", `"country":"BF"`, `"links":[{"type":"resource","id":"`+catAlpha+`"}]`)
	path := "/api/v1/questions/" + q.Id.String()
	tag := s.read(t, nil, path).Header().Get("ETag")

	// Missing If-Match 428, stale 412: nothing written.
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"title":"Nouveau titre"}`, nil), http.StatusPreconditionRequired, "PRECONDITION_REQUIRED")
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"title":"Nouveau titre"}`, map[string]string{"If-Match": `"stale"`}), http.StatusPreconditionFailed, "PRECONDITION_FAILED")
	if s.read(t, nil, path).Header().Get("ETag") != tag {
		t.Fatal("a refused edit changed the question")
	}
	// lang is immutable and unknown fields are refused.
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"lang":"en"}`, map[string]string{"If-Match": tag}), http.StatusUnprocessableEntity, "VALIDATION_ERROR")

	// The right tag edits; the merge patch keeps what it does not name, null clears the country, links
	// are replaced, and the ETag changes.
	rec := s.write(t, &alice, http.MethodPatch, path, `{"title":"Nouveau titre","country":null,"tags":["python"],"links":[{"type":"project","id":"`+catProject+`"}]}`, map[string]string{"If-Match": tag})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	updated := decodeAs[apigen.Question](t, rec)
	if updated.Title != "Nouveau titre" || updated.BodyMd != solveBody || updated.Country != nil || len(updated.Tags) != 1 || updated.Tags[0] != "python" ||
		len(updated.Related) != 1 || updated.Related[0].Type != "project" || updated.Lang != "fr" {
		t.Fatalf("updated %+v", updated)
	}
	newTag := rec.Header().Get("ETag")
	if newTag == "" || newTag == tag {
		t.Fatalf("ETag %q after the edit (was %q)", newTag, tag)
	}
	// The old tag is stale now; a rejected link leaves the question untouched.
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"title":"Encore"}`, map[string]string{"If-Match": tag}), http.StatusPreconditionFailed, "PRECONDITION_FAILED")
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"title":"Encore","links":[{"type":"project","id":"`+solvePrivate+`"}]}`, map[string]string{"If-Match": newTag}), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if got := decodeAs[apigen.Question](t, s.read(t, nil, path)); got.Title != "Nouveau titre" || len(got.Related) != 1 {
		t.Fatalf("a rejected edit left %+v", got)
	}
	// A new answer changes the representation, so the tag taken before it is stale.
	s.answer(t, &bob, q.Id.String(), "Une réponse.")
	assertProblem(t, s.write(t, &alice, http.MethodPatch, path, `{"title":"Encore"}`, map[string]string{"If-Match": newTag}), http.StatusPreconditionFailed, "PRECONDITION_FAILED")

	// Two edits with the same tag at once: exactly one wins.
	cur := s.read(t, nil, path).Header().Get("ETag")
	var wg sync.WaitGroup
	codes := make([]int, 6)
	for i := range codes {
		wg.Go(func() {
			codes[i] = s.write(t, &alice, http.MethodPatch, path, fmt.Sprintf(`{"title":"Titre %d valide"}`, i), map[string]string{"If-Match": cur}).Code
		})
	}
	wg.Wait()
	wins := 0
	for _, c := range codes {
		if c == http.StatusOK {
			wins++
		} else if c != http.StatusPreconditionFailed {
			t.Fatalf("codes %v", codes)
		}
	}
	if wins != 1 {
		t.Fatalf("%d edits won with one tag: %v", wins, codes)
	}
}

func TestSolveListQuestionsOrderFiltersAndCursor(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-list-alice", true, 80)
	bob := s.signIn(t, "solve-list-bob", true, 81)
	var ids []string
	for i := range 5 {
		ids = append(ids, s.ask(t, &alice, fmt.Sprintf("Question numéro %d", i)).Id.String())
	}
	// Equal timestamps must still page deterministically (UUID tie-breaker).
	if _, err := s.conn.Exec(t.Context(), `UPDATE questions SET created_at = '2026-11-01T10:00:00Z' WHERE id = ANY($1::uuid[])`, ids[:4]); err != nil {
		t.Fatal(err)
	}
	s.answer(t, &bob, ids[0], "Une réponse.")
	if _, err := s.conn.Exec(t.Context(), `UPDATE questions SET status = 'removed', removed_at = now() WHERE id = $1`, ids[3]); err != nil {
		t.Fatal(err)
	}
	walk := func(query string) (seen []string) {
		cursorValue := ""
		for range 10 {
			path := "/api/v1/questions?limit=2" + query
			if cursorValue != "" {
				path += "&cursor=" + cursorValue
			}
			rec := s.read(t, nil, path)
			if rec.Code != http.StatusOK {
				t.Fatalf("list %s: %d %s", path, rec.Code, rec.Body.String())
			}
			page := decodeAs[apigen.QuestionPage](t, rec)
			for _, q := range page.Items {
				seen = append(seen, q.Id.String())
			}
			if page.NextCursor == nil {
				return seen
			}
			cursorValue = *page.NextCursor
		}
		t.Fatal("cursor never ended")
		return nil
	}
	all := walk("")
	if len(all) != 4 {
		t.Fatalf("listed %v, want 4 published questions (the removed one never)", all)
	}
	if all[0] != ids[4] {
		t.Fatalf("newest first: %v, newest is %s", all, ids[4])
	}
	// The four with one timestamp come back in id descending order, no duplicate, none skipped.
	want := []string{ids[2], ids[1], ids[0]}
	if strings.Join(all[1:], ",") != strings.Join(want, ",") {
		t.Fatalf("tie-break order %v, want %v", all[1:], want)
	}
	if got := walk("&state=answered"); len(got) != 1 || got[0] != ids[0] {
		t.Fatalf("answered %v", got)
	}
	if got := walk("&state=open"); len(got) != 3 {
		t.Fatalf("open %v", got)
	}
	if got := walk("&tag=go"); len(got) != 4 {
		t.Fatalf("tag %v", got)
	}
	if got := walk("&tag=absent"); len(got) != 0 {
		t.Fatalf("tag absent %v", got)
	}
	if got := walk("&q=num%C3%A9ro"); len(got) != 4 {
		t.Fatalf("q %v", got)
	}
	// A cursor is bound to its filters; limit above 50 is 422.
	page := decodeAs[apigen.QuestionPage](t, s.read(t, nil, "/api/v1/questions?limit=1"))
	assertProblem(t, s.read(t, nil, "/api/v1/questions?limit=1&state=open&cursor="+*page.NextCursor), http.StatusBadRequest, "INVALID_CURSOR")
	assertProblem(t, s.read(t, nil, "/api/v1/questions?limit=51"), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
}

func TestSolveAnswersOrderAcceptedFirstThenScoreThenStable(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-ans-alice", true, 82)
	bob := s.signIn(t, "solve-ans-bob", true, 83)
	q := s.ask(t, &alice, "Titre pour les réponses")
	qid := q.Id.String()
	var ids []string
	for i := range 5 {
		a := s.answer(t, &bob, qid, fmt.Sprintf("Réponse %d", i))
		ids = append(ids, a.Id.String())
	}
	if rows, count := s.answerRows(t, qid); rows != 5 || count != 5 {
		t.Fatalf("%d rows, answer_count %d", rows, count)
	}
	// ids[1] accepted with a low score, ids[3] and ids[4] share the top score, the rest 0.
	for _, stmt := range []string{
		`UPDATE answers SET vote_score = 7 WHERE id IN ('` + ids[3] + `', '` + ids[4] + `')`,
		`UPDATE answers SET vote_score = -2 WHERE id = '` + ids[1] + `'`,
		`UPDATE questions SET accepted_answer_id = '` + ids[1] + `' WHERE id = '` + qid + `'`,
		`UPDATE answers SET created_at = created_at + interval '1 second' WHERE id = '` + ids[4] + `'`,
	} {
		if _, err := s.conn.Exec(t.Context(), stmt); err != nil {
			t.Fatal(err)
		}
	}
	var seen []string
	cursorValue := ""
	for range 10 {
		path := "/api/v1/questions/" + qid + "/answers?limit=2"
		if cursorValue != "" {
			path += "&cursor=" + cursorValue
		}
		rec := s.read(t, nil, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("answers %d %s", rec.Code, rec.Body.String())
		}
		page := decodeAs[apigen.AnswerPage](t, rec)
		for _, a := range page.Items {
			seen = append(seen, a.Id.String())
			if a.IsAccepted != (a.Id.String() == ids[1]) || len(a.Snippets) != 0 || a.Sources == nil {
				t.Fatalf("answer %+v", a)
			}
		}
		if page.NextCursor == nil {
			break
		}
		cursorValue = *page.NextCursor
	}
	want := []string{ids[1], ids[3], ids[4], ids[0], ids[2]}
	if strings.Join(seen, ",") != strings.Join(want, ",") {
		t.Fatalf("answer order %v, want %v", seen, want)
	}
	// The question shows its derived state; answers of an unknown or removed question are 404.
	if got := decodeAs[apigen.Question](t, s.read(t, nil, "/api/v1/questions/"+qid)); got.State != "resolved" || got.AnswerCount != 5 || got.AcceptedAnswerId == nil {
		t.Fatalf("question %+v", got)
	}
	assertProblem(t, s.read(t, nil, "/api/v1/questions/"+catGoneQ+"/answers"), http.StatusNotFound, "NOT_FOUND")
	assertProblem(t, s.read(t, nil, "/api/v1/questions/"+qid+"/answers?cursor=garbage"), http.StatusBadRequest, "INVALID_CURSOR")
}

func TestSolveAnswerWithSourcesAndLimits(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-src-alice", true, 84)
	q := s.ask(t, &alice, "Titre pour les sources")
	path := "/api/v1/questions/" + q.Id.String() + "/answers"
	rec := s.write(t, &alice, http.MethodPost, path, `{"body_md":"Voir la doc.","lang":"en","sources":["https://docs.example/a","https://docs.example/a","https://docs.example/b"]}`, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("answer %d %s", rec.Code, rec.Body.String())
	}
	if a := decodeAs[apigen.Answer](t, rec); len(a.Sources) != 2 || a.Lang != "en" || a.QuestionId.String() != q.Id.String() {
		t.Fatalf("answer %+v", a)
	}
	// No snippet column for answers: refuse instead of dropping the code.
	assertProblem(t, s.write(t, &alice, http.MethodPost, path, `{"body_md":"x","lang":"fr","snippets":[{"language":"go","code":"a"}]}`, nil), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	assertProblem(t, s.write(t, &alice, http.MethodPost, path, `{"body_md":"x","lang":"fr","sources":["ftp://example.com/a"]}`, nil), http.StatusUnprocessableEntity, "VALIDATION_ERROR")
	if rows, count := s.answerRows(t, q.Id.String()); rows != 1 || count != 1 {
		t.Fatalf("%d rows, answer_count %d", rows, count)
	}
}

func TestSolveConcurrentAnswersKeepAnswerCountExact(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-race-alice", true, 85)
	q := s.ask(t, &alice, "Titre pour la concurrence")
	qid := q.Id.String()
	const n = 12
	var wg sync.WaitGroup
	codes := make([]int, n)
	for i := range n {
		wg.Go(func() {
			codes[i] = s.write(t, &alice, http.MethodPost, "/api/v1/questions/"+qid+"/answers",
				fmt.Sprintf(`{"body_md":"Réponse concurrente %d","lang":"fr"}`, i), nil).Code
		})
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusCreated {
			t.Fatalf("answer %d: %d", i, c)
		}
	}
	if rows, count := s.answerRows(t, qid); rows != n || count != n {
		t.Fatalf("%d rows, answer_count %d, want %d", rows, count, n)
	}
	// A question removed while answers race in: late answers are refused and the counter equals the rows.
	var wg2 sync.WaitGroup
	for i := range n {
		wg2.Go(func() {
			if i == n/2 {
				if _, err := s.conn.Exec(t.Context(), `UPDATE questions SET status = 'removed', removed_at = now() WHERE id = $1`, qid); err != nil {
					t.Error(err)
				}
				return
			}
			s.write(t, &alice, http.MethodPost, "/api/v1/questions/"+qid+"/answers", fmt.Sprintf(`{"body_md":"Tardive %d","lang":"fr"}`, i), nil)
		})
	}
	wg2.Wait()
	if rows, count := s.answerRows(t, qid); rows != count {
		t.Fatalf("%d rows but answer_count %d after a removal", rows, count)
	}
}

func TestSolveWriteRateLimitFailsClosed(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-limit-alice", true, 86)
	q := s.ask(t, &alice, "Titre pour la limite")
	if _, err := s.conn.Exec(t.Context(), `DROP TABLE rate_limit_buckets`); err != nil {
		t.Fatal(err)
	}
	rec := s.write(t, &alice, http.MethodPost, "/api/v1/questions/"+q.Id.String()+"/answers", `{"body_md":"x","lang":"fr"}`, nil)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("limiter down: %d %s", rec.Code, rec.Body.String())
	}
	if rows, count := s.answerRows(t, q.Id.String()); rows != 0 || count != 0 {
		t.Fatalf("an answer was written with the limiter down: %d/%d", rows, count)
	}
}

func TestSolveUpdateDrawsOnTheWriteBudget(t *testing.T) {
	awaitFreshMinute(t)
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-patch-limit", true, 87)
	q := s.ask(t, &alice, "Titre pour la limite PATCH")
	path := "/api/v1/questions/" + q.Id.String()
	last := 0
	// One write is the question itself; the 60th and later counted writes are refused with 429.
	for i := 0; i < 61; i++ {
		rec := s.write(t, &alice, http.MethodPatch, path, `{"tags":["a","b","c","d","e"]}`, map[string]string{"If-Match": `"stale"`})
		if last = rec.Code; last == http.StatusTooManyRequests {
			assertProblem(t, rec, http.StatusTooManyRequests, "RATE_LIMITED")
			if rec.Header().Get("Retry-After") == "" {
				t.Fatal("no Retry-After")
			}
			return
		}
		if last != http.StatusPreconditionFailed {
			t.Fatalf("patch %d: %d %s", i, last, rec.Body.String())
		}
	}
	t.Fatalf("61 PATCHes never reached the budget (last %d)", last)
}

func TestSolveEditKeepsAHiddenLinkHidden(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "solve-hidden-alice", true, 88)
	q := s.ask(t, &alice, "Titre pour le lien caché", `"links":[{"type":"resource","id":"`+catAlpha+`"}]`)
	qid := q.Id.String()
	if _, err := s.conn.Exec(t.Context(), `UPDATE content_relations SET status = 'hidden' WHERE src_question_id = $1`, qid); err != nil {
		t.Fatal(err)
	}
	tag := s.read(t, nil, "/api/v1/questions/"+qid).Header().Get("ETag")
	rec := s.write(t, &alice, http.MethodPatch, "/api/v1/questions/"+qid, `{"title":"Nouveau titre","links":[{"type":"resource","id":"`+catAlpha+`"}]}`, map[string]string{"If-Match": tag})
	if rec.Code != http.StatusOK {
		t.Fatalf("patch %d %s", rec.Code, rec.Body.String())
	}
	var status string
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*), min(status) FROM content_relations WHERE src_question_id = $1`, qid).Scan(&n, &status); err != nil {
		t.Fatal(err)
	}
	if n != 1 || status != "hidden" {
		t.Fatalf("%d relations, status %q: the moderator's decision was undone", n, status)
	}
	if got := decodeAs[apigen.Question](t, s.read(t, nil, "/api/v1/questions/"+qid)); len(got.Related) != 0 {
		t.Fatalf("hidden link shown: %+v", got.Related)
	}
}
