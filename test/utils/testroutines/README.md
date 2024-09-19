# Package testroutines

## Purpose

This package provides base test case for creating `Test Suites`.

## Description

Most common connector methods can be tested using:

* testroutines.Read - Read
* testroutines.Write - Write
* testroutines.Metadata - ListObjectMetadata
* testroutines.Delete - Delete

They can be used as a template to declare your unique test case type.
The main difference among them is

* input type
* output type
* tested method

Below is the example of the run method:

```go
// Declaration
func (r Read) Run(t *testing.T, builder ConnectorBuilder[connectors.ReadConnector]) {
t.Helper()
conn := builder.Build(t, r.Name) // builder will return Connector of certain type
output, err := conn.Read(context.Background(), r.Input) // select a method that you want to test and pass input
ReadType(r).Validate(t, err, output) // invoke TestCase[InputType,OutputType].Validate(...)
}

// Example calling method
tt.Run(t, func () (connectors.ReadConnector, error) {
return constructTestConnector(tt.Server.URL)
})
```

### TestCase

A **Test case** consists of:

* `InputType`: captures all inputs to the tested method. Inside the `Run` method,
  it is wired and passed as arguments required by the method.
* `OutputType`: represents the result of a successful method execution, which is then compared
  against `TestCase.Expected` using `TestCase.Comparator`for equality check (or deep equal if none specified).

A test case can also include `TestCase.ExpectedErrs`,
which ensures that all expected errors are present in the returned error (checked as a subset, not strict equality).

```go
type Read TestCase[common.ReadParams, *common.ReadResult]

func TestRead(t *testing.T) {
  t.Parallel()
  
  // Common test setup
  // ...

  // Suite definition
  tests := []testroutines.Read{
    {
      Name:         "Title of the test",
      Input:        &common.ReadParams{} // This object represents `InputType`
      Server:       mockserver.Dummy(),  // Configure mock server to respond on requests.
      Comparator:   func (baseURL string, actual, expected *common.ReadResult) bool {
        return true // Custom function to compare expected vs given `OutputType`.
      },
      ExpectedErrs: []error{common.ErrMissingObjects}, // List of expected errors to be inside error object 
      Expected:     &common.ReadResult{} // This object represents `OutputType`
    },
    
    // ... other test cases ...

  }
  
  // Running tests in the loop
  // ...
}
```