package mock

import "testing"

func TestDefaultNewConnector(t *testing.T) {
	t.Parallel()

	// It should be possible to construct a new mock connector without any args.
	conn, err := NewConnector()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if conn == nil {
		t.Fatal("expected a connector instance, got nil")
	}
}
