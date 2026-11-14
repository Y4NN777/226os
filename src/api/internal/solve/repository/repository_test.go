package repository

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/Y4NN777/226os/src/api/internal/solve/domain"
)

const (
	qID    = "00000000-0000-7000-8000-0000000000d1"
	aID    = "00000000-0000-7000-8000-0000000000a1"
	userID = "00000000-0000-7000-8000-0000000000b1"
	otherU = "00000000-0000-7000-8000-0000000000b2"
)

var errDB = errors.New("db down")

func uid(s string) pgtype.UUID { return pgtype.UUID{Bytes: uuid.MustParse(s), Valid: true} }

func ts() pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: time.Date(2026, 11, 14, 9, 0, 0, 0, time.UTC), Valid: true}
}

// step is one scripted database call: Query/QueryRow return rows, Exec reports affected rows.
type step struct {
	rows     [][]any
	affected int64
	err      error
}

type fakeDB struct {
	t         *testing.T
	steps     []step
	n         int
	beginErr  error
	commitErr error
	committed bool
}

func (f *fakeDB) next(kind string) step {
	f.t.Helper()
	if f.n >= len(f.steps) {
		f.t.Fatalf("unexpected %s (call %d)", kind, f.n+1)
	}
	s := f.steps[f.n]
	f.n++
	return s
}

func (f *fakeDB) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	s := f.next("exec")
	return pgconn.NewCommandTag(fmt.Sprintf("INSERT 0 %d", s.affected)), s.err
}

func (f *fakeDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	s := f.next("query")
	if s.err != nil {
		return nil, s.err
	}
	return &fakeRows{rows: s.rows}, nil
}

func (f *fakeDB) QueryRow(context.Context, string, ...any) pgx.Row {
	s := f.next("query row")
	if s.err != nil {
		return fakeRow{err: s.err}
	}
	if len(s.rows) == 0 {
		return fakeRow{err: pgx.ErrNoRows}
	}
	return fakeRow{vals: s.rows[0]}
}

func (f *fakeDB) Begin(context.Context) (pgx.Tx, error) {
	if f.beginErr != nil {
		return nil, f.beginErr
	}
	return fakeTx{db: f}, nil
}

type fakeTx struct {
	pgx.Tx
	db *fakeDB
}

func (t fakeTx) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	return t.db.Exec(ctx, sql, args...)
}

func (t fakeTx) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	return t.db.Query(ctx, sql, args...)
}

func (t fakeTx) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	return t.db.QueryRow(ctx, sql, args...)
}

func (t fakeTx) Commit(context.Context) error {
	if t.db.commitErr == nil {
		t.db.committed = true
	}
	return t.db.commitErr
}

func (t fakeTx) Rollback(context.Context) error { return nil }

type fakeRow struct {
	vals []any
	err  error
}

func (r fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	return scan(dest, r.vals)
}

type fakeRows struct {
	rows [][]any
	i    int
}

func (r *fakeRows) Close()                                       {}
func (r *fakeRows) Err() error                                   { return nil }
func (r *fakeRows) CommandTag() pgconn.CommandTag                { return pgconn.CommandTag{} }
func (r *fakeRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (r *fakeRows) Values() ([]any, error)                       { return nil, nil }
func (r *fakeRows) RawValues() [][]byte                          { return nil }
func (r *fakeRows) Conn() *pgx.Conn                              { return nil }
func (r *fakeRows) TypeMap() *pgtype.Map                         { return nil }
func (r *fakeRows) Next() bool {
	if r.i < len(r.rows) {
		r.i++
		return true
	}
	return false
}
func (r *fakeRows) Scan(dest ...any) error { return scan(dest, r.rows[r.i-1]) }

func scan(dest, vals []any) error {
	if len(dest) != len(vals) {
		return fmt.Errorf("scan width %d, want %d", len(vals), len(dest))
	}
	for i, d := range dest {
		reflect.ValueOf(d).Elem().Set(reflect.ValueOf(vals[i]))
	}
	return nil
}

func db(t *testing.T, steps ...step) *fakeDB { return &fakeDB{t: t, steps: steps} }

func rows(r ...[]any) step { return step{rows: r} }

func fail(err error) step { return step{err: err} }

// questionRow is a published question: id, title, body, language, snippet, snippet language, country,
// answer_count, vote_score, accepted answer, created, updated, author id, username, display name, author country.
func questionRow(snippet, country string, accepted pgtype.UUID) []any {
	return []any{uid(qID), "Titre", "Corps", "fr", snippet, "go", country, int32(2), int32(3), accepted, ts(), ts(), uid(userID), "alice", "Alice", "BF"}
}

// listRow is questionRow plus the relevance rank the list query adds.
func listRow(snippet, country string, accepted pgtype.UUID, rank float32) []any {
	return append(questionRow(snippet, country, accepted), rank)
}

func lockRow() []any {
	return []any{uid(qID), uid(userID), int32(2), int32(3), pgtype.UUID{}, ts(), ts()}
}

func answerRow(accepted bool) []any {
	return []any{uid(aID), uid(qID), "Réponse", "fr", int32(4), ts(), ts(), accepted, uid(userID), "alice", "Alice", ""}
}

func tagRows() step { return rows([]any{uid(qID), "api"}, []any{uid(qID), "go"}) }
func relatedRows() step {
	return rows([]any{uid(qID), "resource", uid(aID), "Alpha", "fr"})
}
func yes() step { return rows([]any{true}) }
func no() step  { return rows([]any{false}) }

func fkErr(constraint string) error {
	return &pgconn.PgError{Code: "23503", ConstraintName: constraint}
}

func TestListQuestionsAttachesTagsAndLinks(t *testing.T) {
	at := ts().Time
	d := db(t, rows(listRow("code", "BF", uid(aID), 0.25), listRow("", "", pgtype.UUID{}, 0)), tagRows(), relatedRows())
	got, err := New(d).ListQuestions(t.Context(), domain.QuestionQuery{Q: "paiement", Tag: "go", State: "open", Country: "BF", Lang: "fr", AfterCreated: &at, AfterID: qID, Limit: 21})
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	q := got[0]
	if q.Rank != 0.25 || got[1].Rank != 0 {
		t.Fatalf("ranks %v %v", q.Rank, got[1].Rank)
	}
	if q.ID != qID || q.Author.Username != "alice" || q.Author.Country != "BF" || q.Country != "BF" || q.AcceptedAnswerID != aID || q.AnswerCount != 2 ||
		q.VoteScore != 3 || len(q.Snippets) != 1 || q.Snippets[0].Code != "code" || q.Snippets[0].Language != "go" ||
		len(q.Tags) != 2 || q.Tags[0] != "api" || len(q.Related) != 1 || q.Related[0].Title != "Alpha" || q.Related[0].Type != "resource" || q.State() != "resolved" {
		t.Fatalf("question %+v", q)
	}
	if plain := got[1]; plain.AcceptedAnswerID != "" || plain.Snippets == nil || len(plain.Snippets) != 0 || plain.Country != "" {
		t.Fatalf("plain %+v", plain)
	}
}

func TestListQuestionsEmptyAndErrors(t *testing.T) {
	got, err := New(db(t, rows())).ListQuestions(t.Context(), domain.QuestionQuery{Limit: 21})
	if err != nil || len(got) != 0 {
		t.Fatalf("empty %+v, %v", got, err)
	}
	if _, err := New(db(t)).ListQuestions(t.Context(), domain.QuestionQuery{AfterID: "bad"}); err == nil {
		t.Fatal("bad cursor id accepted")
	}
	for name, steps := range map[string][]step{
		"list":    {fail(errDB)},
		"tags":    {rows(listRow("", "", pgtype.UUID{}, 0)), fail(errDB)},
		"related": {rows(listRow("", "", pgtype.UUID{}, 0)), tagRows(), fail(errDB)},
	} {
		if _, err := New(db(t, steps...)).ListQuestions(t.Context(), domain.QuestionQuery{}); !errors.Is(err, errDB) {
			t.Fatalf("%s err %v", name, err)
		}
	}
}

func TestGetQuestion(t *testing.T) {
	got, err := New(db(t, rows(questionRow("code", "", pgtype.UUID{})), tagRows(), rows())).GetQuestion(t.Context(), qID)
	if err != nil || got.ID != qID || len(got.Related) != 0 || got.Related == nil || len(got.Tags) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, err := New(db(t)).GetQuestion(t.Context(), "bad"); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id %v", err)
	}
	if _, err := New(db(t, rows())).GetQuestion(t.Context(), qID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing %v", err)
	}
	for name, steps := range map[string][]step{
		"get":     {fail(errDB)},
		"tags":    {rows(questionRow("", "", pgtype.UUID{})), fail(errDB)},
		"related": {rows(questionRow("", "", pgtype.UUID{})), tagRows(), fail(errDB)},
	} {
		if _, err := New(db(t, steps...)).GetQuestion(t.Context(), qID); !errors.Is(err, errDB) {
			t.Fatalf("%s err %v", name, err)
		}
	}
}

func newQuestion() domain.NewQuestion {
	return domain.NewQuestion{
		AuthorID: userID, Title: "Titre", BodyMD: "Corps", Lang: "fr", Country: "BF", Tags: []string{"go"},
		Snippets: []domain.Snippet{{Language: "go", Code: "x"}},
		Links:    []domain.LinkRef{{Type: "resource", ID: aID}, {Type: "project", ID: aID}, {Type: "article", ID: aID}},
	}
}

func createOK() []step {
	return []step{rows([]any{uid(qID)}), {}, {}, yes(), yes(), yes(), rows(questionRow("x", "BF", pgtype.UUID{})), tagRows(), relatedRows()}
}

func TestCreateQuestionWritesEverythingInOneTransaction(t *testing.T) {
	d := db(t, createOK()...)
	got, err := New(d).CreateQuestion(t.Context(), newQuestion())
	if err != nil || got.ID != qID || !d.committed || len(got.Tags) != 2 {
		t.Fatalf("got %+v, %v, committed %v", got, err, d.committed)
	}
	// No snippet, no country, no link.
	in := newQuestion()
	in.Snippets, in.Country, in.Links = nil, "", nil
	steps := []step{rows([]any{uid(qID)}), {}, {}, rows(questionRow("", "", pgtype.UUID{})), tagRows(), relatedRows()}
	if _, err := New(db(t, steps...)).CreateQuestion(t.Context(), in); err != nil {
		t.Fatal(err)
	}
}

func TestCreateQuestionErrors(t *testing.T) {
	ok := createOK()
	with := func(i int, s step) []step {
		out := append([]step(nil), ok[:i]...)
		return append(out, s)
	}
	cases := map[string]struct {
		steps []step
		want  error
	}{
		"insert":          {with(0, fail(errDB)), errDB},
		"unknown country": {with(0, fail(fkErr("questions_country_fk"))), domain.ErrUnknownCountry},
		"other fk":        {with(0, fail(fkErr("questions_author_fk"))), nil},
		"ensure tags":     {with(1, fail(errDB)), errDB},
		"question tags":   {with(2, fail(errDB)), errDB},
		"link resource":   {with(3, fail(errDB)), errDB},
		"link project":    {with(4, fail(errDB)), errDB},
		"link article":    {with(5, fail(errDB)), errDB},
		"link unusable":   {with(3, no()), domain.ErrLinkTarget},
		"load":            {with(6, fail(errDB)), errDB},
		"load tags":       {with(7, fail(errDB)), errDB},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := db(t, tc.steps...)
			_, err := New(d).CreateQuestion(t.Context(), newQuestion())
			if err == nil || (tc.want != nil && !errors.Is(err, tc.want)) || d.committed {
				t.Fatalf("err %v, committed %v", err, d.committed)
			}
		})
	}
	d := db(t, createOK()...)
	d.commitErr = errDB
	if _, err := New(d).CreateQuestion(t.Context(), newQuestion()); !errors.Is(err, errDB) {
		t.Fatalf("commit err %v", err)
	}
	d = db(t)
	d.beginErr = errDB
	if _, err := New(d).CreateQuestion(t.Context(), newQuestion()); !errors.Is(err, errDB) {
		t.Fatalf("begin err %v", err)
	}
	bad := newQuestion()
	bad.AuthorID = "bad"
	if _, err := New(db(t)).CreateQuestion(t.Context(), bad); err == nil {
		t.Fatal("bad author id accepted")
	}
	for name, links := range map[string][]domain.LinkRef{
		"bad target id": {{Type: "resource", ID: "bad"}},
		"unknown type":  {{Type: "series", ID: aID}},
	} {
		in := newQuestion()
		in.Links = links
		if _, err := New(db(t, rows([]any{uid(qID)}), step{}, step{})).CreateQuestion(t.Context(), in); !errors.Is(err, domain.ErrLinkTarget) {
			t.Fatalf("%s: %v", name, err)
		}
	}
}

func updateTail() []step {
	return []step{rows(questionRow("", "", pgtype.UUID{})), tagRows(), relatedRows()}
}

func guardOK(domain.Question) error { return nil }

func TestUpdateQuestionAppliesAMergePatchUnderTheLock(t *testing.T) {
	title, country := "Nouveau", ""
	tags := []string{"go"}
	links := []domain.LinkRef{{Type: "project", ID: aID}}
	snippets := []domain.Snippet{}
	var seen domain.Question
	// Everything set: update, tags (delete, ensure, insert), links (delete, one link), then the reload.
	steps := append([]step{rows(lockRow()), {}, {}, {}, {}, {}, yes()}, updateTail()...)
	d := db(t, steps...)
	got, err := New(d).UpdateQuestion(t.Context(), qID, domain.QuestionPatch{Title: &title, Country: &country, Tags: &tags, Links: &links, Snippets: &snippets},
		func(q domain.Question) error { seen = q; return nil })
	if err != nil || got.ID != qID || !d.committed {
		t.Fatalf("got %+v, %v, committed %v", got, err, d.committed)
	}
	if seen.ID != qID || seen.Author.ID != userID || seen.AnswerCount != 2 || seen.VoteScore != 3 || seen.AcceptedAnswerID != "" || !seen.UpdatedAt.Equal(ts().Time) {
		t.Fatalf("the guard saw %+v", seen)
	}
	// Only a title: one update, then the reload. A snippet patch with one snippet.
	steps = append([]step{rows(lockRow()), {}}, updateTail()...)
	if _, err := New(db(t, steps...)).UpdateQuestion(t.Context(), qID, domain.QuestionPatch{Title: &title}, guardOK); err != nil {
		t.Fatal(err)
	}
	one := []domain.Snippet{{Language: "go", Code: "x"}}
	country = "GH"
	steps = append([]step{rows(lockRow()), {}}, updateTail()...)
	if _, err := New(db(t, steps...)).UpdateQuestion(t.Context(), qID, domain.QuestionPatch{Snippets: &one, Country: &country}, guardOK); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateQuestionStopsAtTheGuardAndNothingIsWritten(t *testing.T) {
	d := db(t, rows(lockRow()))
	if _, err := New(d).UpdateQuestion(t.Context(), qID, domain.QuestionPatch{}, func(domain.Question) error { return domain.ErrForbidden }); !errors.Is(err, domain.ErrForbidden) || d.committed {
		t.Fatalf("err %v, committed %v", err, d.committed)
	}
}

func TestUpdateQuestionErrors(t *testing.T) {
	title := "Nouveau"
	tags := []string{"go"}
	links := []domain.LinkRef{{Type: "project", ID: aID}}
	full := domain.QuestionPatch{Title: &title, Tags: &tags, Links: &links}
	ok := append([]step{rows(lockRow()), {}, {}, {}, {}, {}, yes()}, updateTail()...)
	with := func(i int, s step) []step { return append(append([]step(nil), ok[:i]...), s) }
	cases := map[string]struct {
		steps []step
		want  error
	}{
		"missing":         {[]step{rows()}, domain.ErrNotFound},
		"lock":            {[]step{fail(errDB)}, errDB},
		"update":          {with(1, fail(errDB)), errDB},
		"unknown country": {with(1, fail(fkErr("questions_country_fk"))), domain.ErrUnknownCountry},
		"delete tags":     {with(2, fail(errDB)), errDB},
		"ensure tags":     {with(3, fail(errDB)), errDB},
		"insert tags":     {with(4, fail(errDB)), errDB},
		"delete links":    {with(5, fail(errDB)), errDB},
		"link unusable":   {with(6, no()), domain.ErrLinkTarget},
		"reload":          {with(7, fail(errDB)), errDB},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := db(t, tc.steps...)
			if _, err := New(d).UpdateQuestion(t.Context(), qID, full, guardOK); !errors.Is(err, tc.want) || d.committed {
				t.Fatalf("err %v, committed %v", err, d.committed)
			}
		})
	}
	d := db(t, ok...)
	d.commitErr = errDB
	if _, err := New(d).UpdateQuestion(t.Context(), qID, full, guardOK); !errors.Is(err, errDB) {
		t.Fatalf("commit err %v", err)
	}
	d = db(t)
	d.beginErr = errDB
	if _, err := New(d).UpdateQuestion(t.Context(), qID, full, guardOK); !errors.Is(err, errDB) {
		t.Fatalf("begin err %v", err)
	}
	if _, err := New(db(t)).UpdateQuestion(t.Context(), "bad", full, guardOK); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id %v", err)
	}
}

func TestListAnswersAttachesSources(t *testing.T) {
	key := &domain.AnswerKey{Rank: 1, NegScore: -2, CreatedAt: ts().Time, ID: aID}
	d := db(t, yes(), rows(answerRow(true), answerRow(false)), rows([]any{uid(aID), "https://a.example"}))
	got, err := New(d).ListAnswers(t.Context(), domain.AnswerQuery{QuestionID: qID, After: key, Limit: 21})
	if err != nil || len(got) != 2 {
		t.Fatalf("got %+v, %v", got, err)
	}
	a := got[0]
	if a.ID != aID || a.QuestionID != qID || !a.Accepted || a.VoteScore != 4 || a.Author.Username != "alice" || a.Snippets == nil || len(a.Snippets) != 0 ||
		len(a.Sources) != 1 || got[1].Sources == nil || len(got[1].Sources) != 1 {
		// both rows share the id, so both get the one source
		t.Fatalf("answer %+v", a)
	}
	// An empty page reads no sources.
	if got, err := New(db(t, yes(), rows())).ListAnswers(t.Context(), domain.AnswerQuery{QuestionID: qID, Limit: 21}); err != nil || len(got) != 0 {
		t.Fatalf("empty %+v, %v", got, err)
	}
}

func TestListAnswersErrors(t *testing.T) {
	q := domain.AnswerQuery{QuestionID: qID, Limit: 21}
	if _, err := New(db(t)).ListAnswers(t.Context(), domain.AnswerQuery{QuestionID: "bad"}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad id %v", err)
	}
	if _, err := New(db(t, no())).ListAnswers(t.Context(), q); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("missing question %v", err)
	}
	for name, steps := range map[string][]step{
		"exists":  {fail(errDB)},
		"list":    {yes(), fail(errDB)},
		"sources": {yes(), rows(answerRow(false)), fail(errDB)},
	} {
		if _, err := New(db(t, steps...)).ListAnswers(t.Context(), q); !errors.Is(err, errDB) {
			t.Fatalf("%s err %v", name, err)
		}
	}
}

func newAnswer() domain.NewAnswer {
	return domain.NewAnswer{QuestionID: qID, AuthorID: userID, BodyMD: "Réponse", Lang: "fr", Sources: []string{"https://a.example"}}
}

func answerOK() []step {
	return []step{rows([]any{uid(qID)}), rows([]any{uid(aID)}), {}, rows(answerRow(false)), rows([]any{uid(aID), "https://a.example"})}
}

func TestCreateAnswerCountsAndInsertsInOneTransaction(t *testing.T) {
	d := db(t, answerOK()...)
	got, err := New(d).CreateAnswer(t.Context(), newAnswer())
	if err != nil || got.ID != aID || len(got.Sources) != 1 || !d.committed {
		t.Fatalf("got %+v, %v, committed %v", got, err, d.committed)
	}
	// No sources: nothing to insert.
	in := newAnswer()
	in.Sources = nil
	steps := []step{rows([]any{uid(qID)}), rows([]any{uid(aID)}), rows(answerRow(false)), rows()}
	if got, err := New(db(t, steps...)).CreateAnswer(t.Context(), in); err != nil || got.Sources == nil || len(got.Sources) != 0 {
		t.Fatalf("got %+v, %v", got, err)
	}
}

func TestCreateAnswerErrors(t *testing.T) {
	ok := answerOK()
	with := func(i int, s step) []step { return append(append([]step(nil), ok[:i]...), s) }
	cases := map[string]struct {
		steps []step
		want  error
	}{
		"question gone": {[]step{rows()}, domain.ErrNotFound},
		"bump":          {with(0, fail(errDB)), errDB},
		"insert":        {with(1, fail(errDB)), errDB},
		"sources":       {with(2, fail(errDB)), errDB},
		"read back":     {with(3, fail(errDB)), errDB},
		"read sources":  {with(4, fail(errDB)), errDB},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			d := db(t, tc.steps...)
			if _, err := New(d).CreateAnswer(t.Context(), newAnswer()); !errors.Is(err, tc.want) || d.committed {
				t.Fatalf("err %v, committed %v", err, d.committed)
			}
		})
	}
	d := db(t, ok...)
	d.commitErr = errDB
	if _, err := New(d).CreateAnswer(t.Context(), newAnswer()); !errors.Is(err, errDB) {
		t.Fatalf("commit err %v", err)
	}
	d = db(t)
	d.beginErr = errDB
	if _, err := New(d).CreateAnswer(t.Context(), newAnswer()); !errors.Is(err, errDB) {
		t.Fatalf("begin err %v", err)
	}
	if _, err := New(db(t)).CreateAnswer(t.Context(), domain.NewAnswer{QuestionID: "bad", AuthorID: userID}); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("bad question id %v", err)
	}
	if _, err := New(db(t)).CreateAnswer(t.Context(), domain.NewAnswer{QuestionID: qID, AuthorID: "bad"}); err == nil {
		t.Fatal("bad author id accepted")
	}
}
