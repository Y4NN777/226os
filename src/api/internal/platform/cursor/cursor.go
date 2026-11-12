// Package cursor encodes and verifies opaque pagination cursors (ADR-0004).
//
// Wire format: base64url(no padding) of
//
//	version(2) | len|Sort | len|ID | len|FilterHash | issuedAt(8) | HMAC-SHA256(all preceding bytes)
//
// where len is an unsigned varint and issuedAt is big-endian unix seconds. The
// key is a server secret; a cursor is bound to the filter it was issued for so
// it cannot be replayed on another query. A cursor expires DefaultTTL (24 h,
// WithTTL) after issue and is refused when issued more than 5 minutes in the
// future (clock skew). Version 1 (no timestamp) was never issued in production
// and is refused as invalid.
package cursor

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const version byte = 2

const minKeyLen = 32

// Payload is the keyset position a cursor carries.
type Payload struct {
	Sort       string // sort key value, e.g. the last row's ordering column
	ID         string // tie-breaker UUID of the last row
	FilterHash string // hash of the query filters the cursor was issued for
}

// Codec signs and verifies cursors with a secret key.
type Codec struct {
	key []byte
	now func() time.Time
	ttl time.Duration
}

// Option tunes a Codec.
type Option func(*Codec)

// WithClock injects the time source (tests); the default is time.Now.
func WithClock(now func() time.Time) Option { return func(c *Codec) { c.now = now } }

// WithTTL sets how long an issued cursor stays valid; the default is DefaultTTL.
func WithTTL(ttl time.Duration) Option { return func(c *Codec) { c.ttl = ttl } }

// DefaultTTL is the lifetime of a cursor unless WithTTL says otherwise.
const DefaultTTL = 24 * time.Hour

// maxSkew is how far in the future an issued-at may be (clock skew between instances).
const maxSkew = 5 * time.Minute

// New builds a Codec; the key must be at least 32 bytes. The error never echoes the key.
func New(key []byte, opts ...Option) (*Codec, error) {
	if len(key) < minKeyLen {
		return nil, errors.New("cursor key must be at least 32 bytes")
	}
	if bytes.Count(key, key[:1]) == len(key) {
		return nil, errors.New("cursor key must not repeat a single byte")
	}
	c := &Codec{key: append([]byte(nil), key...), now: time.Now, ttl: DefaultTTL}
	for _, o := range opts {
		o(c)
	}
	return c, nil
}

func (c *Codec) sign(body []byte) []byte {
	m := hmac.New(sha256.New, c.key)
	m.Write(body) // hash.Hash.Write never returns an error
	return m.Sum(nil)
}

// Encode returns the opaque cursor string for p.
func (c *Codec) Encode(p Payload) string {
	body := []byte{version}
	for _, f := range []string{p.Sort, p.ID, p.FilterHash} {
		body = binary.AppendUvarint(body, uint64(len(f)))
		body = append(body, f...)
	}
	body = binary.BigEndian.AppendUint64(body, uint64(c.now().Unix())) //nolint:gosec // unix seconds are positive
	return base64.RawURLEncoding.EncodeToString(append(body, c.sign(body)...))
}

func invalid() error {
	return apperr.New(apperr.CodeInvalidCursor, "The cursor is invalid or expired.")
}

// Decode verifies s and returns its payload. Any malformed, tampered,
// wrong-version, wrong-key or wrong-filter cursor yields INVALID_CURSOR.
func (c *Codec) Decode(s, filterHash string) (Payload, error) {
	raw, err := base64.RawURLEncoding.Strict().DecodeString(s)
	if err != nil || len(raw) <= sha256.Size {
		return Payload{}, invalid()
	}
	body, mac := raw[:len(raw)-sha256.Size], raw[len(raw)-sha256.Size:]
	if !hmac.Equal(mac, c.sign(body)) {
		return Payload{}, invalid()
	}
	p, issuedAt, ok := parse(body)
	if !ok || !c.fresh(issuedAt) || subtle.ConstantTimeCompare([]byte(p.FilterHash), []byte(filterHash)) != 1 {
		return Payload{}, invalid()
	}
	return p, nil
}

// fresh reports whether a cursor issued at issuedAt (unix seconds) is within its TTL
// and not further in the future than the allowed skew.
func (c *Codec) fresh(issuedAt int64) bool {
	age := c.now().Sub(time.Unix(issuedAt, 0))
	return age <= c.ttl && age >= -maxSkew
}

func parse(body []byte) (Payload, int64, bool) {
	if body[0] != version {
		return Payload{}, 0, false
	}
	rest := body[1:]
	var fields [3]string
	for i := range fields {
		n, w := binary.Uvarint(rest)
		// A field is a few dozen bytes; bounding n first keeps the int conversion lossless.
		if w <= 0 || n > math.MaxInt32 || int(n) > len(rest)-w {
			return Payload{}, 0, false
		}
		size := int(n)
		fields[i] = string(rest[w : w+size])
		rest = rest[w+size:]
	}
	if len(rest) != 8 {
		return Payload{}, 0, false
	}
	issuedAt := int64(binary.BigEndian.Uint64(rest)) //nolint:gosec // wraps to a negative, i.e. ancient or absurd, instant that fresh refuses
	return Payload{Sort: fields[0], ID: fields[1], FilterHash: fields[2]}, issuedAt, true
}
