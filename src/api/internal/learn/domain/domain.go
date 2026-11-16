// Package domain holds the Learn read and write model (US-LEARN-001). It has no database or HTTP types.
package domain

import (
	"errors"
	"time"
)

// Errors the store returns; the service maps them to problem+json.
var (
	// ErrNotFound means the article or series is missing, removed, or not visible to the viewer.
	ErrNotFound = errors.New("learn item not found")
	// ErrUnknownTopic means the topic code is not a learn_topics row.
	ErrUnknownTopic = errors.New("learn topic unknown")
	// ErrLinkTarget means a link target is missing, not published, or not public.
	ErrLinkTarget = errors.New("learn link target not linkable")
)

// Statuses an article can be created with.
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
)

// LinkTypes are the link targets an article can store: content_relations keeps each pair in one
// canonical orientation (question, resource or project -> article) and has no series side and no
// article-article pair.
var LinkTypes = []string{"question", "resource", "project"}

// Author is the public byline. Country is empty unless the profile and the country are public.
type Author struct {
	ID, Username, DisplayName, Country string
}

// Link is a resolved related item (read side).
type Link struct {
	Type, ID, Title, Lang string
}

// LinkRef is a typed link to write.
type LinkRef struct {
	Type, ID string
}

// ArticleSummary is one list row.
type ArticleSummary struct {
	ID, Kind, Title, Summary, Lang, Level string
	Tags                                  []string
	Author                                Author
	PublishedAt                           time.Time
}

// Article is the full article. BodyMD is the original text, stored and returned unchanged.
type Article struct {
	ID, Kind, Title, Summary, BodyMD, Lang, Level, Status string
	Tags                                                  []string
	Author                                                Author
	Related                                               []Link
	PublishedAt                                           *time.Time
	UpdatedAt                                             time.Time
}

// NewArticle is a validated article to insert. Summary and Topic are empty when not given.
type NewArticle struct {
	AuthorID, Kind, Title, Summary, BodyMD, Lang, Level, Topic, Status string
	Tags                                                               []string
	Links                                                              []LinkRef
}

// ArticleQuery is one article page. Empty filters are unset. Limit includes the look-ahead row.
type ArticleQuery struct {
	Q, Kind, Topic, Level, Lang string
	AfterPublished              *time.Time
	AfterID                     string
	Limit                       int
}

// SeriesQuery is one series page. Empty filters are unset. Limit includes the look-ahead row.
type SeriesQuery struct {
	Topic, Level, Lang string
	AfterPublished     *time.Time
	AfterID            string
	Limit              int
}

// SeriesSummary is one series list row.
type SeriesSummary struct {
	ID, Title, Summary, Lang, Level string
	Author                          Author
	LessonCount                     int
	PublishedAt                     time.Time
}

// Lesson is one published lesson of a series, backed by a published article.
type Lesson struct {
	ID, Title, ArticleID string
	Position             int
}

// Series is a training series with its ordered lessons.
type Series struct {
	ID, Title, Summary, Lang, Level string
	Author                          Author
	Lessons                         []Lesson
	PublishedAt                     time.Time
}
