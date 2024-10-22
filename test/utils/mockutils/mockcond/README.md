# Description

This package is used by `mockserver` it provides common HTTP request checks.

## Key Features
* Validate HTTP request properties (method, headers, query parameters, body, etc.).
* Combine simple conditions into complex, nested conditional expressions.
* Extendable by implementing custom conditions.

# Usage

## Simple Conditions

Simple conditions check if an HTTP request meets specific criteria.
You can create your own custom conditions by implementing the `mockcond.Condition` interface.

### Examples

Match a specific URL path suffix:
```go
If:   mockcond.PathSuffix("/v2/tasks"),
```

Ensure the request method is POST:
```go
If:    mockcond.MethodPOST(),
```

Check for a query parameter with a specific value:
```go
If:    mockcond.QueryParam("startAt", "17"),
```

Ensure a query parameter is missing:
```go
If:    mockcond.QueryParamsMissing("updated_at[gte]"),
```

Check for a specific header:
```go
If:    mockcond.Header(testApiVersionHeader),
```

Match the exact request body content:
```go
If:    mockcond.Body(bodyRequest),
```

## Complex Conditions

You can combine conditions using logical operators like `And` & `Or`. 
These operators allow for nested conditions, enabling you to create complex validation expressions 
that resemble logical conditionals in code.


### Examples

Check for a header and ensure the method is DELETE:
```go
If: mockcond.And{
	mockcond.MethodDELETE(),
	mockcond.Header(testApiVersionHeader),
},
```
Match a specific URL path suffix for either GET or POST methods: 
```go
If: mockcond.And{
    mockcond.PathSuffix("/api/resource"),
    mockcond.Or{
        mockcond.MethodGET(),
        mockcond.MethodPOST(),
    },
},
```

## Custom Conditions
You can create custom conditions by implementing the `mockcond.Condition` interface to suit your specific requirements.
Here is a basic example of how to create a custom condition:
```go
type CustomCondition struct {}

func (c CustomCondition) EvaluateCondition(w http.ResponseWriter, r *http.Request) bool {
	// Implementation goes here. 
	return true
}
```