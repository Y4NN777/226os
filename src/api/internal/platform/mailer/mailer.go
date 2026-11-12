// Package mailer is the outgoing-mail port. It exists so that no domain code imports net/smtp:
// the session mail flows (S1: verification, password reset) depend on the Mailer interface only,
// the composition root picks the SMTP adapter, the Disabled stub or, in tests, the Fake.
// It uses the standard library only.
package mailer

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"
)

// Sentinel errors. Use errors.Is: they survive wrapping. Error texts carry the SMTP phase and the
// numeric reply code only, never credentials and never a server banner.
var (
	// ErrNotConfigured means mail cannot be sent: no SMTP settings, or settings the adapter refuses.
	ErrNotConfigured = errors.New("mailer: not configured")
	// ErrInvalidMessage means the message was refused before any network I/O.
	ErrInvalidMessage = errors.New("mailer: invalid message")
	// ErrTemporary means a retry may succeed: network failure, timeout, or a 4xx reply.
	ErrTemporary = errors.New("mailer: temporary failure")
	// ErrPermanent means a retry will not help: a 5xx reply or a refused TLS requirement.
	ErrPermanent = errors.New("mailer: permanent failure")
)

// Limits enforced on every message.
const (
	maxRecipients   = 50
	maxSubjectRunes = 200
	maxTextBytes    = 262144
)

// Message is a plain-text mail. The sender is fixed by the adapter configuration.
type Message struct {
	To      []string
	Subject string
	Text    string
}

// Mailer sends one message. Implementations bound the call by ctx.
type Mailer interface {
	Send(ctx context.Context, m Message) error
}

// Disabled is the Mailer used when no SMTP settings exist: every send fails with ErrNotConfigured.
type Disabled struct{}

// Send always returns ErrNotConfigured.
func (Disabled) Send(context.Context, Message) error { return ErrNotConfigured }

// validate checks m and returns its parsed recipients. Error texts never echo message content.
func validate(m Message) ([]*mail.Address, error) {
	if n := len(m.To); n < 1 || n > maxRecipients {
		return nil, fmt.Errorf("%w: between 1 and %d recipients required", ErrInvalidMessage, maxRecipients)
	}
	to := make([]*mail.Address, len(m.To))
	for i, raw := range m.To {
		a, err := parseAddress(raw)
		if err != nil {
			return nil, fmt.Errorf("%w: recipient %d is not a valid address", ErrInvalidMessage, i+1)
		}
		to[i] = a
	}
	if n := utf8.RuneCountInString(m.Subject); n < 1 || n > maxSubjectRunes || !utf8.ValidString(m.Subject) {
		return nil, fmt.Errorf("%w: subject must be valid UTF-8 of 1 to %d characters", ErrInvalidMessage, maxSubjectRunes)
	}
	if strings.ContainsFunc(m.Subject, isControl) {
		return nil, fmt.Errorf("%w: subject contains a control character", ErrInvalidMessage)
	}
	if n := len(m.Text); n < 1 || n > maxTextBytes || !utf8.ValidString(m.Text) {
		return nil, fmt.Errorf("%w: text must be valid UTF-8 of 1 to %d bytes", ErrInvalidMessage, maxTextBytes)
	}
	return to, nil
}

// isControl reports the C0 controls other than tab, and DEL: they have no place in a Subject.
func isControl(r rune) bool { return r < ' ' && r != '\t' || r == 0x7f }

// Address limits (RFC 5321 section 4.5.3 and RFC 1035).
const (
	maxAddressLen = 254
	maxLocalLen   = 64
	maxLabelLen   = 63
)

var errBadAddress = errors.New("not a bare ASCII dot-atom address")

// ValidateAddress accepts only a bare ASCII address: a dot-atom local part (RFC 5322 atext characters
// separated by single dots, no quoted string), "@", and a plain host name. Display names, angle
// brackets, comments, quoted local parts, IP literals, spaces, control characters and every
// non-ASCII byte are refused (SMTPUTF8 is not supported), so the text written into the SMTP
// envelope is exactly the address that was validated. The error never echoes the input.
func ValidateAddress(raw string) error {
	local, domain, ok := strings.Cut(raw, "@")
	if !ok || len(raw) > maxAddressLen || len(local) > maxLocalLen || !dotAtom(local) || !hostName(domain) {
		return errBadAddress
	}
	return nil
}

// dotAtom reports whether s is one or more atext runs separated by single dots.
func dotAtom(s string) bool {
	for part := range strings.SplitSeq(s, ".") {
		if part == "" || strings.ContainsFunc(part, func(r rune) bool { return !isAtext(r) }) {
			return false
		}
	}
	return true
}

func isAtext(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || strings.ContainsRune("!#$%&'*+-/=?^_`{|}~", r)
}

// hostName reports whether s is a plain host name: labels of ASCII letters, digits and hyphens,
// none starting or ending with a hyphen. The 253-character host limit is implied by the 254-character
// limit on the whole address.
func hostName(s string) bool {
	for label := range strings.SplitSeq(s, ".") {
		if label == "" || len(label) > maxLabelLen || label[0] == '-' || label[len(label)-1] == '-' ||
			strings.ContainsFunc(label, func(r rune) bool { return r != '-' && !isAlnum(r) }) {
			return false
		}
	}
	return true
}

func isAlnum(r rune) bool {
	return r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9'
}

// parseAddress validates raw and returns it as a mail.Address with no display name.
func parseAddress(raw string) (*mail.Address, error) {
	if err := ValidateAddress(raw); err != nil {
		return nil, err
	}
	return &mail.Address{Address: raw}, nil
}
