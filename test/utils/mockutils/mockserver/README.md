# Description

This folder contains utility methods used to mock API calls made by connector to provider APIs.
Several common mock server structures are provided to check the request and select the desired response.

There are multiple `mockservers`: 
* Dummy
* Fixed
* Conditional(equivalent to If)
* Switch

# Mocks without this package

A mock server without using this package would look as follows:
```go
Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
})),
```
This logic can become more complex for sophisticated connectors, where some may make multiple API calls, 
many request conditions to be met. Common use cases are summarized below.

# Examples

## Dummy

```go
Server: mockserver.Dummy()
```
Mockserver will consistently return `StatusTeapot`.
Consider this when the test doesn't yet require handling API calls.

## Fixed

```go
Server: mockserver.Fixed{
	Setup:  mockserver.ContentJSON(),
	Always: mockserver.Response(http.StatusOK, customers),
}.Server(),
```
Mockserver will always return customers data as response body with status 200.
This is the simplest server, focused on returning a status code and/or some dummy data.

## Conditional

```go
Server: mockserver.Conditional{
	Setup: mockserver.ContentJSON(),
	If: mockcond.And{
		mockcond.PathSuffix("/services/data/v59.0/jobs/ingest"),
		mockcond.Body(bodyRequest),
	},
	Then: mockserver.Response(http.StatusOK, responseCreateJob),
    Else: mockserver.Response(http.StatusOK, []byte{})
}.Server(),
```
Mockserver will return a status of 200 along with response data to indicate that a job was created, 
but only if a condition is met. The condition is nested: both the request URL path suffix
and the request body data must match the expected values for the condition to evaluate to true.
Note: The else case is optional. If omitted, it defaults to a status of 500 with a JSON message, causing the test to fail.

This is the most useful mock server. The test ensures that the deep connector builds requests according 
to our expectations by providing a list of checks that the HTTP request must satisfy.

## Switch

```go
Server: mockserver.Switch{
	Setup: mockserver.ContentJSON(),
	Cases: []mockserver.Case{{
		If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='account')"),
		Then: mockserver.Response(http.StatusOK, responseContactsSchema),
	}, {
		If:   mockcond.PathSuffix("EntityDefinitions(LogicalName='account')/Attributes"),
		Then: mockserver.ResponseString(http.StatusOK, `{"value":[]}`),
	}},
	Default: mockserver.Response(http.StatusOK, []byte{}),
}.Server(),
```
Mockserver behaves differently based on the request bodies. In the example above,
it makes its judgment based on the request URL path. The first matching condition will trigger the then clause.
The check progresses from the first case to the last, resembling a Go switch statement.

Consider using this when the connector makes multiple API calls, and you want to respond differently to each.
The default can serve as a safeguard, returning a response that will ensure the connector fails the test.
