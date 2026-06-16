package parallelfetch

import (
	"context"
	"errors"
	"testing"
)

func TestExecute(t *testing.T) {
	t.Parallel()

	maxConcurrent := -1

	t.Run("Empty tasks", func(t *testing.T) {
		t.Parallel()

		result := Execute[string, int](context.Background(), nil, maxConcurrent)
		if len(result.Records) != 0 {
			t.Errorf("expected 0 records, got %d", len(result.Records))
		}
		if len(result.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(result.Errors))
		}
	})

	t.Run("All success", func(t *testing.T) {
		t.Parallel()

		tasks := []Task[string, int]{
			func(ctx context.Context) (string, *int, error) {
				return "task1", new(10), nil
			},
			func(ctx context.Context) (string, *int, error) {
				return "task2", new(20), nil
			},
		}

		result := Execute(context.Background(), tasks, maxConcurrent)

		if len(result.Records) != 2 {
			t.Errorf("expected 2 records, got %d", len(result.Records))
		}
		if result.Records["task1"] != 10 {
			t.Errorf("expected 10 for task1, got %d", result.Records["task1"])
		}
		if result.Records["task2"] != 20 {
			t.Errorf("expected 20 for task2, got %d", result.Records["task2"])
		}
		if len(result.Errors) != 0 {
			t.Errorf("expected 0 errors, got %d", len(result.Errors))
		}
	})

	t.Run("All failures", func(t *testing.T) {
		t.Parallel()

		errFail := errors.New("fail")
		tasks := []Task[string, int]{
			func(ctx context.Context) (string, *int, error) {
				return "task1", nil, errFail
			},
			func(ctx context.Context) (string, *int, error) {
				return "task2", nil, errFail
			},
		}

		result := Execute(context.Background(), tasks, maxConcurrent)

		if len(result.Records) != 0 {
			t.Errorf("expected 0 records, got %d", len(result.Records))
		}
		if len(result.Errors) != 2 {
			t.Errorf("expected 2 errors, got %d", len(result.Errors))
		}
		if !errors.Is(errFail, result.Errors["task1"]) {
			t.Errorf("expected error for task1")
		}
	})

	t.Run("Mixed success and failure", func(t *testing.T) {
		t.Parallel()

		errFail := errors.New("fail")
		tasks := []Task[string, int]{
			func(ctx context.Context) (string, *int, error) {
				return "success", new(10), nil
			},
			func(ctx context.Context) (string, *int, error) {
				return "failure", nil, errFail
			},
		}

		result := Execute(context.Background(), tasks, maxConcurrent)

		if len(result.Records) != 1 {
			t.Errorf("expected 1 record, got %d", len(result.Records))
		}
		if result.Records["success"] != 10 {
			t.Errorf("expected 10 for success task")
		}
		if len(result.Errors) != 1 {
			t.Errorf("expected 1 error, got %d", len(result.Errors))
		}
		if !errors.Is(errFail, result.Errors["failure"]) {
			t.Errorf("expected error for failure task")
		}
	})
}
