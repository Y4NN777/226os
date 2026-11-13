// Package domain holds the search vocabulary: groups, hits, query limits and the two text rules
// (query truncation and plain-text snippets). It has no IO.
package domain

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	// MaxTextRunes is the length the server keeps of q before it reaches the tsquery parser.
	// C4 DECIDED by the owner (option 1, PR #92), ADR-0005 point 6.
	MaxTextRunes = 100
	// MaxSnippetRunes is the contract bound of SearchHit.snippet.
	MaxSnippetRunes = 300
)

// Group is one searchable content family (the SearchGroup enum of the contract).
// Answers have no group, so they are not searched.
type Group string

// The four groups, in contract order.
const (
	Questions Group = "questions"
	Resources Group = "resources"
	Projects  Group = "projects"
	Learn     Group = "learn"
)

// Groups lists every group in contract order.
func Groups() []Group { return []Group{Questions, Resources, Projects, Learn} }

// Valid reports whether g is one of the four groups.
func (g Group) Valid() bool {
	switch g {
	case Questions, Resources, Projects, Learn:
		return true
	}
	return false
}

// Query is what a Reader runs for one group. Limit already includes the extra row that
// tells has_more.
type Query struct {
	Text    string
	Country string
	Lang    string
	Limit   int
}

// Hit is one search result. Lang is empty when the content has no language.
type Hit struct {
	Type    Group
	ID      string
	Title   string
	Snippet string
	Lang    string
}

// TruncateQuery applies C4 DECIDED option 1 (ADR-0005 point 6): keep at most MaxTextRunes
// runes, cut at a word boundary, and hard-cut at MaxTextRunes when the first MaxTextRunes
// runes contain no usable space. A boundary that would leave only spaces is ignored (hard cut).
func TruncateQuery(q string) string {
	runes := []rune(q)
	if len(runes) <= MaxTextRunes {
		return q
	}
	cut := MaxTextRunes
	if !unicode.IsSpace(runes[MaxTextRunes]) {
		for i := MaxTextRunes - 1; i > 0; i-- {
			if unicode.IsSpace(runes[i]) {
				cut = i
				break
			}
		}
	}
	out := strings.TrimRightFunc(string(runes[:cut]), unicode.IsSpace)
	if strings.TrimSpace(out) == "" {
		// Only spaces before the boundary: hard-cut instead of emptying the query.
		out = string(runes[:MaxTextRunes])
	}
	return out
}

var markdownLink = regexp.MustCompile(`\[([^\]]*)\]\([^)]*\)`)

// markupRemover drops the characters that make markdown or HTML: a snippet is plain text, so
// no tag can survive (angle brackets go) and emphasis, heading and code markers go too.
var markupRemover = strings.NewReplacer("<", "", ">", "", "`", "", "*", "", "#", "", "~", "")

// Snippet turns a markdown excerpt into at most MaxSnippetRunes runes of plain text.
func Snippet(markdown string) string {
	plain := markupRemover.Replace(markdownLink.ReplaceAllString(markdown, "$1"))
	plain = strings.Join(strings.Fields(plain), " ")
	runes := []rune(plain)
	if len(runes) > MaxSnippetRunes {
		plain = strings.TrimRightFunc(string(runes[:MaxSnippetRunes]), unicode.IsSpace)
	}
	return plain
}
