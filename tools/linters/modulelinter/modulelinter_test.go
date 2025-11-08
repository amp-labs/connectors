package modulelinter_test

import (
	"testing"

	"github.com/amp-labs/connectors/tools/linters/modulelinter"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestModuleLinter(t *testing.T) {
	t.Parallel()

	// Create the linter
	linter, err := modulelinter.New(modulelinter.Settings{})
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
