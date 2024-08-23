# Package api3

## Purpose
This package extracts schemas metadata from OpenAPI files.

## Description
Some connectors cannot serve metadata via APIs and do this via static files.
Those files are a processed version of OpenAPI spec. 

## Usage
Scripts that use this package are located under `scripts/openapi/<connector_name>/main.go`.

```go
// Pseudo code, omitting err.
schemas = api3.NewOpenapiFileManager(openapiBytesDataFile).GetExplorer().GetBasicReadObjects(
    ignoreEndpoints,
    objectEndpoints,
    displayNameOverride,
    objectCheck,
)
```

Argument description:

* **ignoreEndpoints** - list of URL paths. This way you can hard code which paths to skip when processing file.
  * Full path string: `/v1/order`
  * Any path that has suffix batch: `*/batch`
  * Any path that has prefix v2: `/v2/*`


* **objectEndpoints** - this is a mapping from URL path to the Object Name. 
  * By default, last URI part is used as Object Name.


* **objectCheck** - function that accepts JSON response field and Object name. Using both you can determine 
if this is the correct response field that will hold your schema. Some common implementations are provided.
  * api3.IdenticalObjectCheck - expects data to be stored under the same name as object name. Ex: {"contacts":[...]}
  * api3.DataObjectCheck - expects schema to be returned under `data` field. Ex: {"data":[...]}
  * Your implementation can have exception or do combination of the two based on different objects.

Additionally, Explorer can be configured to apply display name processing after data is extracted. For example, you can capitalize every word of display for better look.
Edge cases should still be directly specified via **displayNameOverride** map.
