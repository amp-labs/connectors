package utils

import (
	"fmt"
	"time"
)

type CompletableJob interface {
	IsStatusDone() bool
}

// CycleUntilComplete will invoke producer function every interval.
// Once the producer gives non-nil value the cycle will come to an end. Either due to concrete error or result value.
func CycleUntilComplete[R CompletableJob](
	interval time.Duration,
	producer func() (R, error),
) (R, error) {
	defer func() {
		fmt.Println(".")
	}()

	for {
		fmt.Print(".")
		time.Sleep(interval)

		job, err := producer()
		if err != nil {
			var tmp R
			return tmp, err
		}

		if job.IsStatusDone() {
			return job, nil
		}
	}
}
