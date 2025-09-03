package custom

import (
	"context"
	"sync"
)

// concurrentProcessing executes the provided process function in parallel for each element in the list.
//
// All calls share the same context. If any invocation returns a non-nil error,
// the context is cancelled and the function returns immediately with that error.
//
// The process function receives a context, the element, and a mutex to safely
// synchronize access to shared resources (e.g., maps) across goroutines.
func concurrentProcessing[T any, P func(context.Context, T, *sync.Mutex) error](
	ctx context.Context, list []T, process P,
) error {
	// Create a cancellable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		group sync.WaitGroup
		mutex sync.Mutex
		errCh = make(chan error, 1)
	)

	for _, element := range list {
		group.Add(1)

		go func() {
			defer group.Done()

			// Return early if context is cancelled
			select {
			case <-ctx.Done():
				return
			default:
			}

			if err := process(ctx, element, &mutex); err != nil {
				// Attempt to send the first error only
				select {
				case errCh <- err:
					// Cancel context to stop other goroutines
					cancel()
				default:
				}
			}
		}()
	}

	// Wait for all goroutines to finish
	group.Wait()

	// Check if an error was reported
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
