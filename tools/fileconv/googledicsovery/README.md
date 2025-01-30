# Package googledicsovery

## Purpose
This package extracts schemas metadata from `Google Discovery` files.

## Description
The Google connector does not provide API endpoints for object metadata. 
Instead, **Google Discovery Files** describe operations and response schemas for each module.

Refer to:
* [Google API explorer](https://developers.google.com/apis-explorer/)
* [Google API Go Client](https://github.com/googleapis/google-api-go-client/) (for discovery files)

## Loading File

```go
var (
    // Static file containing google discovery file.
    //
    //go:embed discovery-file.json
    apiFile []byte

    FileManager = googledicsovery.NewFileManager(apiFile) // nolint:gochecknoglobals
)
```

## Usage

```go
// Pseudo code, omitting err.
schemas = packageWithDiscoveryFile.FileManager.GetExplorer().ReadObjectsGet(
    displayNameOverride,
)
```

Argument description:

The Explorer can be configured to process display names after extraction. For example, you can format names to improve readability.
Edge cases should be handled explicitly using the **displayNameOverride** map.


### Configuration

The Google discovery schema explorer can be configured to tailor the handling of edge cases. 
```go
packageWithDiscoveryFile.FileManager.GetExplorer(
    api3.WithDisplayNamePostProcessors(
        api3.CamelCaseToSpaceSeparated,
        api3.CapitalizeFirstLetterEveryWord,
    )
)
```

**Display Name.**
Display names can be transformed using chained formatters.
For example:
1. Convert camel case to space-separated words.
2. Capitalize the first letter of each word.
While a single formatter is sufficient, chaining allows better composition of built-in utilities.
