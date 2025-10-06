# Description

This folder contains utility methods to mock API calls made by connectors to provider APIs. 
It offers several types of mock servers to simulate different scenarios by evaluating incoming 
requests and returning appropriate responses.

## Mock Servers
The package provides multiple mock server types:
* **Dummy**: always returns a predefined status.
* **Fixed**: returns a fixed response regardless of the request.
* **Conditional**: (equivalent to `if`) returns responses based on request conditions.
* **Switch**: matches requests against multiple conditions and selects the first matching response.

# Without this package

A typical mock server setup without using this package would look like:
```go
Server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
})),
```
For more complex connectors, this approach becomes unwieldy. 
A connector may make multiple API calls, each requiring specific request conditions to be met.
This package simplifies such scenarios.

# Usage Examples

## Dummy Mock Server

This mock server always returns StatusTeapot. It’s useful for tests that don’t yet need to handle API calls.
```go
Server: mockserver.Dummy()
```

## Fixed Mock Server

This server will always return a 200 status code with a predefined response (in this case, customers).
It’s ideal for simple tests that only need a static response.
```go
Server: mockserver.Fixed{
	Setup:  mockserver.ContentJSON(),
	Always: mockserver.Response(http.StatusOK, customers),
}.Server(),
```

## Conditional Mock Server

This mock server returns a 200 status code along with response data (e.g., indicating a job was created) 
only if the conditions are met. The conditions in this example require that both the URL path and the request 
body match expected values. If the conditions are not met, an optional Else clause can specify a default response 
(otherwise the server defaults to a 500 status with a failure message).
This is highly useful to ensure that requests are constructed properly by the connector with respect to provider APIs.
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

## Switch Mock Server

This mock server behaves like a Go switch statement, evaluating multiple cases and returning the response 
for the first matching condition. In this example, it selects the response based on the request URL path, 
with a default response if no cases match.
This is particularly useful when a connector makes multiple API calls, each requiring a different response.

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
