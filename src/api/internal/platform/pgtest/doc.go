// Package pgtest holds the real PostgreSQL 18 integration tests for the goose
// migrations in db/migrations. All executable code lives in *_integration_test.go
// files behind the "integration" build tag; this file only anchors the package
// so the default "go test ./..." run has nothing of it to instrument.
//
// Run: go test -tags integration ./src/api/internal/platform/pgtest/...
// Requires the docker CLI; the tests start (and always remove) one throwaway
// container from a digest-pinned postgres:18 image.
package pgtest
