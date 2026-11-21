package nosleep_test

import (
	"testing"

	"github.com/amp-labs/connectors/tools/linters/nosleep"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestNoSleep(t *testing.T) {
	t.Parallel()

	linter, err := nosleep.New(nil)
	if err != nil {
		t.Fatalf("failed to create linter: %v", err)
	}

	analyzers, err := linter.BuildAnalyzers()
	if err != nil {
		t.Fatalf("failed to build analyzers: %v", err)
	}

	analysistest.Run(t, analysistest.TestData(), analyzers[0], "./connector")
}
