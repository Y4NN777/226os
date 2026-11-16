package committers

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/identity/domain"
)

const (
	// DefaultEvery is the refresh period; committers.top regenerates its files about once a day.
	DefaultEvery = 24 * time.Hour
	// DefaultDelay lets the API finish starting before the first import.
	DefaultDelay = time.Minute
)

// Country is one ISO code and its committers.top location slug.
type Country struct{ Code, Slug string }

// Countries is the West African MVP set (ADR-0016 point 5).
var Countries = []Country{
	{"BF", "burkina_faso"}, {"CI", "ivory_coast"}, {"SN", "senegal"}, {"ML", "mali"}, {"NE", "niger"},
	{"BJ", "benin"}, {"TG", "togo"}, {"GN", "guinea"}, {"GH", "ghana"}, {"NG", "nigeria"},
}

// Fetcher reads one location file.
type Fetcher interface {
	Fetch(ctx context.Context, slug string) ([]byte, error)
}

// Store keeps the snapshot; *identity/repository.Repository implements it.
type Store interface {
	GitHubGeneratedAt(ctx context.Context, country string) (time.Time, bool, error)
	ReplaceGitHubDevelopers(ctx context.Context, country string, devs []domain.GitHubDeveloper) error
}

// Job imports every country on a schedule.
type Job struct {
	fetch     Fetcher
	store     Store
	logger    *slog.Logger
	countries []Country
	delay     time.Duration
	every     time.Duration
}

// Option tunes a Job.
type Option func(*Job)

// WithCountries replaces the country list (tests).
func WithCountries(c []Country) Option { return func(j *Job) { j.countries = c } }

// WithSchedule sets the delay before the first import and the period after it (tests).
func WithSchedule(delay, every time.Duration) Option {
	return func(j *Job) { j.delay, j.every = delay, every }
}

// NewJob builds the job; a nil logger discards logs.
func NewJob(f Fetcher, s Store, logger *slog.Logger, opts ...Option) *Job {
	if logger == nil {
		logger = slog.New(slog.DiscardHandler)
	}
	j := &Job{fetch: f, store: s, logger: logger, countries: Countries, delay: DefaultDelay, every: DefaultEvery}
	for _, o := range opts {
		o(j)
	}
	return j
}

// Start runs the job until the returned function is called, which cancels it and waits for it.
func Start(ctx context.Context, j *Job) (stop func()) {
	ctx, cancel := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Go(func() { j.Run(ctx) })
	return func() {
		cancel()
		wg.Wait()
	}
}

// Run imports once after the start delay, then every period, until ctx is cancelled.
func (j *Job) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(j.delay):
	}
	ticker := time.NewTicker(j.every)
	defer ticker.Stop()
	for ctx.Err() == nil {
		j.ImportOnce(ctx)
		select {
		case <-ctx.Done():
		case <-ticker.C:
		}
	}
}

// ImportOnce imports each country in turn. A failure is logged and leaves that country's rows as they were.
func (j *Job) ImportOnce(ctx context.Context) {
	for _, c := range j.countries {
		if ctx.Err() != nil {
			return
		}
		j.importCountry(ctx, c)
	}
}

func (j *Job) importCountry(ctx context.Context, c Country) {
	body, err := j.fetch.Fetch(ctx, c.Slug)
	if err != nil {
		j.logger.WarnContext(ctx, "committers import: fetch failed", slog.String("country", c.Code), slog.Any("error", err))
		return
	}
	snap, err := Parse(body, c.Code)
	if err != nil {
		j.logger.WarnContext(ctx, "committers import: invalid file", slog.String("country", c.Code), slog.Any("error", err))
		return
	}
	prev, ok, err := j.store.GitHubGeneratedAt(ctx, c.Code)
	if err != nil {
		j.logger.WarnContext(ctx, "committers import: read failed", slog.String("country", c.Code), slog.Any("error", err))
		return
	}
	if ok && prev.Equal(snap.Generated) {
		j.logger.DebugContext(ctx, "committers import: unchanged", slog.String("country", c.Code))
		return
	}
	if err := j.store.ReplaceGitHubDevelopers(ctx, c.Code, snap.Developers); err != nil {
		j.logger.WarnContext(ctx, "committers import: replace failed", slog.String("country", c.Code), slog.Any("error", err))
		return
	}
	j.logger.InfoContext(ctx, "committers import", slog.String("country", c.Code), slog.Int("developers", len(snap.Developers)))
}
