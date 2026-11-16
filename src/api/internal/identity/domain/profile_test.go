package domain

import "testing"

func TestVisibilityAllows(t *testing.T) {
	cases := []struct {
		v         Visibility
		anon, own bool
		want      bool
	}{
		{VisibilityPublic, true, false, true},
		{VisibilityPublic, false, true, true},
		{VisibilityMembers, true, false, false},
		{VisibilityMembers, false, false, true}, // member caller
		{VisibilityMembers, true, true, true},   // owner is never anonymous
		{VisibilityPrivate, true, false, false},
		{VisibilityPrivate, false, false, false}, // member, not owner
		{VisibilityPrivate, false, true, true},   // owner
		{Visibility("bogus"), false, true, false},
	}
	for _, tc := range cases {
		if got := tc.v.Allows(tc.anon, tc.own); got != tc.want {
			t.Errorf("Allows(%v, anon=%v, own=%v) = %v, want %v", tc.v, tc.anon, tc.own, got, tc.want)
		}
	}
}
