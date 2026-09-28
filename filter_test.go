package connectors

import (
	"errors"
	"testing"

	"github.com/amp-labs/connectors/common"
)

var errTestRejected = errors.New("rejected")

// testRawFilterConnector accepts only raw filters of type "soql".
type testRawFilterConnector struct {
	Connector
}

func (testRawFilterConnector) ValidateRawFilter(filter common.RawFilter) error {
	if filter.Type != "soql" {
		return errTestRejected
	}

	return nil
}

// testPlainConnector doesn't accept raw filters.
type testPlainConnector struct {
	Connector
}

// testWrapper wraps another connector, like a metrics wrapper does.
type testWrapper struct {
	Connector

	inner Connector
}

func (w testWrapper) Unwrap() Connector { return w.inner }

func TestNewRawFilterValidator(t *testing.T) {
	t.Parallel()

	soql := common.RawFilter{Type: "soql", Filter: "IsDeleted = true"}

	if err := NewRawFilterValidator(testRawFilterConnector{})(soql); err != nil {
		t.Fatalf("supported filter: got %v", err)
	}

	if err := NewRawFilterValidator(testWrapper{inner: testRawFilterConnector{}})(soql); err != nil {
		t.Fatalf("supported filter behind a wrapper: got %v", err)
	}

	err := NewRawFilterValidator(testRawFilterConnector{})(common.RawFilter{Type: "sql", Filter: "x"})
	if !errors.Is(err, errTestRejected) {
		t.Fatalf("filter the connector rejects: got %v, want the connector's error", err)
	}

	for name, conn := range map[string]Connector{
		"no raw filter support":       testPlainConnector{},
		"no support behind a wrapper": testWrapper{inner: testPlainConnector{}},
		"nil connector":               nil,
	} {
		if err := NewRawFilterValidator(conn)(soql); !errors.Is(err, common.ErrInvalidRawFilter) {
			t.Fatalf("%s: got %v, want ErrInvalidRawFilter", name, err)
		}
	}
}
