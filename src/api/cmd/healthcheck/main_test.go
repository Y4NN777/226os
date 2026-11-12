package main

import (
	"os"
	"testing"
)

// main exits only on failure, so the in-process path is the explicit "help" subcommand, which needs
// neither configuration nor network.
func TestMainPrintsUsageAndReturns(t *testing.T) {
	old := os.Args
	os.Args = []string{"healthcheck", "help"}
	t.Cleanup(func() { os.Args = old })

	main()
}
