# Description

This package is used by `mockserver` it provides common HTTP request checks.

# Conditions

## Simple

These conditions ask if HTTP request matches a criteria.
You can create custom condition by implementing `mockcond.Condition` interface.

### Examples

Request URL path must match a string.
```go
If:   mockcond.PathSuffix("/v2/tasks"),
```

Request method must be POST.
```go
If:    mockcond.MethodPOST(),
```

Request must have `startAt` query parameter with value `"17"`.
```go
If:    mockcond.QueryParam("startAt", "17"),
```

Request must NOT have query parameter.
```go
If:    mockcond.QueryParamsMissing("updated_at[gte]"),
```

Request must have header.
```go
If:    mockcond.Header(testApiVersionHeader),
```

Request body data must be exact as expected text.
```go
If:    mockcond.Body(bodyRequest),
```

## Complex

Conditions can be joined and nested using `and` with `or` clauses. 
They themselves act as a condition, and therefore they can be stacked in multiple levels
resembling `()` brackets in conditionals. Ex: (cond1 and (cond2 or cond3))


### Examples

Request must have header and be a DELETE method.
```go
If: mockcond.And{
	mockcond.MethodDELETE(),
	mockcond.Header(testApiVersionHeader),
},
```
