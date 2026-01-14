# Package api2

## Purpose
This package extracts schemas metadata from OpenAPI files version 2.
If your OpenAPI file is using newer version 3 check out `api3` package.

## Description

Before extracting metadata from OpenAPI, the schema is initially converted from V2 to V3 format.
For further details, refer to the `api3` package README, as the remaining process follows the same approach as `api3`.
This package acts as a gateway and a wrapper of `api3`.

## Loading File

```go
var (
    // Static file containing openapi spec.
    //
    //go:embed specs.json
    apiFile []byte

	FileManager = api2.NewOpenapiFileManager(apiFile) // nolint:gochecknoglobals
)
```
