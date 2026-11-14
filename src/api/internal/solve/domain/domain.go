// Package domain holds the Solve model (US-SOLVE-001): questions and answers. It has no database or
// HTTP types.
package domain

import (
	"errors"
	"time"
)

// Errors the store returns; the service maps them to problem+json.
var (
	// ErrNotFound means the question is missing or removed: one answer for both, so a removed
	// question is indistinguishable from an unknown one.
	ErrNotFound = errors.New("solve item not found")
	// ErrForbidden means the question is visible but the caller is not its author.
	ErrForbidden = errors.New("solve question not owned by the caller")
	// ErrLinkTarget means a link target is missing, not published, or not public.
	ErrLinkTarget = errors.New("solve link target not linkable")
	// ErrUnknownCountry means the country code is not a countries row.
	ErrUnknownCountry = errors.New("solve country unknown")
)

// Derived solve states of a question (never stored, never client-set).
const (
	StateOpen     = "open"
	StateAnswered = "answered"
	StateResolved = "resolved"
)

// LinkTypes are the link targets a question can store. content_relations keeps each pair in one
// canonical orientation with the question as the source, and has no question-question pair and no
// series side.
var LinkTypes = []string{"resource", "project", "article"}

// Author is the public byline. Country is empty unless the profile and the country are public.
type Author struct {
	ID, Username, DisplayName, Country string
}

// Snippet is code stored for display only; it is never executed.
type Snippet struct {
	Language, Code string
}

// Link is a resolved related item (read side).
type Link struct {
	Type, ID, Title, Lang string
}

// LinkRef is a typed link to write.
type LinkRef struct {
	Type, ID string
}

// Question is a published question with its derived state inputs. AcceptedAnswerID is empty when no
// answer is accepted; Country is empty when unset.
type Question struct {
	ID, Title, BodyMD, Lang, Country, AcceptedAnswerID string
	Tags                                               []string
	Snippets                                           []Snippet
	Author                                             Author
	AnswerCount, VoteScore                             int
	Related                                            []Link
	CreatedAt, UpdatedAt                               time.Time
	// Rank is the search relevance of a listed question (0 without a query); it only feeds the cursor.
	Rank float32
}

// State is open without answers, answered with answers and none accepted, resolved with an accepted one.
func (q Question) State() string {
	switch {
	case q.AcceptedAnswerID != "":
		return StateResolved
	case q.AnswerCount > 0:
		return StateAnswered
	default:
		return StateOpen
	}
}

// NewQuestion is a validated question to insert. Country is empty when not given.
type NewQuestion struct {
	AuthorID, Title, BodyMD, Lang, Country string
	Tags                                   []string
	Snippets                               []Snippet
	Links                                  []LinkRef
}

// QuestionPatch is a validated JSON Merge Patch: nil means unchanged. Country "" clears it. A set
// Snippets slice replaces them (empty clears). The language is immutable and is not here.
type QuestionPatch struct {
	Title, BodyMD, Country *string
	Tags                   *[]string
	Snippets               *[]Snippet
	Links                  *[]LinkRef
}

// QuestionQuery is one question page. Empty filters are unset. Limit includes the look-ahead row.
type QuestionQuery struct {
	Q, Tag, State, Country, Lang string
	AfterCreated                 *time.Time
	AfterRank                    float32
	AfterID                      string
	Limit                        int
}

// Answer is a published answer. Snippets are always empty at P0 (no column, migration 00005).
type Answer struct {
	ID, QuestionID, BodyMD, Lang string
	Snippets                     []Snippet
	Sources                      []string
	Author                       Author
	VoteScore                    int
	Accepted                     bool
	CreatedAt, UpdatedAt         time.Time
}

// NewAnswer is a validated answer to insert.
type NewAnswer struct {
	QuestionID, AuthorID, BodyMD, Lang string
	Sources                            []string
}

// AnswerKey is the position after which an answer page continues: the accepted answer is rank 0 and
// the others rank 1, NegScore is minus the vote score (the order is accepted, score desc, oldest, id).
type AnswerKey struct {
	Rank, NegScore int
	CreatedAt      time.Time
	ID             string
}

// AnswerQuery is one answer page of a question. After is nil on the first page. Limit includes the
// look-ahead row.
type AnswerQuery struct {
	QuestionID string
	After      *AnswerKey
	Limit      int
}
