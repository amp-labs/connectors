package connector

import (
	"context"
	"time"

	ampcommon "github.com/amp-labs/amp-common/utils"
)

func badExample(ctx context.Context) {
	time.Sleep(time.Second) // want `Sleeping in connector code is forbidden \(time.Sleep\)`

	_ = ampcommon.SleepCtx(ctx, time.Second) // want `Sleeping in connector code is forbidden \(utils.SleepCtx\)`
}

type utils struct{}

func (utils) SleepCtx(ctx context.Context, dur time.Duration) error { return nil }

func Sleep(dur time.Duration) {}

func goodExample(ctx context.Context) {
	// A same-named function or method outside the forbidden packages is allowed.
	Sleep(time.Second)
	_ = utils{}.SleepCtx(ctx, time.Second)

	// Other functions from the forbidden packages are allowed.
	_ = time.Now().Add(time.Second)
	timer := time.NewTimer(time.Second)
	timer.Stop()
}
