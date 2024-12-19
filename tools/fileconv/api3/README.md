# Package api3

## Purpose
This package extracts schemas metadata from OpenAPI files version 3.
If your OpenAPI file is using older version 2 check out `api2` package.

## Description
Some connectors cannot serve metadata via APIs and do this via static files.
Those files are a processed version of OpenAPI spec. 

## Loading File

```go
var (
    // Static file containing openapi spec.
    //
    //go:embed specs.json
    apiFile []byte

	FileManager = api3.NewOpenapiFileManager(apiFile) // nolint:gochecknoglobals
)
```

## Usage
Scripts that use this package are located under `scripts/openapi/<connector_name>/main.go`.

```go
// Pseudo code, omitting err.
schemas = yourconnector.FileManager.GetExplorer().GetBasicReadObjects(
    ignoreEndpoints,
    objectEndpoints,
    displayNameOverride,
    objectArrayLocator,
)
```

Argument description:

* **ignoreEndpoints** - list of URL paths. This way you can hard code which paths to skip when processing file.
  * Full path string: `/v1/order`
  * Any path that has suffix batch: `*/batch`
  * Any path that has prefix v2: `/v2/*`


* **objectEndpoints** - this is a mapping from URL path to the Object Name. 
  * By default, last URI part is used as Object Name.


* **objectArrayLocator** - function that accepts JSON response field and Object name. Using both you can determine 
if this is the correct response field that will hold your schema. Some common implementations are provided.
  * api3.IdenticalObjectLocator - expects data to be stored under the same name as object name. Ex: {"contacts":[...]}
  * api3.DataObjectLocator - expects schema to be returned under `data` field. Ex: {"data":[...]}
  * Your implementation can have exception or do combination of the two based on different objects.

Additionally, Explorer can be configured to apply display name processing after data is extracted. For example, you can capitalize every word of display for better look.
Edge cases should still be directly specified via **displayNameOverride** map.


### Configuration

The OpenAPI schema explorer can be configured to tailor the handling of edge cases. 
```go
openapi.FileManager.GetExplorer(
    api3.WithDisplayNamePostProcessors(
        api3.CamelCaseToSpaceSeparated,
        api3.CapitalizeFirstLetterEveryWord,
    )
    api3.WithParameterFilterGetMethod(
        api3.OnlyOptionalQueryParameters        		
    )
    api3.WithMediaType("application/vnd.api+json"),
    api3.WithPropertyFlattening(func(objectName, fieldName string) bool {
        // Nested attributes object holds most important fields.
        return fieldName == "attributes"
    }),
    api3.WithArrayItemAutoSelection(),
)
```
**Display Name.**
You can define a chained formatters of display name. For example: 
first, convert camel case into space-separated words; then, capitalize the first letters.
Of course, a single method will suffice, but chained processors allow for better composition of out-of-the-box utility methods.

**Parameter Filter.**
Some GET methods can be ignored based on the endpoint's input parameters. For example, retain endpoints that have exclusively optional query parameters.

**Media Type.**
By default, the API response is in `application/json`, but this can be configured as needed.

**Property Flattening.**
You can specify a field name for flattening, which will relocate nested fields to the top level.

**Array Item auto Selection.**
Enabling this flag allows for automatic selection of the object schema when the response contains a single array. If the response includes multiple arrays, the `objectArrayLocator` will still be invoked to resolve any ambiguity. 
