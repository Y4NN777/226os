package domain

import (
	"testing"
	"time"
)

func TestSessionActive(t *testing.T) {
	now := time.Date(2026, 11, 13, 12, 0, 0, 0, time.UTC)
	live := Session{
		IdleExpiresAt:     now.Add(time.Hour),
		AbsoluteExpiresAt: now.Add(24 * time.Hour),
	}
	cases := []struct {
		name   string
		mutate func(*Session)
		want   bool
	}{
		{"live", func(*Session) {}, true},
		{"revoked", func(s *Session) { t := now; s.RevokedAt = &t }, false},
		{"idle expired", func(s *Session) { s.IdleExpiresAt = now }, false},
		{"absolute expired", func(s *Session) { s.AbsoluteExpiresAt = now }, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := live
			tc.mutate(&s)
			if got := s.Active(now); got != tc.want {
				t.Fatalf("Active=%v want %v", got, tc.want)
			}
		})
	}
}
