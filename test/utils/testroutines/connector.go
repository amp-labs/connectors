// Package testroutines holds a collection of common test procedures.
// They provide a framework to write mock tests.
package testroutines

import (
	"testing"
)

// ConnectorBuilder is a callback method to construct and configure connector for testing.
// This is a factory method called for every test suite.
type ConnectorBuilder[Conn any] func() (Conn, error)

func (builder ConnectorBuilder[C]) Build(t *testing.T, testCaseName string) C {
	conn, err := builder()
	if err != nil {
		t.Fatalf("%s: error in test while constructing connector %v", testCaseName, err)
	}

	return conn
}
