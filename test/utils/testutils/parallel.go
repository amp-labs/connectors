package testutils

import (
	"context"

	"github.com/amp-labs/connectors/internal/simultaneously"
	"github.com/amp-labs/connectors/tools/fileconv"
)

const testLineBreak = "\n=============================================\n"

type ParallelRunner[C any] struct {
	FilePath  string
	TestTitle string
	Function  func(ctx context.Context, conn C, filePath string) (string, error)
}

type ParallelRunners[C any] []ParallelRunner[C]

// Run will execute test suites in parallel.
// Note: filePath must be neighbouring to the caller of this method.
func (r ParallelRunners[C]) Run(ctx context.Context, conn C) []string {
	logs := make([]string, len(r))

	callbacks := make([]simultaneously.Job, 0, len(r))

	for i, test := range r {
		idx := i    // capture loop variable
		tst := test // capture loop variable

		locator := fileconv.NewLevelFileLocator(fileconv.LevelParent)
		filePath := locator.AbsPathTo(tst.FilePath)

		callbacks = append(callbacks, func(ctx context.Context) error {
			logText, err := tst.Function(ctx, conn, filePath)
			if err != nil {
				logText = err.Error()
			}

			logs[idx] = formatLog(tst.TestTitle, logText)

			return nil
		})
	}

	if err := simultaneously.DoCtx(ctx, -1, callbacks...); err != nil {
		// Log error but continue - we want to collect all test results
		logs = append(logs, formatLog("Error", err.Error()))
	}

	return logs
}

func formatLog(title, logText string) string {
	return testLineBreak + title + testLineBreak + "\n" + logText
}
