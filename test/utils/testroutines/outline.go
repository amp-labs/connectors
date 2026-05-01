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

func (c TestCase[Input, Output]) Close() {
	c.Server.Close()
}

// Validate checks if supplied input conforms to the test intention.
func (c TestCase[Input, Output]) Validate(t *testing.T, err error, output Output) {
	// performs validation of error output using described test suite outline.
	c.checkError(t, err)
	// performs validation of data output using described test suite outline.
	c.checkValue(t, output)
}

func (c TestCase[Input, Output]) checkError(t *testing.T, err error) {
	testutils.CheckErrors(t, c.Name, c.ExpectedErrs, err)
}

func (c TestCase[Input, Output]) checkValue(t *testing.T, output Output) {
	// compare desired output
	var result *testutils.CompareResult
	if c.Comparator == nil {
		// default comparison is concerned about all fields
		result = c.defaultDeepCompare(output, c.Expected)
	} else {
		result = c.Comparator(c.Server.URL, output, c.Expected)
	}

	result.Validate(t, c.Name)
}

func (c TestCase[Input, Output]) defaultDeepCompare(actual, expected Output) *testutils.CompareResult {
	result := testutils.NewCompareResult()

	if !reflect.DeepEqual(actual, expected) {
		for _, diff := range deep.Equal(actual, expected) {
			result.AddDifference(diff)
		}
	}

	return result
}
