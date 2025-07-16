# Package testscenario

**`testscenario`** provides reusable, configurable procedures for **live testing** connector operations.

## Purpose

This package defines common test scenarios for connectors that make real API calls. It helps ensure that different connector implementations correctly handle key combinations of high-level interfaces, like **Create**, **Read**, **Update**, **Delete**, and **Object Metadata**.

# Features

| Method                             | Testing Target                                          | Primary Focus & Scenario                                        |
|------------------------------------|---------------------------------------------------------|-----------------------------------------------------------------|
| ValidateCreateDelete               | ReadConnector, WriteConnector (Create), DeleteConnector | Create a record, verify it can be found, then delete it         |
| ValidateCreateUpdateDelete         | ReadConnector, WriteConnector, DeleteConnector          | Full CRUD: create, update, verify update, and delete            |
| ValidateMetadataExactlyMatchesRead | ObjectMetadataConnector, ReadConnector                  | Validate that metadata **exactly** matches Read output          |
| ValidateMetadataContainsRead       | ObjectMetadataConnector, ReadConnector                  | Validate that metadata **includes all** fields returned by Read |

# Usage

## CRUD

**`ValidateCreateDelete`** and **`ValidateCreateUpdateDelete`** use `CRDTestSuite` and `CRUDTestSuite` respectively to control scenario configuration.

**Key parameters:**
- `ReadFields` — required. Fields that must be verified in Read/Update.
- `WaitBeforeSearch` — optional. Add a delay if the backend needs time to reflect changes.
- `SearchBy` — optional. Lookup a record by a unique property instead of the primary ID. Good for validating create.
- `RecordIdentifierKey` — required. The field name for the record ID, primary key.
- `UpdatedFields` — for update scenarios: checks that specific fields have the expected new values.
- `PreprocessUpdatePayload` — optional. Modify the update payload dynamically using the result from Create.
- `ValidateUpdatedFields` — optional. Provide custom logic to verify updates instead of static `UpdatedFields`.

```go
testscenario.ValidateCreateUpdateDelete(ctx, conn,
    "<object_name>",
    createPayload{},
    updatePayload{},
    testscenario.CRUDTestSuite{
        ReadFields: datautils.NewSet("id", "name"),
        WaitBeforeSearch: time.Second,
        SearchBy: testscenario.Property{
            Key:   "<fieldName>",
            Value: fieldValue,
        },
        RecordIdentifierKey: "id", 
        PreprocessUpdatePayload: func(createResult *common.WriteResult, updatePayload any) {},
        UpdatedFields: map[string]string{
            "name": updatedName,
        },
        ValidateUpdatedFields: func(record map[string]any){}
    },
)
```

## Metadata

**`ValidateMetadataExactlyMatchesRead`** and **`ValidateMetadataContainsRead`** validate that the connector's declared metadata is in sync with it's Read behavior.
** Use **`ValidateMetadataExactlyMatchesRead`** if your connector always returns the same complete set of fields for an object. This test ensures ListObjectMetadata describes exactly what Read returns — no extra, no missing.
** Use **`ValidateMetadataContainsRead`** if your connector may omit fields in Read output (e.g., optional or empty fields). This test ensures that every field returned by Read is known to ListObjectMetadata — extra declared fields are fine, but undocumented fields are not.

```go
testscenario.ValidateMetadataExactlyMatchesRead(ctx, conn, "users")
testscenario.ValidateMetadataContainsRead(ctx, conn, "Organization", sanitizeReadResponse)
```
Tip: Use the `responsePostProcess` argument in `ValidateMetadataContainsRead` to adjust or filter the raw Read response before comparison.
