package testdata

import "context"

func badExample() {
	// This should trigger the linter
	go func() { // want "Direct use of 'go' keyword is forbidden"
		println("bad")
	}()

	// This should also trigger
	go someFunction() // want "Direct use of 'go' keyword is forbidden"
}

func someFunction() {
	println("test")
}

func goodExample() {
	// These would be the correct approaches (not tested here as they require imports)
	// future.GoContext(ctx, func(ctx context.Context) (int, error) { return 42, nil })
	// simultaneously.DoCtx(ctx, 1, func(ctx context.Context) error { return nil })
}
