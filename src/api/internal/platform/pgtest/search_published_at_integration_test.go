//go:build integration

package pgtest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/search/domain"
	"github.com/Y4NN777/226os/src/api/internal/search/repository"
)

// The assistant dates its sources from the publication date of each search hit (AI-5). All four
// groups return the row's own published_at, in UTC; the search HTTP response does not carry it.
func TestSearchHitsCarryThePublicationDate(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()
	e := newSearchEnv(ctx, t)
	u, p := sid(sUser), sid(sProvider)
	dates := map[domain.Group]time.Time{
		domain.Questions: time.Date(2026, 3, 4, 10, 15, 0, 0, time.UTC),
		domain.Resources: time.Date(2026, 5, 6, 11, 30, 0, 0, time.UTC),
		domain.Projects:  time.Date(2026, 7, 8, 12, 45, 0, 0, time.UTC),
		domain.Learn:     time.Date(2026, 9, 10, 13, 59, 0, 0, time.UTC),
	}
	ids := map[domain.Group]string{domain.Questions: sid(600), domain.Resources: sid(601), domain.Projects: sid(602), domain.Learn: sid(603)}
	ts := func(g domain.Group) string { return dates[g].Format(time.RFC3339) }
	e.exec(ctx, t,
		fmt.Sprintf(`INSERT INTO questions (id, author_id, title, body_md, language, published_at) VALUES ('%s', '%s', 'Okapi question', 'corps', 'fr', '%s')`,
			ids[domain.Questions], u, ts(domain.Questions)),
		fmt.Sprintf(`INSERT INTO resources (id, provider_id, category_code, type, slug, name, summary, summary_language, created_by, published_at)
		  VALUES ('%s', '%s', 'mobile-money', 'api', 'okapi-api', 'Okapi API', 'resume', 'fr', '%s', '%s')`,
			ids[domain.Resources], p, u, ts(domain.Resources)),
		fmt.Sprintf(`INSERT INTO projects (id, owner_id, name, summary, language, visibility, published_at) VALUES ('%s', '%s', 'Okapi Projet', 'resume', 'fr', 'public', '%s')`,
			ids[domain.Projects], u, ts(domain.Projects)),
		fmt.Sprintf(`INSERT INTO learn_articles (id, author_id, kind, title, summary, body_md, language, level, status, published_at)
		  VALUES ('%s', '%s', 'article', 'Okapi Guide', 'resume', 'corps', 'fr', 'beginner', 'published', '%s')`,
			ids[domain.Learn], u, ts(domain.Learn)),
	)
	repo := repository.New(e.conn)
	for _, g := range domain.Groups() {
		t.Run(string(g), func(t *testing.T) {
			hits, err := repo.Search(ctx, g, domain.Query{Text: "okapi", Limit: 6})
			if err != nil {
				t.Fatal(err)
			}
			if len(hits) != 1 || hits[0].ID != ids[g] {
				t.Fatalf("hits = %+v, want only %s", hits, ids[g])
			}
			if !hits[0].PublishedAt.Equal(dates[g]) || hits[0].PublishedAt.Location() != time.UTC {
				t.Fatalf("PublishedAt = %v, want %v in UTC", hits[0].PublishedAt, dates[g])
			}
		})
	}
}
