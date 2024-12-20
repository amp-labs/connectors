package testroutines

import (
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/amp-labs/connectors/test/utils/testutils"
	"github.com/go-test/deep"
)

// TestCase describes major components that are used to test any Connector methods.
// It is universal and generic `Input` data type is what the tested method accepts,
// while `Output` value represents the data type of the expected output.
type TestCase[Input any, Output any] struct {
	// Name of the test suite.
	Name string
	// Input passed to the tested method.
	Input Input
	// Mock Server which connector will call.
	Server *httptest.Server
	// Custom Comparator of how expected output agrees with actual output.
	Comparator Comparator[Output]
	// Expected return value.
	Expected Output
	// ExpectedErrs is a list of errors that must be present in error output.
	ExpectedErrs []error
}

// Validate checks if supplied input conforms to the test intention.
func (o TestCase[Input, Output]) Validate(t *testing.T, err error, output Output) {
	defer o.Server.Close()

	// performs validation of error output using described test suite outline.
	o.checkError(t, err)
	// performs validation of data output using described test suite outline.
	o.checkValue(t, output)
}

func (o TestCase[Input, Output]) checkError(t *testing.T, err error) {
	testutils.CheckErrors(t, o.Name, o.ExpectedErrs, err)
}

func (o TestCase[Input, Output]) checkValue(t *testing.T, output Output) {
	// compare desired output
	var ok bool
	if o.Comparator == nil {
		// default comparison is concerned about all fields
		ok = reflect.DeepEqual(output, o.Expected)
	} else {
		ok = o.Comparator(o.Server.URL, output, o.Expected)
	}

	if !ok {
		diff := deep.Equal(output, o.Expected)
		t.Fatalf("%s:, \nexpected: (%v), \ngot: (%v), \ndiff: (%v)", o.Name, o.Expected, output, diff)
	}
}
