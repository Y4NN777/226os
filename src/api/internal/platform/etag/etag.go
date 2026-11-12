// Package etag implements strong entity tags and the If-Match precondition
// (RFC 9110 sections 8.8.3 and 13.1.1, RFC 6585 section 3; ADR-0004 point 7).
package etag

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"net/http"
	"strings"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const (
	headerIfMatch = "If-Match"
	headerETag    = "ETag"
	tagBytes      = 16
)

// Tag derives a strong entity tag from parts (for example an id and an
// updated_at in RFC 3339 nano). Each part is hashed with an 8-byte big-endian length
// prefix, so the encoding is unambiguous whatever the parts contain (NUL included): a
// different split of the same characters yields a different tag, and Tag() differs from
// Tag("").
func Tag(parts ...string) string {
	h := sha256.New()
	var n [8]byte
	for _, p := range parts {
		binary.BigEndian.PutUint64(n[:], uint64(len(p)))
		h.Write(n[:])
		h.Write([]byte(p))
	}
	return `"` + base64.RawURLEncoding.EncodeToString(h.Sum(nil)[:tagBytes]) + `"`
}

// Set writes the ETag response header.
func Set(w http.ResponseWriter, tag string) {
	w.Header().Set(headerETag, tag)
}

// Check enforces the If-Match precondition of an unsafe request against the
// representation whose current entity tag is current (empty when the resource
// does not exist). It returns nil when the precondition holds, otherwise an
// *apperr.Error: 428 when the header is absent or has no entity tag (empty list
// elements are ignored, RFC 9110 section 5.6.1.2), 400 when it is syntactically
// invalid, 412 when no tag matches. Comparison is strong
// (RFC 9110 section 13.1.1): weak tags never match. Error details never echo
// the header value or the current tag.
func Check(r *http.Request, current string) error {
	// Empty list elements are ignored (RFC 9110 section 5.6.1.2: a recipient MUST parse and
	// ignore a reasonable number of them), so a header of empty members only carries no tag.
	header := strings.Trim(strings.Join(r.Header.Values(headerIfMatch), ","), " \t,")
	if header == "" {
		return apperr.New(apperr.CodePreconditionRequired, "If-Match header is required")
	}
	if header == "*" {
		if current == "" {
			return apperr.New(apperr.CodePreconditionFailed, "precondition failed")
		}
		return nil
	}

	matched := false
	for rest := header; rest != ""; rest = strings.TrimLeft(rest, " \t,") {
		var member string
		var weak, ok bool
		member, weak, rest, ok = nextMember(rest)
		if !ok {
			return apperr.New(apperr.CodeMalformedRequest, "If-Match header is malformed").
				WithViolations(apperr.Violation{Field: headerIfMatch, Code: "invalid_format"})
		}
		// Every member is compared so timing does not reveal which one matched.
		if !weak && current != "" &&
			subtle.ConstantTimeCompare([]byte(member), []byte(current)) == 1 {
			matched = true
		}
	}
	if !matched {
		return apperr.New(apperr.CodePreconditionFailed, "precondition failed")
	}
	return nil
}

// nextMember scans one entity-tag (RFC 9110 section 8.8.3) from s, which must
// start at a member (after optional whitespace). It returns the quoted tag,
// whether it was weak, the remainder after the separating comma (empty when
// the list ended; it may start with further empty members) and false when the
// syntax is invalid.
func nextMember(s string) (member string, weak bool, rest string, ok bool) {
	s = strings.TrimLeft(s, " \t")
	if strings.HasPrefix(s, "W/") {
		weak = true
		s = s[2:]
	}
	if s == "" || s[0] != '"' {
		return "", false, "", false
	}
	end := 1
	for end < len(s) && s[end] != '"' {
		if !etagc(s[end]) {
			return "", false, "", false
		}
		end++
	}
	if end == len(s) {
		return "", false, "", false // unterminated quote
	}
	member = s[:end+1]
	s = strings.TrimLeft(s[end+1:], " \t")
	switch {
	case s == "":
		return member, weak, "", true
	case s[0] == ',':
		return member, weak, s[1:], true
	default:
		return "", false, "", false
	}
}

// etagc reports whether c is allowed inside the quotes of an entity tag:
// %x21 / %x23-7E / obs-text (the double quote is handled by the caller).
func etagc(c byte) bool {
	return c == 0x21 || (c >= 0x23 && c <= 0x7E) || c >= 0x80
}
