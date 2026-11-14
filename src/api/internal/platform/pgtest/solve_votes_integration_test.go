//go:build integration

package pgtest

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/Y4NN777/226os/src/api/internal/platform/apigen"
)

// US-SOLVE-001 votes and the accepted answer, on the production router over a migrated database: one vote
// per target, a server-side score that always equals the sum of the votes, own content refused, one
// accepted answer under concurrency, and the answer cursor under moving scores.

const unknownID = "019a0000-0000-7000-8000-00000000ffff"

func (s *s2Stack) vote(t *testing.T, m *member, targetType, id string, value int, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodPost, "/api/v1/votes",
		fmt.Sprintf(`{"target_type":%q,"target_id":%q,"value":%d}`, targetType, id, value), headers)
}

func (s *s2Stack) unvote(t *testing.T, m *member, targetType, id string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodDelete, "/api/v1/votes?target_type="+targetType+"&target_id="+id, "", headers)
}

func (s *s2Stack) accept(t *testing.T, m *member, question, answer string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodPost, "/api/v1/questions/"+question+"/accept", `{"answer_id":"`+answer+`"}`, headers)
}

func (s *s2Stack) unaccept(t *testing.T, m *member, question string, headers map[string]string) *httptest.ResponseRecorder {
	t.Helper()
	return s.write(t, m, http.MethodPost, "/api/v1/questions/"+question+"/unaccept", "", headers)
}

func (s *s2Stack) score(t *testing.T, table, id string) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT vote_score FROM `+table+` WHERE id = $1`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) voteRows(t *testing.T, user, target string) (rows int, sum int) {
	t.Helper()
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*), COALESCE(sum(value), 0) FROM votes WHERE user_id = $1 AND target_id = $2`, user, target).Scan(&rows, &sum); err != nil {
		t.Fatal(err)
	}
	return rows, sum
}

func (s *s2Stack) totalVotes(t *testing.T) int {
	t.Helper()
	var n int
	if err := s.conn.QueryRow(t.Context(), `SELECT count(*) FROM votes`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// unreconciled counts the questions and answers whose stored vote_score is not the sum of their votes.
func (s *s2Stack) unreconciled(t *testing.T) int {
	t.Helper()
	var n int
	err := s.conn.QueryRow(t.Context(), `
	  SELECT (SELECT count(*) FROM questions q
	            WHERE q.vote_score <> COALESCE((SELECT sum(v.value) FROM votes v WHERE v.question_id = q.id), 0))
	       + (SELECT count(*) FROM answers a
	            WHERE a.vote_score <> COALESCE((SELECT sum(v.value) FROM votes v WHERE v.answer_id = a.id), 0))`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func (s *s2Stack) acceptedAnswer(t *testing.T, question string) string {
	t.Helper()
	var id *string
	if err := s.conn.QueryRow(t.Context(), `SELECT accepted_answer_id::text FROM questions WHERE id = $1`, question).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if id == nil {
		return ""
	}
	return *id
}

// answersOf lists a question's answers in the contract order, one page of 50.
func (s *s2Stack) answersOf(t *testing.T, question string) []apigen.Answer {
	t.Helper()
	rec := s.read(t, nil, "/api/v1/questions/"+question+"/answers?limit=50")
	if rec.Code != http.StatusOK {
		t.Fatalf("answers %d %s", rec.Code, rec.Body.String())
	}
	return decodeAs[apigen.AnswerPage](t, rec).Items
}

func TestSolveVoteAndAcceptAuthorizationMatrix(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "vote-authz-alice", true, 101)
	bob := s.signIn(t, "vote-authz-bob", true, 102)
	unverified := s.signIn(t, "vote-authz-unverified", false, 103)
	q := s.ask(t, &alice, "Matrice d'autorisation des votes")
	qid := q.Id.String()
	a := s.answer(t, &bob, qid, "Réponse de Bob")
	aid := a.Id.String()
	wrong := map[string]string{"X-CSRF-Token": "wrong"}

	// Anonymous 401, unverified 403 EMAIL_NOT_VERIFIED, wrong CSRF 403 CSRF_FAILED, on every route.
	routes := map[string]func(m *member, h map[string]string) *httptest.ResponseRecorder{
		"castVote": func(m *member, h map[string]string) *httptest.ResponseRecorder {
			return s.vote(t, m, "answer", aid, 1, h)
		},
		"retractVote": func(m *member, h map[string]string) *httptest.ResponseRecorder {
			return s.unvote(t, m, "answer", aid, h)
		},
		"acceptAnswer":   func(m *member, h map[string]string) *httptest.ResponseRecorder { return s.accept(t, m, qid, aid, h) },
		"unacceptAnswer": func(m *member, h map[string]string) *httptest.ResponseRecorder { return s.unaccept(t, m, qid, h) },
	}
	for name, call := range routes {
		assertProblem(t, call(nil, nil), http.StatusUnauthorized, "UNAUTHENTICATED")
		assertProblem(t, call(&unverified, nil), http.StatusForbidden, "EMAIL_NOT_VERIFIED")
		assertProblem(t, call(&alice, wrong), http.StatusForbidden, "CSRF_FAILED")
		if s.totalVotes(t) != 0 || s.acceptedAnswer(t, qid) != "" || s.score(t, "answers", aid) != 0 {
			t.Fatalf("%s: a denied request changed the data", name)
		}
	}

	// An unknown target and a removed one are the same 404, whoever asks, on every route that names a target.
	if _, err := s.conn.Exec(t.Context(), `UPDATE questions SET status = 'removed', removed_at = now() WHERE id = $1`, catGoneQ); err != nil {
		t.Fatal(err)
	}
	removedAnswer := s.answer(t, &bob, qid, "Réponse retirée").Id.String()
	if _, err := s.conn.Exec(t.Context(), `UPDATE answers SET status = 'removed', removed_at = now() WHERE id = $1`, removedAnswer); err != nil {
		t.Fatal(err)
	}
	pairs := map[string][2]*httptest.ResponseRecorder{
		"vote on a question":    {s.vote(t, &bob, "question", unknownID, 1, nil), s.vote(t, &bob, "question", catGoneQ, 1, nil)},
		"vote on an answer":     {s.vote(t, &bob, "answer", unknownID, 1, nil), s.vote(t, &bob, "answer", removedAnswer, 1, nil)},
		"retract on a question": {s.unvote(t, &bob, "question", unknownID, nil), s.unvote(t, &bob, "question", catGoneQ, nil)},
		"retract on an answer":  {s.unvote(t, &bob, "answer", unknownID, nil), s.unvote(t, &bob, "answer", removedAnswer, nil)},
		"accept":                {s.accept(t, &alice, unknownID, aid, nil), s.accept(t, &alice, catGoneQ, aid, nil)},
		"unaccept":              {s.unaccept(t, &alice, unknownID, nil), s.unaccept(t, &alice, catGoneQ, nil)},
	}
	for name, pair := range pairs {
		assertProblem(t, pair[0], http.StatusNotFound, "NOT_FOUND")
		assertProblem(t, pair[1], http.StatusNotFound, "NOT_FOUND")
		if problemWithoutInstance(t, pair[0]) != problemWithoutInstance(t, pair[1]) {
			t.Fatalf("%s: unknown and removed differ: %s / %s", name, pair[0].Body.String(), pair[1].Body.String())
		}
	}
	// An answer of a removed question is also not votable.
	assertProblem(t, s.vote(t, &bob, "answer", removedAnswer, 1, nil), http.StatusNotFound, "NOT_FOUND")
	if s.totalVotes(t) != 0 {
		t.Fatal("a vote reached the database through a 404")
	}
}

func TestSolveVoteOnePerTargetAndOwnContentRefused(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "vote-alice", true, 104)
	bob := s.signIn(t, "vote-bob", true, 105)
	carol := s.signIn(t, "vote-carol", true, 106)
	q := s.ask(t, &alice, "Un vote par cible")
	qid := q.Id.String()
	a := s.answer(t, &bob, qid, "Réponse de Bob")
	aid := a.Id.String()

	// Own content is refused on both kinds of target; nothing is written.
	assertProblem(t, s.vote(t, &alice, "question", qid, 1, nil), http.StatusForbidden, "SELF_VOTE_FORBIDDEN")
	assertProblem(t, s.vote(t, &bob, "answer", aid, -1, nil), http.StatusForbidden, "SELF_VOTE_FORBIDDEN")
	if s.totalVotes(t) != 0 || s.score(t, "questions", qid) != 0 || s.score(t, "answers", aid) != 0 {
		t.Fatal("a refused self-vote changed the data")
	}

	// Cast, repeat (idempotent), change, second voter, retract, retract again.
	check := func(rec *httptest.ResponseRecorder, wantValue, wantScore int) {
		t.Helper()
		if rec.Code != http.StatusOK {
			t.Fatalf("status %d %s", rec.Code, rec.Body.String())
		}
		st := decodeAs[apigen.VoteState](t, rec)
		if int(st.Value) != wantValue || st.Score != wantScore || st.TargetId.String() != aid || st.TargetType != "answer" {
			t.Fatalf("state %+v, want value %d score %d", st, wantValue, wantScore)
		}
	}
	check(s.vote(t, &carol, "answer", aid, 1, nil), 1, 1)
	check(s.vote(t, &carol, "answer", aid, 1, nil), 1, 1)
	if rows, sum := s.voteRows(t, carol.id, aid); rows != 1 || sum != 1 {
		t.Fatalf("%d rows, sum %d after a repeated vote", rows, sum)
	}
	check(s.vote(t, &carol, "answer", aid, -1, nil), -1, -1)
	check(s.vote(t, &alice, "answer", aid, 1, nil), 1, 0)
	if got := s.score(t, "answers", aid); got != 0 || s.unreconciled(t) != 0 {
		t.Fatalf("score %d, unreconciled %d", got, s.unreconciled(t))
	}
	check(s.unvote(t, &carol, "answer", aid, nil), 0, 1)
	check(s.unvote(t, &carol, "answer", aid, nil), 0, 1) // retracting a missing vote succeeds
	if rows, _ := s.voteRows(t, carol.id, aid); rows != 0 {
		t.Fatalf("%d rows after retract", rows)
	}
	// Retracting on your own content is the same success: you cannot have a vote there.
	if rec := s.unvote(t, &bob, "answer", aid, nil); rec.Code != http.StatusOK {
		t.Fatalf("own retract %d %s", rec.Code, rec.Body.String())
	}
	// The question shows the score the votes made; a vote does not touch updated_at (the ETag).
	before := s.read(t, nil, "/api/v1/questions/"+qid).Header().Get("ETag")
	if rec := s.vote(t, &bob, "question", qid, 1, nil); rec.Code != http.StatusOK {
		t.Fatalf("question vote %d %s", rec.Code, rec.Body.String())
	}
	got := s.read(t, nil, "/api/v1/questions/"+qid)
	if q := decodeAs[apigen.Question](t, got); q.VoteScore != 1 || got.Header().Get("ETag") != before {
		t.Fatalf("question %+v, etag %q -> %q", q, before, got.Header().Get("ETag"))
	}
	// Bad bodies are refused before any write.
	for name, body := range map[string]string{
		"value 0":         `{"target_type":"answer","target_id":"` + aid + `","value":0}`,
		"value 2":         `{"target_type":"answer","target_id":"` + aid + `","value":2}`,
		"unknown type":    `{"target_type":"resource","target_id":"` + aid + `","value":1}`,
		"unknown field":   `{"target_type":"answer","target_id":"` + aid + `","value":1,"score":99}`,
		"missing value":   `{"target_type":"answer","target_id":"` + aid + `"}`,
		"bad target id":   `{"target_type":"answer","target_id":"nope","value":1}`,
		"owner in a body": `{"target_type":"answer","target_id":"` + aid + `","value":1,"user_id":"` + alice.id + `"}`,
	} {
		rec := s.write(t, &carol, http.MethodPost, "/api/v1/votes", body, nil)
		if rec.Code != http.StatusUnprocessableEntity && rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d %s", name, rec.Code, rec.Body.String())
		}
	}
	if rows, _ := s.voteRows(t, carol.id, aid); rows != 0 || s.unreconciled(t) != 0 {
		t.Fatalf("a bad body wrote a vote: %d rows", rows)
	}
}

func TestSolveConcurrentDuplicateVotesCreateOneRow(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "vote-race-alice", true, 107)
	bob := s.signIn(t, "vote-race-bob", true, 108)
	q := s.ask(t, &alice, "Votes concurrents")
	aid := s.answer(t, &alice, q.Id.String(), "Réponse").Id.String()

	// 20 identical votes of one user: one row, one point.
	var wg sync.WaitGroup
	codes := make([]int, 20)
	for i := range codes {
		wg.Go(func() { codes[i] = s.vote(t, &bob, "answer", aid, 1, nil).Code })
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusOK {
			t.Fatalf("vote %d: %d", i, c)
		}
	}
	if rows, sum := s.voteRows(t, bob.id, aid); rows != 1 || sum != 1 || s.score(t, "answers", aid) != 1 {
		t.Fatalf("%d rows, sum %d, score %d after 20 identical votes", rows, sum, s.score(t, "answers", aid))
	}
	// Up, down and retract racing: whatever order wins, the score is the one row's value or 0.
	for i := range codes {
		wg.Go(func() {
			switch i % 3 {
			case 0:
				codes[i] = s.vote(t, &bob, "answer", aid, 1, nil).Code
			case 1:
				codes[i] = s.vote(t, &bob, "answer", aid, -1, nil).Code
			default:
				codes[i] = s.unvote(t, &bob, "answer", aid, nil).Code
			}
		})
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusOK {
			t.Fatalf("racing vote %d: %d", i, c)
		}
	}
	if rows, sum := s.voteRows(t, bob.id, aid); rows > 1 || sum != s.score(t, "answers", aid) {
		t.Fatalf("%d rows, sum %d, score %d after racing votes", rows, sum, s.score(t, "answers", aid))
	}
	if n := s.unreconciled(t); n != 0 {
		t.Fatalf("%d unreconciled targets", n)
	}
}

// After a seeded random sequence of casts, changes and retractions by several users, on several targets,
// the stored score equals the sum of the votes for every question and answer.
func TestSolveVoteScoreReconcilesWithTheVotes(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "vote-rec-alice", true, 109)
	voters := []member{
		s.signIn(t, "vote-rec-bob", true, 110),
		s.signIn(t, "vote-rec-carol", true, 111),
		s.signIn(t, "vote-rec-dave", true, 112),
		s.signIn(t, "vote-rec-erin", true, 113),
	}
	q := s.ask(t, &alice, "Réconciliation des votes")
	type target struct{ kind, id string }
	targets := []target{{"question", q.Id.String()}}
	for i := range 3 {
		targets = append(targets, target{"answer", s.answer(t, &voters[i], q.Id.String(), fmt.Sprintf("Réponse %d", i)).Id.String()})
	}
	everyone := append([]*member{&alice}, func() []*member {
		var out []*member
		for i := range voters {
			out = append(out, &voters[i])
		}
		return out
	}()...)

	// plan draws one operation now (the generator is not shared across goroutines) and runs it later.
	rng := rand.New(rand.NewPCG(2026, 11))
	plan := func() func() {
		who, tg, op := everyone[rng.IntN(len(everyone))], targets[rng.IntN(len(targets))], rng.IntN(3)
		return func() {
			var rec *httptest.ResponseRecorder
			switch op {
			case 0:
				rec = s.vote(t, who, tg.kind, tg.id, 1, nil)
			case 1:
				rec = s.vote(t, who, tg.kind, tg.id, -1, nil)
			default:
				rec = s.unvote(t, who, tg.kind, tg.id, nil)
			}
			if rec.Code != http.StatusOK && rec.Code != http.StatusForbidden {
				t.Errorf("vote: %d %s", rec.Code, rec.Body.String())
			}
		}
	}
	// 30 sequential operations, then a concurrent burst, each user staying under the write budget.
	for range 30 {
		plan()()
	}
	var wg sync.WaitGroup
	for range 16 {
		wg.Go(plan())
	}
	wg.Wait()
	if s.totalVotes(t) == 0 {
		t.Fatal("the sequence left no vote: the check would prove nothing")
	}
	if n := s.unreconciled(t); n != 0 {
		t.Fatalf("%d targets whose vote_score is not the sum of their votes", n)
	}
	// And the API shows the same numbers.
	for _, a := range s.answersOf(t, q.Id.String()) {
		if got := s.score(t, "answers", a.Id.String()); a.VoteScore != got {
			t.Fatalf("answer %s: API %d, row %d", a.Id, a.VoteScore, got)
		}
	}
}

func TestSolveAcceptAnswerRules(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "accept-alice", true, 114)
	bob := s.signIn(t, "accept-bob", true, 115)
	q := s.ask(t, &alice, "Accepter une réponse")
	qid := q.Id.String()
	other := s.ask(t, &alice, "Une autre question")
	a1 := s.answer(t, &bob, qid, "Première").Id.String()
	a2 := s.answer(t, &bob, qid, "Seconde").Id.String()
	foreign := s.answer(t, &bob, other.Id.String(), "D'une autre question").Id.String()
	removed := s.answer(t, &bob, qid, "Retirée").Id.String()
	if _, err := s.conn.Exec(t.Context(), `UPDATE answers SET status = 'removed', removed_at = now() WHERE id = $1`, removed); err != nil {
		t.Fatal(err)
	}
	etag := s.read(t, nil, "/api/v1/questions/"+qid).Header().Get("ETag")

	// Bob, a third party who even wrote the answers, gets 403 and nothing changes.
	assertProblem(t, s.accept(t, &bob, qid, a1, nil), http.StatusForbidden, "FORBIDDEN")
	assertProblem(t, s.unaccept(t, &bob, qid, nil), http.StatusForbidden, "FORBIDDEN")
	if s.acceptedAnswer(t, qid) != "" {
		t.Fatal("Bob changed the accepted answer")
	}
	// The author cannot accept an answer of another question, an unknown one, or a removed one: the same 422.
	var bodies []string
	for _, bad := range []string{foreign, unknownID, removed} {
		rec := s.accept(t, &alice, qid, bad, nil)
		assertProblem(t, rec, http.StatusUnprocessableEntity, "VALIDATION_ERROR")
		bodies = append(bodies, problemWithoutInstance(t, rec))
	}
	if bodies[0] != bodies[1] || bodies[1] != bodies[2] {
		t.Fatalf("the refusals differ, so an answer id would tell about other questions: %v", bodies)
	}
	if s.acceptedAnswer(t, qid) != "" || s.acceptedAnswer(t, other.Id.String()) != "" {
		t.Fatal("a refused accept changed the data")
	}

	// Accept, accept again (idempotent), replace atomically, answers list shows the new one first.
	for _, want := range []string{a1, a1, a2} {
		rec := s.accept(t, &alice, qid, want, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("accept %d %s", rec.Code, rec.Body.String())
		}
		if got := decodeAs[apigen.Question](t, rec); got.AcceptedAnswerId == nil || got.AcceptedAnswerId.String() != want || got.State != "resolved" {
			t.Fatalf("question %+v, want accepted %s", got, want)
		}
		if s.acceptedAnswer(t, qid) != want {
			t.Fatalf("accepted %s, want %s", s.acceptedAnswer(t, qid), want)
		}
	}
	answers := s.answersOf(t, qid)
	accepted := 0
	for _, a := range answers {
		if a.IsAccepted {
			accepted++
		}
	}
	if len(answers) < 2 || answers[0].Id.String() != a2 || !answers[0].IsAccepted || accepted != 1 {
		t.Fatalf("answers %+v: want %s first, one accepted", answers, a2)
	}
	if s.read(t, nil, "/api/v1/questions/"+qid).Header().Get("ETag") != etag {
		t.Fatal("accepting changed the ETag")
	}
	// Unaccept clears it, and clearing again is a success.
	if rec := s.unaccept(t, &alice, qid, nil); rec.Code != http.StatusOK || decodeAs[apigen.Question](t, rec).AcceptedAnswerId != nil || s.acceptedAnswer(t, qid) != "" {
		t.Fatalf("unaccept %d %s", rec.Code, rec.Body.String())
	}
	if rec := s.unaccept(t, &alice, qid, nil); rec.Code != http.StatusOK { // nothing accepted: success
		t.Fatalf("second unaccept %d %s", rec.Code, rec.Body.String())
	}
	// Unaccept on an answer-less question is a success too, and other questions stay untouched.
	if rec := s.unaccept(t, &alice, other.Id.String(), nil); rec.Code != http.StatusOK {
		t.Fatalf("unaccept without answers %d", rec.Code)
	}
}

func TestSolveConcurrentAcceptsLeaveExactlyOneAcceptedAnswer(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "accept-race-alice", true, 116)
	bob := s.signIn(t, "accept-race-bob", true, 117)
	q := s.ask(t, &alice, "Acceptations concurrentes")
	qid := q.Id.String()
	var ids []string
	for i := range 8 {
		ids = append(ids, s.answer(t, &bob, qid, fmt.Sprintf("Réponse %d", i)).Id.String())
	}
	// The author accepts eight different answers at once while voters vote on all of them: no error leaks,
	// no deadlock, and the question ends with exactly one accepted answer, one of its own.
	var wg sync.WaitGroup
	codes := make([]int, len(ids))
	for i, id := range ids {
		wg.Go(func() { codes[i] = s.accept(t, &alice, qid, id, nil).Code })
	}
	votes := make([]int, len(ids))
	for i, id := range ids {
		wg.Go(func() { votes[i] = s.vote(t, &alice, "answer", id, 1, nil).Code })
	}
	wg.Wait()
	for i := range ids {
		if codes[i] != http.StatusOK || votes[i] != http.StatusOK {
			t.Fatalf("accept %d: %d, vote %d: %d", i, codes[i], i, votes[i])
		}
	}
	final := s.acceptedAnswer(t, qid)
	if !slices.Contains(ids, final) {
		t.Fatalf("accepted answer %q is none of the question's answers", final)
	}
	flagged := 0
	for _, a := range s.answersOf(t, qid) {
		if a.IsAccepted {
			flagged++
			if a.Id.String() != final {
				t.Fatalf("API accepts %s, row says %s", a.Id, final)
			}
		}
	}
	if flagged != 1 {
		t.Fatalf("%d answers flagged accepted", flagged)
	}
	if n := s.unreconciled(t); n != 0 {
		t.Fatalf("%d unreconciled targets", n)
	}
	// Accepts racing unaccepts: any outcome is fine as long as it is one coherent state and no call fails.
	for i := range 8 {
		wg.Go(func() {
			if i%2 == 0 {
				codes[i] = s.accept(t, &alice, qid, ids[i], nil).Code
			} else {
				codes[i] = s.unaccept(t, &alice, qid, nil).Code
			}
		})
	}
	wg.Wait()
	for i, c := range codes {
		if c != http.StatusOK {
			t.Fatalf("racing call %d: %d", i, c)
		}
	}
	if final := s.acceptedAnswer(t, qid); final != "" && !slices.Contains(ids, final) {
		t.Fatalf("accepted answer %q", final)
	}
}

// The answer cursor is a keyset on the live (accepted, score, age, id) order. A vote between two pages can
// move an answer across the cut: it may then show twice or be missed by that walk, never an error. Answers
// whose position did not change are returned exactly once.
func TestSolveAnswerPagingUnderMovingScores(t *testing.T) {
	s := newSolveStack(t)
	alice := s.signIn(t, "page-alice", true, 118)
	bob := s.signIn(t, "page-bob", true, 119)
	carol := s.signIn(t, "page-carol", true, 120)
	erin := s.signIn(t, "page-erin", true, 121)
	q := s.ask(t, &alice, "Pagination sous les votes")
	qid := q.Id.String()
	var ids []string
	for i := range 4 {
		ids = append(ids, s.answer(t, &bob, qid, fmt.Sprintf("Réponse %d", i)).Id.String())
	}
	page := func(cursor string) ([]string, string) {
		path := "/api/v1/questions/" + qid + "/answers?limit=2"
		if cursor != "" {
			path += "&cursor=" + cursor
		}
		rec := s.read(t, nil, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("page %d %s", rec.Code, rec.Body.String())
		}
		p := decodeAs[apigen.AnswerPage](t, rec)
		var out []string
		for _, a := range p.Items {
			out = append(out, a.Id.String())
		}
		next := ""
		if p.NextCursor != nil {
			next = *p.NextCursor
		}
		return out, next
	}
	walk := func(first []string, cursor string) []string {
		seen := append([]string(nil), first...)
		for range 10 {
			if cursor == "" {
				return seen
			}
			var got []string
			got, cursor = page(cursor)
			seen = append(seen, got...)
		}
		t.Fatal("the walk does not end")
		return nil
	}
	count := func(list []string, id string) int { return strings.Count(strings.Join(list, ","), id) }

	// Carol's up vote puts ids[0] first: page 1 is [0, 1]. She then flips it to down, so ids[0] falls
	// below the cut: the walk goes on without an error and shows ids[0] again at the end.
	if rec := s.vote(t, &carol, "answer", ids[0], 1, nil); rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	first, cursor := page("")
	if strings.Join(first, ",") != ids[0]+","+ids[1] || cursor == "" {
		t.Fatalf("first page %v", first)
	}
	if rec := s.vote(t, &carol, "answer", ids[0], -1, nil); rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	seen := walk(first, cursor)
	if count(seen, ids[0]) != 2 || count(seen, ids[1]) != 1 || count(seen, ids[2]) != 1 || count(seen, ids[3]) != 1 {
		t.Fatalf("walk %v: want ids[0] twice (it fell below the cut), the others once", seen)
	}

	// Erin's up vote lifts ids[3] above the cut after page 1: this walk misses it, the next one shows it first.
	if rec := s.vote(t, &carol, "answer", ids[0], 1, nil); rec.Code != http.StatusOK { // back to +1
		t.Fatal(rec.Body.String())
	}
	first, cursor = page("")
	if strings.Join(first, ",") != ids[0]+","+ids[1] {
		t.Fatalf("first page %v", first)
	}
	if rec := s.vote(t, &erin, "answer", ids[3], 1, nil); rec.Code != http.StatusOK {
		t.Fatal(rec.Body.String())
	}
	if rec := s.vote(t, &bob, "answer", ids[3], 1, nil); rec.Code != http.StatusForbidden { // Bob wrote it
		t.Fatalf("self vote %d", rec.Code)
	}
	seen = walk(first, cursor)
	if count(seen, ids[3]) != 0 || count(seen, ids[2]) != 1 {
		t.Fatalf("walk %v: want ids[3] missed by this walk and ids[2] once", seen)
	}
	if fresh, _ := page(""); fresh[0] != ids[0] && fresh[0] != ids[3] {
		t.Fatalf("a fresh walk starts with %v", fresh)
	}

	// Votes landing while a client pages: every call stays 200, and the answers nobody voted on come back
	// exactly once each.
	q2 := s.ask(t, &alice, "Pagination pendant une rafale de votes")
	q2id := q2.Id.String()
	var all []string
	for i := range 12 {
		all = append(all, s.answer(t, &bob, q2id, fmt.Sprintf("Réponse %d", i)).Id.String())
	}
	hot := all[:4]
	var wg sync.WaitGroup
	stop := make(chan struct{})
	for _, voter := range []*member{&carol, &erin, &alice} {
		wg.Go(func() {
			rng := rand.New(rand.NewPCG(7, uint64(len(voter.id))))
			for n := 0; n < 12; n++ {
				select {
				case <-stop:
					return
				default:
				}
				h := hot[rng.IntN(len(hot))]
				switch rng.IntN(3) {
				case 0:
					s.vote(t, voter, "answer", h, 1, nil)
				case 1:
					s.vote(t, voter, "answer", h, -1, nil)
				default:
					s.unvote(t, voter, "answer", h, nil)
				}
			}
		})
	}
	var walked []string
	cursor = ""
	for range 20 {
		path := "/api/v1/questions/" + q2id + "/answers?limit=3"
		if cursor != "" {
			path += "&cursor=" + cursor
		}
		rec := s.read(t, nil, path)
		if rec.Code != http.StatusOK {
			t.Fatalf("page during the storm: %d %s", rec.Code, rec.Body.String())
		}
		p := decodeAs[apigen.AnswerPage](t, rec)
		for _, a := range p.Items {
			walked = append(walked, a.Id.String())
		}
		if p.NextCursor == nil {
			break
		}
		cursor = *p.NextCursor
	}
	close(stop)
	wg.Wait()
	for _, id := range all[4:] {
		if count(walked, id) != 1 {
			t.Fatalf("answer %s seen %d times in %v", id, count(walked, id), walked)
		}
	}
	if n := s.unreconciled(t); n != 0 {
		t.Fatalf("%d unreconciled targets", n)
	}
}

func TestSolveVotesAndAcceptDrawOnTheWriteBudget(t *testing.T) {
	awaitFreshMinute(t)
	s := newSolveStack(t)
	alice := s.signIn(t, "vote-budget-alice", true, 122)
	bob := s.signIn(t, "vote-budget-bob", true, 123)
	q := s.ask(t, &alice, "Budget des écritures")
	s.answer(t, &bob, q.Id.String(), "Réponse")
	// Bob's answer is one counted write on his budget of 60 per minute, so the 60th vote is the 61st write.
	for i := 0; i < 61; i++ {
		rec := s.vote(t, &bob, "question", q.Id.String(), 1-2*(i%2), nil)
		if rec.Code == http.StatusTooManyRequests {
			assertProblem(t, rec, http.StatusTooManyRequests, "RATE_LIMITED")
			if rec.Header().Get("Retry-After") == "" {
				t.Fatal("no Retry-After")
			}
			if n := s.unreconciled(t); n != 0 {
				t.Fatalf("%d unreconciled targets", n)
			}
			return
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("vote %d: %d %s", i, rec.Code, rec.Body.String())
		}
	}
	t.Fatal("61 votes never reached the budget")
}
