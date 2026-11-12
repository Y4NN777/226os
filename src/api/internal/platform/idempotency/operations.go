package idempotency

import "net/http"

// Operation is one operation that binds the Idempotency-Key parameter: the
// HTTP method and the chi route pattern (base URL included).
type Operation struct {
	Method  string
	Pattern string
}

// operations must equal the operations of api/openapi.yaml that reference the
// IdempotencyKey parameter; operations_test.go enforces it. castVote does not
// bind the key.
var operations = [...]Operation{
	{http.MethodPost, "/api/v1/questions"},
	{http.MethodPost, "/api/v1/questions/{questionId}/answers"},
	{http.MethodPost, "/api/v1/projects"},
	{http.MethodPost, "/api/v1/projects/{projectId}/contribution-needs"},
	{http.MethodPost, "/api/v1/contribution-needs/{needId}/applications"},
	{http.MethodPost, "/api/v1/learn/articles"},
}

// Operations returns the operations the middleware acts on.
func Operations() []Operation {
	out := make([]Operation, len(operations))
	copy(out, operations[:])
	return out
}

func bound(method, pattern string) bool {
	for _, op := range operations {
		if op.Method == method && op.Pattern == pattern {
			return true
		}
	}
	return false
}
