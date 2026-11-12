// Package apispec embeds the OpenAPI contract so the API can validate requests
// against it at runtime (ADR-0004 point 8).
package apispec

import _ "embed"

// YAML is a byte-identical copy of api/openapi.yaml (the source of truth).
// openapi.gen.yaml is refreshed by scripts/check-generated.sh; the test in this
// package fails when the copy drifts.
//
//go:embed openapi.gen.yaml
var YAML []byte
