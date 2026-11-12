// Package cursor encodes and verifies opaque pagination cursors (ADR-0004).
//
// Wire format: base64url(no padding) of
//
//	version(1) | len|Sort | len|ID | len|FilterHash | HMAC-SHA256(all preceding bytes)
//
// where len is an unsigned varint. The key is a server secret; a cursor is
// bound to the filter it was issued for so it cannot be replayed on another query.
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

	"github.com/Y4NN777/226os/src/api/internal/platform/apperr"
)

const version byte = 1

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
}

// New builds a Codec; the key must be at least 32 bytes. The error never echoes the key.
func New(key []byte) (*Codec, error) {
	if len(key) < minKeyLen {
		return nil, errors.New("cursor key must be at least 32 bytes")
	}
	if bytes.Count(key, key[:1]) == len(key) {
		return nil, errors.New("cursor key must not repeat a single byte")
	}
	return &Codec{key: append([]byte(nil), key...)}, nil
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
	p, ok := parse(body)
	if !ok || subtle.ConstantTimeCompare([]byte(p.FilterHash), []byte(filterHash)) != 1 {
		return Payload{}, invalid()
	}
	return p, nil
}

func parse(body []byte) (Payload, bool) {
	if body[0] != version {
		return Payload{}, false
	}
	rest := body[1:]
	var fields [3]string
	for i := range fields {
		n, w := binary.Uvarint(rest)
		// A field is a few dozen bytes; bounding n first keeps the int conversion lossless.
		if w <= 0 || n > math.MaxInt32 || int(n) > len(rest)-w {
			return Payload{}, false
		}
		size := int(n)
		fields[i] = string(rest[w : w+size])
		rest = rest[w+size:]
	}
	if len(rest) != 0 {
		return Payload{}, false
	}
	return Payload{Sort: fields[0], ID: fields[1], FilterHash: fields[2]}, true
}
