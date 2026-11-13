// Package password hashes and verifies account passwords with Argon2id
// (ADR-0003). Hashes are PHC strings that carry their own parameters, so the
// profile can be raised later and old hashes rehashed on successful sign-in.
//
// The work factor is m=64 MiB, t=3, p=1, never below the OWASP floor
// (m=19 MiB, t=2, p=1). Concurrency is capped by a semaphore because each hash
// is memory-hungry. An unknown email still pays one Argon2id against a dummy
// hash so sign-in timing does not reveal whether the account exists.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// OWASP floor and the chosen profile (ADR-0003).
const (
	memKiB  = 64 * 1024
	time    = 3
	threads = 1
	saltLen = 16
	keyLen  = 32

	floorMemKiB = 19 * 1024
	floorTime   = 2
)

// maxConcurrent bounds simultaneous Argon2id runs: each holds ~64 MiB.
const maxConcurrent = 4

var sem = make(chan struct{}, maxConcurrent)

var (
	// ErrInvalidPassword means the password is outside the contract length
	// (12 to 128 characters; composition rules are deliberately absent).
	ErrInvalidPassword = errors.New("password: must be 12 to 128 characters")
	// ErrMalformedHash means a stored hash is not a PHC Argon2id string this
	// package produced or accepts.
	ErrMalformedHash = errors.New("password: malformed hash")
)

// Normalize applies NFKC-ish preparation: trim and fold CR/LF so a pasted
// password cannot smuggle control characters. Unicode normalization beyond
// that is out of scope at P0 (ADR-0003).
func Normalize(pw string) string {
	return strings.Map(func(r rune) rune {
		if r == '\r' || r == '\n' {
			return -1
		}
		return r
	}, strings.TrimSpace(pw))
}

// Valid reports whether pw satisfies the length contract after Normalize.
func Valid(pw string) bool {
	n := len(Normalize(pw))
	return n >= 12 && n <= 128
}

// Hash returns the PHC string for pw under the current profile.
func Hash(pw string) (string, error) {
	if !Valid(pw) {
		return "", ErrInvalidPassword
	}
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: salt: %w", err)
	}
	return encode(pw, salt, memKiB, time, uint8(threads)), nil
}

// Verify reports whether pw matches the PHC hash. It never reveals which part
// failed and compares in constant time. A malformed hash is an error, not a
// mismatch: stored data corruption must not look like a wrong password.
func Verify(pw, phc string) (bool, error) {
	m, err := decode(phc)
	if err != nil {
		return false, err
	}
	sem <- struct{}{}
	defer func() { <-sem }()
	got := argon2.IDKey([]byte(Normalize(pw)), m.salt, m.time, m.mem, m.threads, uint32(len(m.key)))
	return subtle.ConstantTimeCompare(got, m.key) == 1, nil
}

// dummyHash is a valid PHC string under the floor profile, computed once. It
// lets an unknown-email sign-in pay the same Argon2id cost as a real one.
var dummyHash = mustHash("dummy-password-for-timing-equalization")

// VerifyDummy runs one Argon2id against the dummy hash and always reports a
// mismatch. Callers use it on the unknown-email path to equalize timing.
func VerifyDummy(pw string) {
	_, _ = Verify(pw, dummyHash)
}

func mustHash(pw string) string {
	salt := make([]byte, saltLen)
	if _, err := rand.Read(salt); err != nil {
		panic(err)
	}
	// Same profile as Hash: the point is equalizing sign-in timing.
	return encode(pw, salt, memKiB, time, uint8(threads))
}

type params struct {
	mem       uint32
	time      uint32
	threads   uint8
	salt, key []byte
}

func encode(pw string, salt []byte, mem, t uint32, p uint8) string {
	key := argon2.IDKey([]byte(Normalize(pw)), salt, t, mem, p, keyLen)
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, mem, t, p,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key))
}

func decode(phc string) (params, error) {
	var m params
	parts := strings.Split(phc, "$")
	// "$argon2id$v=19$m=...,t=...,p=...$salt$key" splits into 6 fields, first empty.
	if len(parts) != 6 || parts[1] != "argon2id" {
		return m, ErrMalformedHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return m, ErrMalformedHash
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m.mem, &m.time, &m.threads); err != nil {
		return m, ErrMalformedHash
	}
	// Never accept a hash weaker than the OWASP floor.
	if m.mem < floorMemKiB || m.time < floorTime || m.threads < 1 {
		return m, ErrMalformedHash
	}
	var err error
	if m.salt, err = base64.RawStdEncoding.DecodeString(parts[4]); err != nil || len(m.salt) < 8 {
		return m, ErrMalformedHash
	}
	if m.key, err = base64.RawStdEncoding.DecodeString(parts[5]); err != nil || len(m.key) < 16 {
		return m, ErrMalformedHash
	}
	return m, nil
}
