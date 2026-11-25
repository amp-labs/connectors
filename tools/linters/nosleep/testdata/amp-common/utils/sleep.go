// Package utils is a stub of github.com/amp-labs/amp-common/utils for analysistest.
package utils

import (
	"context"
	"time"
)

// SleepCtx stubs amp-common's utils.SleepCtx. It gives the analysistest a package at the real
// path github.com/amp-labs/amp-common/utils, so the test can check that the call is flagged
// under an aliased import without depending on the real amp-common module.
func SleepCtx(ctx context.Context, dur time.Duration) error {
	return nil
}
