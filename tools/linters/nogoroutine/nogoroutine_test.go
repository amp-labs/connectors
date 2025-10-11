package nogoroutine_test

import (
	"testing"

	"github.com/amp-labs/connectors/tools/linters/nogoroutine"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoGoroutine(t *testing.T) {
	t.Parallel()

	// Create the linter
	linter, err := nogoroutine.New(nogoroutine.Settings{
		ExcludePaths: []string{
			"internal/future",
			"internal/simultaneously",
		},
	})
	if err != nil {
		t.Fatalf("failed to create linter: %v", err)
	}

	// Get the analyzers
	analyzers, err := linter.BuildAnalyzers()
	if err != nil {
		t.Fatalf("failed to build analyzers: %v", err)
	}

	// Run the test
	analysistest.Run(t, analysistest.TestData(), analyzers[0], "testdata")
}
