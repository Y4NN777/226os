package apigen_test

import (
	"strings"
	"testing"
)

// TestAssistantContract locks the S9 contract decisions (TODO.md S9, US-INTEL-001, PRD 5.8): one POST
// operation, members only, CSRF protected, a closed outcome (answered or no_source), an AI label,
// server-validated citations, and the 429 / 503 behaviour. Handlers and the provider stay out of scope.
// Helper names carry an "asst" prefix so they never clash with helpers of other contract tests.

func asstOp(t *testing.T) map[string]any {
	t.Helper()
	s, _ := loadSpec(t)
	body, ok := s.Paths["/assistant/answers"]["post"].(map[string]any)
	if !ok {
		t.Fatal("POST /assistant/answers is missing")
	}
	return body
}

func asstMap(v any) map[string]any { m, _ := v.(map[string]any); return m }

func asstStrings(v any) []string {
	list, _ := v.([]any)
	out := make([]string, 0, len(list))
	for _, item := range list {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func asstHas(list []string, want string) bool {
	for _, s := range list {
		if s == want {
			return true
		}
	}
	return false
}

func TestAssistantOperationShape(t *testing.T) {
	op := asstOp(t)
	if op["operationId"] != "askAssistant" {
		t.Errorf("operationId = %v, want askAssistant", op["operationId"])
	}
	if op["x-priority"] != "P0" {
		t.Errorf("x-priority = %v, want P0", op["x-priority"])
	}
	if !asstHas(asstStrings(op["x-spec"]), "US-INTEL-001") {
		t.Errorf("x-spec = %v, want US-INTEL-001", op["x-spec"])
	}
	authz := asstMap(op["x-authz"])
	if allow, _ := authz["allow"].(string); allow != "member" {
		t.Errorf("x-authz allow = %q, want member (no anonymous access at P0)", allow)
	}
	if !strings.Contains(strings.ToLower(asstString(authz["notes"])), "verified email") {
		t.Errorf("x-authz notes must state the verified email rule: %v", authz["notes"])
	}
	sec, _ := op["security"].([]any)
	if len(sec) != 1 {
		t.Fatalf("security = %v, want one requirement (session and CSRF)", op["security"])
	}
	req := asstMap(sec[0])
	if req["cookieAuth"] == nil || req["csrfToken"] == nil {
		t.Errorf("security = %v, want cookieAuth and csrfToken", req)
	}
	for _, p := range asstParams(op) {
		if strings.Contains(p, "IdempotencyKey") {
			t.Errorf("askAssistant must not bind Idempotency-Key (a retry is a new, costed question)")
		}
	}
}

func asstString(v any) string { s, _ := v.(string); return s }

func asstParams(op map[string]any) []string {
	list, _ := op["parameters"].([]any)
	var out []string
	for _, p := range list {
		out = append(out, asstString(asstMap(p)["$ref"]))
	}
	return out
}

func TestAssistantResponses(t *testing.T) {
	op := asstOp(t)
	resp := asstMap(op["responses"])
	for _, code := range []string{"200", "400", "401", "403", "422", "429", "500", "503"} {
		if resp[code] == nil {
			t.Errorf("response %s is missing", code)
		}
	}
	if ref := asstString(asstMap(resp["429"])["$ref"]); ref != "#/components/responses/TooManyRequests" {
		t.Errorf("429 = %q, want TooManyRequests (Retry-After)", ref)
	}
	if ref := asstString(asstMap(resp["503"])["$ref"]); ref != "#/components/responses/ServiceUnavailable" {
		t.Errorf("503 = %q, want ServiceUnavailable (assistant unavailable)", ref)
	}
	ok := asstMap(resp["200"])
	schema := asstMap(asstMap(asstMap(ok["content"])["application/json"])["schema"])
	if ref := asstString(schema["$ref"]); ref != "#/components/schemas/AssistantAnswer" {
		t.Errorf("200 schema = %q, want AssistantAnswer", ref)
	}
}

func TestAssistantSchemas(t *testing.T) {
	s, _ := loadSpec(t)
	q := s.Components.Schemas["AssistantQuestion"]
	if q == nil {
		t.Fatal("AssistantQuestion is missing")
	}
	if q["additionalProperties"] != false {
		t.Error("AssistantQuestion must reject unknown fields")
	}
	if !asstHas(asstStrings(q["required"]), "question") {
		t.Errorf("AssistantQuestion required = %v, want question", q["required"])
	}
	text := asstMap(asstMap(q["properties"])["question"])
	if _, ok := text["maxLength"]; !ok {
		t.Error("question must be length bounded")
	}

	a := s.Components.Schemas["AssistantAnswer"]
	if a == nil {
		t.Fatal("AssistantAnswer is missing")
	}
	props := asstMap(a["properties"])
	for _, name := range []string{"outcome", "generated_by", "answer", "sources", "how_to_verify", "lang"} {
		if props[name] == nil {
			t.Errorf("AssistantAnswer.%s is missing", name)
		}
		if !asstHas(asstStrings(a["required"]), name) {
			t.Errorf("AssistantAnswer.%s must be required", name)
		}
	}
	outcome := asstStrings(asstMap(props["outcome"])["enum"])
	if len(outcome) != 2 || !asstHas(outcome, "answered") || !asstHas(outcome, "no_source") {
		t.Errorf("outcome enum = %v, want answered and no_source", outcome)
	}
	if by := asstStrings(asstMap(props["generated_by"])["enum"]); len(by) != 1 || by[0] != "ai" {
		t.Errorf("generated_by enum = %v, want only ai (AI-9: labelled AI, never a community answer)", by)
	}
	if _, ok := asstMap(props["sources"])["maxItems"]; !ok {
		t.Error("sources must be bounded")
	}

	src := s.Components.Schemas["AssistantSource"]
	if src == nil {
		t.Fatal("AssistantSource is missing")
	}
	for _, name := range []string{"type", "id", "title", "url"} {
		if !asstHas(asstStrings(src["required"]), name) {
			t.Errorf("AssistantSource.%s must be required", name)
		}
	}
	if asstMap(asstMap(src["properties"])["published_at"]) == nil {
		t.Error("AssistantSource.published_at is missing (AI-5: source dates)")
	}
	types := asstStrings(asstMap(asstMap(src["properties"])["type"])["enum"])
	for _, forbidden := range []string{"answer", "vote"} {
		if asstHas(types, forbidden) {
			t.Errorf("AssistantSource.type allows %q; an AI answer is never an answer or a vote target", forbidden)
		}
	}
}
