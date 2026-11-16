package domain

import (
	"testing"
	"time"
)

func TestSourceStatusNeverInventsStale(t *testing.T) {
	checked := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name, kind, resource, source, verified string
		checked                                *time.Time
		wantStatus                             string
		wantTime                               bool
	}{
		{"current verified source", "official_docs", "verified", "s1", "s1", &checked, "verified", true},
		{"claim that is the verified source", "claim", "verified", "s1", "s1", &checked, "verified", true},
		{"other source of a verified resource", "official_docs", "verified", "s2", "s1", &checked, "unverified", false},
		{"claim that is not the verified source", "claim", "verified", "s2", "s1", &checked, "claimed", false},
		{"claim on an unverified resource", "claim", "unverified", "s2", "", nil, "claimed", false},
		{"old check stays unverified without a window", "official_docs", "unverified", "s2", "", &checked, "unverified", false},
		{"empty source id is not verified", "official_docs", "verified", "", "", &checked, "unverified", false},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, at := SourceStatus(tc.kind, tc.resource, tc.source, tc.verified, tc.checked)
			if status != tc.wantStatus {
				t.Fatalf("status = %s, want %s", status, tc.wantStatus)
			}
			if tc.wantTime {
				if at == nil || !at.Equal(checked) {
					t.Fatalf("verified_at = %v, want %v", at, checked)
				}
				return
			}
			if at != nil {
				t.Fatalf("verified_at = %v, want none", at)
			}
		})
	}
}
