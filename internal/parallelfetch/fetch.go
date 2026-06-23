package parallelfetch

import (
	"context"

	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/simultaneously"
)

// Result holds the outcome of executing multiple generic tasks, containing
// successful records and any errors that occurred, keyed by their task ID.
//
// The type parameters are:
//   - ID: The task ID type, must be comparable (usable as map key)
//   - DATA: The data type returned by successful tasks.
type Result[ID comparable, DATA any] struct {
	// Records contains successfully completed task data, mapped by task ID.
	Records datautils.Map[ID, DATA]
	// Errors contains any errors that occurred during task execution,
	// mapped by the task ID that failed.
	Errors datautils.Map[ID, error]
}

// pair holds a key-value pair for channel communication during task execution.
// This is a private helper type used internally by Execute.
type pair[ID comparable, DATA any] struct {
	// ID is the task identifier.
	ID ID
	// Data holds either the task result or error.
	Data DATA
}

// Task is a generic function type representing an asynchronous task that
// executes in a context and returns a task ID, data pointer, and optional error.
//
// The type parameters are:
//   - T: The task ID type, must be comparable
//   - D: The data type returned by the task
type Task[T comparable, D any] func(ctx context.Context) (taskID T, data *D, err error)

// Execute concurrently runs all provided tasks in parallel and collects their
// results into a Result struct containing both successful records and errors.
//
// This function executes tasks concurrently using the simultaneously package,
// communicating results via buffered channels. Errors are not returned from
// the internal callbacks; instead, success and failure data are sent through
// separate channels and aggregated into the returned Result.
//
// Parameters:
//   - ctx: The execution context for cancellation and timeouts
//   - tasks: A slice of Task functions to execute concurrently
//   - maxConcurrent: Number of concurrent tasks that run at a time
//
// Returns:
//   - Result[T, D]: Contains Records for successful tasks and Errors for failed ones, both keyed by task ID
//
// Example:
//
//	tasks := []Task[string, MyData]{
//	    func(ctx context.Context) (string, *MyData, error) {
//	        return "task1", &MyData{...}, nil
//	    },
//	}
//	result := Execute(ctx, tasks)
//	// result.Records["task1"] contains the data
//	// result.Errors["task1"] contains any failures
func Execute[ID comparable, DATA any](ctx context.Context, tasks []Task[ID, DATA], maxConcurrent int) Result[ID, DATA] {
	var (
		numTasks        = len(tasks)
		responseChannel = make(chan pair[ID, DATA], numTasks)
		errChannel      = make(chan pair[ID, error], numTasks)
		result          = Result[ID, DATA]{
			Records: make(map[ID]DATA),
			Errors:  make(map[ID]error),
		}
	)

	if numTasks == 0 {
		return result
	}

	// Build callback functions for each task that wrap the task execution
	// and send results to the appropriate channel.
	callbacks := make([]simultaneously.Job, numTasks)
	for index, task := range tasks {
		callbacks[index] = func(ctx context.Context) error {
			taskID, body, err := task(ctx)
			if err != nil {
				// Task failed: send error to error channel
				errChannel <- pair[ID, error]{
					ID:   taskID,
					Data: err,
				}

				return nil // nolint:nilerr
			}

			// Task succeeded with data: send to response channel
			responseChannel <- pair[ID, DATA]{
				ID:   taskID,
				Data: *body,
			}

			// No errors returned here; communication is done via channels
			// to allow Execute to aggregate all results centrally.
			return nil
		}
	}

	// Execute all jobs concurrently.
	// Blocks until all callbacks complete.
	_ = simultaneously.DoCtx(ctx, maxConcurrent, callbacks...)

	// All jobs are complete, so safely close both channels.
	// This unblocks the range loops below.
	close(responseChannel)
	close(errChannel)

	// Collect successful results from the response channel.
	for data := range responseChannel {
		result.Records[data.ID] = data.Data
	}

	// Collect errors from the error channel.
	for data := range errChannel {
		result.Errors[data.ID] = data.Data
	}

	return result
}
