package testutils

import (
	"context"
	"github.com/amp-labs/connectors/tools/fileconv"
	"sync"
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

	var wg sync.WaitGroup
	for i, test := range r {
		wg.Add(1)

		locator := fileconv.NewLevelFileLocator(fileconv.LevelParent)
		filePath := locator.AbsPathTo(test.FilePath)

		go func(test ParallelRunner[C], idx int) {
			defer wg.Done()

			logText, err := test.Function(ctx, conn, filePath)
			if err != nil {
				logText = err.Error()
			}
			logs[idx] = formatLog(test.TestTitle, logText)
		}(test, i)
	}

	wg.Wait()

	return logs
}

func formatLog(title, logText string) string {
	return testLineBreak + title + testLineBreak + "\n" + logText
}
