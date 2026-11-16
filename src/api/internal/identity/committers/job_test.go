package committers

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

type fakeFetcher struct {
	mu     sync.Mutex
	bodies map[string]string
	err    error
	calls  []string
}

func (f *fakeFetcher) Fetch(_ context.Context, slug string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, slug)
	if f.err != nil {
		return nil, f.err
	}
	b, ok := f.bodies[slug]
	if !ok {
		return nil, errors.New("no body")
	}
	return []byte(b), nil
}

type fakeStore struct {
	mu        sync.Mutex
	generated map[string]time.Time
	rows      map[string][]domain.GitHubDeveloper
	genErr    error
	replErr   error
}

func newFakeStore() *fakeStore {
	return &fakeStore{generated: map[string]time.Time{}, rows: map[string][]domain.GitHubDeveloper{}}
}

func (s *fakeStore) GitHubGeneratedAt(_ context.Context, country string) (time.Time, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.genErr != nil {
		return time.Time{}, false, s.genErr
	}
	g, ok := s.generated[country]
	return g, ok, nil
}

func (s *fakeStore) ReplaceGitHubDevelopers(_ context.Context, country string, devs []domain.GitHubDeveloper) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.replErr != nil {
		return s.replErr
	}
	s.rows[country] = devs
	s.generated[country] = devs[0].GeneratedAt
	return nil
}

func file(generated, login string) string {
	return "users:\n- rank: 1\n  name: \"N\"\n  login: \"" + login + "\"\n  avatarUrl: \"https://avatars.githubusercontent.com/u/1\"\n  contributions: 3\n  organizations: \"\"\ngenerated: " + generated + "\n"
}

func TestCountriesAreTheWestAfricanMVP(t *testing.T) {
	want := "BF:burkina_faso CI:ivory_coast SN:senegal ML:mali NE:niger BJ:benin TG:togo GN:guinea GH:ghana NG:nigeria"
	var got []string
	for _, c := range Countries {
		got = append(got, c.Code+":"+c.Slug)
	}
	if strings.Join(got, " ") != want {
		t.Fatalf("countries %v", got)
	}
}

func TestImportOnceReplacesSkipsUnchangedAndKeepsOnFailure(t *testing.T) {
	f := &fakeFetcher{bodies: map[string]string{
		"burkina_faso": file("2026-11-14T06:00:00Z", "ada"),
		"senegal":      file("2026-11-14T06:00:00Z", "sow"),
		"mali":         "users: [", // a broken file
	}}
	st := newFakeStore()
	countries := []Country{{"BF", "burkina_faso"}, {"SN", "senegal"}, {"ML", "mali"}, {"GH", "ghana"}}
	st.rows["ML"] = []domain.GitHubDeveloper{{Login: "kept"}}
	j := NewJob(f, st, slog.New(slog.DiscardHandler), WithCountries(countries))
	j.ImportOnce(context.Background())
	if len(st.rows["BF"]) != 1 || st.rows["BF"][0].Login != "ada" || st.rows["SN"][0].Country != "SN" {
		t.Fatalf("rows %+v", st.rows)
	}
	if st.rows["ML"][0].Login != "kept" {
		t.Fatal("a broken file keeps the previous rows")
	}
	if _, ok := st.rows["GH"]; ok {
		t.Fatal("a failed fetch writes nothing")
	}
	// Same generated date: skipped, even when the body changed.
	f.bodies["burkina_faso"] = file("2026-11-14T06:00:00Z", "other")
	j.ImportOnce(context.Background())
	if st.rows["BF"][0].Login != "ada" {
		t.Fatal("an unchanged generated date is skipped")
	}
	// A new date replaces.
	f.bodies["burkina_faso"] = file("2026-11-15T06:00:00Z", "newer")
	j.ImportOnce(context.Background())
	if st.rows["BF"][0].Login != "newer" {
		t.Fatal("a new generated date replaces")
	}
	// Store failures are logged, never fatal.
	st.genErr = errors.New("db down")
	j.ImportOnce(context.Background())
	st.genErr, st.replErr = nil, errors.New("tx failed")
	f.bodies["burkina_faso"] = file("2026-11-16T06:00:00Z", "lost")
	j.ImportOnce(context.Background())
	if st.rows["BF"][0].Login != "newer" {
		t.Fatal("a failed replace keeps the previous rows")
	}
}

func TestImportOnceStopsWhenCancelled(t *testing.T) {
	f := &fakeFetcher{err: errors.New("offline")}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	NewJob(f, newFakeStore(), nil).ImportOnce(ctx)
	if len(f.calls) != 0 {
		t.Fatalf("calls after cancel: %v", f.calls)
	}
}

func TestRunImportsAfterTheDelayThenOnEveryTickAndStops(t *testing.T) {
	f := &fakeFetcher{err: errors.New("offline")}
	j := NewJob(f, newFakeStore(), nil, WithCountries([]Country{{"BF", "burkina_faso"}}),
		WithSchedule(time.Millisecond, 2*time.Millisecond))
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { j.Run(ctx); close(done) }()
	deadline := time.After(5 * time.Second)
	for {
		f.mu.Lock()
		n := len(f.calls)
		f.mu.Unlock()
		if n >= 3 {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("only %d runs", n)
		case <-time.After(time.Millisecond):
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop on cancel")
	}
}

func TestRunStopsDuringTheStartDelay(t *testing.T) {
	f := &fakeFetcher{}
	j := NewJob(f, newFakeStore(), nil, WithSchedule(time.Hour, time.Hour))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	j.Run(ctx)
	if len(f.calls) != 0 {
		t.Fatal("no import before the delay")
	}
}

func TestStartStopWaits(t *testing.T) {
	f := &fakeFetcher{err: errors.New("offline")}
	stop := Start(context.Background(), NewJob(f, newFakeStore(), nil, WithSchedule(time.Hour, time.Hour)))
	stop() // returns only once Run has returned
}
