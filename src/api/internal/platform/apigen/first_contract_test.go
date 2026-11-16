package apigen_test

import (
	"strings"
	"testing"
)

// TestFirstContractDecisions locks the three owner decisions in the first
// contract change (TODO.md: C2 topic code, C1 owner-only decideApplication,
// C14b profile_visibility). Handlers and the column migration stay out of scope.

func TestFirstContractDecisions(t *testing.T) {
	s, ops := loadSpec(t)
	byID := map[string]operation{}
	for _, op := range ops {
		byID[op.id] = op
	}

	topic := paramByName(byID["listLearnArticles"], "topic")
	if topic == nil {
		t.Error("listLearnArticles has no topic parameter")
	} else {
		if ref := schemaRef(topic["schema"]); ref != "#/components/schemas/TopicCode" {
			t.Errorf("listLearnArticles topic schema = %q, want TopicCode", ref)
		}
		desc, _ := topic["description"].(string)
		if !strings.Contains(desc, "learn_topics.code") || strings.Contains(desc, "article tag") {
			t.Errorf("listLearnArticles topic description = %q, want the topic code, not a tag", desc)
		}
	}

	seriesTopic := paramByName(byID["listTrainingSeries"], "topic")
	if seriesTopic == nil {
		t.Error("listTrainingSeries has no topic parameter")
	} else if ref := schemaRef(seriesTopic["schema"]); ref != "#/components/schemas/TopicCode" {
		t.Errorf("listTrainingSeries topic schema = %q, want TopicCode", ref)
	}

	code := s.Components.Schemas["TopicCode"]
	if code == nil {
		t.Error("TopicCode schema is missing")
	} else {
		if code["pattern"] != `^[a-z][a-z0-9-]{1,29}$` {
			t.Errorf("TopicCode pattern = %v, want the learn_topics check", code["pattern"])
		}
		if n, ok := asInt(code["minLength"]); !ok || n != 2 {
			t.Errorf("TopicCode minLength = %v, want 2", code["minLength"])
		}
		if n, ok := asInt(code["maxLength"]); !ok || n != 30 {
			t.Errorf("TopicCode maxLength = %v, want 30", code["maxLength"])
		}
	}

	create := s.Components.Schemas["ArticleCreate"]
	if ref := schemaRef(prop(create, "topic")); ref != "#/components/schemas/TopicCode" {
		t.Errorf("ArticleCreate.topic schema = %q, want TopicCode", ref)
	}
	for _, name := range stringList(create["required"]) {
		if name == "topic" {
			t.Error("ArticleCreate.topic is required, want optional")
		}
	}

	decide := operationMap(t, s, "/applications/{applicationId}", "patch")
	allow, _ := mapField(decide["x-authz"], "allow").(string)
	if strings.Contains(allow, "maintainer") || !strings.Contains(allow, "owner") || !strings.Contains(allow, "applicant") {
		t.Errorf("decideApplication allow = %q, want owner and applicant, with no maintainer", allow)
	}

	visibility := s.Components.Schemas["ProfileVisibility"]
	if visibility == nil {
		t.Error("ProfileVisibility schema is missing")
	} else {
		if visibility["default"] != "public" {
			t.Errorf("ProfileVisibility default = %v, want public", visibility["default"])
		}
		gotEnum := map[string]bool{}
		for _, v := range stringList(visibility["enum"]) {
			gotEnum[v] = true
		}
		for _, want := range []string{"public", "members", "private"} {
			if !gotEnum[want] {
				t.Errorf("ProfileVisibility enum missing %q (got %v)", want, visibility["enum"])
			}
		}
	}
	for _, schemaName := range []string{"MyProfile", "PublicProfile", "ProfileUpdate"} {
		got := schemaRef(prop(s.Components.Schemas[schemaName], "profile_visibility"))
		if got != "#/components/schemas/ProfileVisibility" {
			t.Errorf("%s.profile_visibility schema = %q, want ProfileVisibility", schemaName, got)
		}
	}
	for _, schemaName := range []string{"MyProfile", "PublicProfile"} {
		found := false
		for _, name := range stringList(s.Components.Schemas[schemaName]["required"]) {
			if name == "profile_visibility" {
				found = true
			}
		}
		if !found {
			t.Errorf("%s does not require profile_visibility", schemaName)
		}
	}
	for _, name := range stringList(s.Components.Schemas["ProfileUpdate"]["required"]) {
		if name == "profile_visibility" {
			t.Error("ProfileUpdate.profile_visibility is required, want optional on the merge patch")
		}
	}

	profile := operationMap(t, s, "/profiles/{username}", "get")
	authz, _ := profile["x-authz"].(map[string]any)
	if !denyIs404(authz["deny_status"]) {
		t.Errorf("getProfile deny_status = %v, want 404", authz["deny_status"])
	}
}

func paramByName(op operation, name string) map[string]any {
	for _, p := range op.params {
		if p["name"] == name {
			return p
		}
	}
	return nil
}

func prop(schema map[string]any, name string) any {
	props, _ := schema["properties"].(map[string]any)
	if props == nil {
		return nil
	}
	return props[name]
}

func schemaRef(v any) string {
	s, ok := v.(map[string]any)
	if !ok {
		return ""
	}
	if ref, ok := s["$ref"].(string); ok {
		return ref
	}
	items, _ := s["anyOf"].([]any)
	for _, item := range items {
		if ref := schemaRef(item); ref != "" {
			return ref
		}
	}
	return ""
}

func operationMap(t *testing.T, s spec, path, method string) map[string]any {
	t.Helper()
	body, ok := s.Paths[path][method].(map[string]any)
	if !ok {
		t.Fatalf("missing %s %s", method, path)
	}
	return body
}

func mapField(v any, key string) any {
	m, _ := v.(map[string]any)
	return m[key]
}

func stringList(v any) []string {
	list, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func asInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int64:
		return int(n), true
	case uint64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}

func denyIs404(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	for _, status := range m {
		switch n := status.(type) {
		case int:
			if n == 404 {
				return true
			}
		case uint64:
			if n == 404 {
				return true
			}
		case float64:
			if n == 404 {
				return true
			}
		}
	}
	return false
}
