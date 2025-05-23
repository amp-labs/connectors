### Context

Ampersand is a platform for building  **connectors**  to third-party SaaS APIs (e.g., Salesforce, HubSpot, Zendesk, etc.). These connectors enable read/write operations and metadata retrieval from external provider APIs within our system. There are different types of connectors in this project:

- **Provider**: Each third party API that Ampersand wants to build a connector against is called a provider.

- **Catalog**: All providers define their own configuration under `providers/<provider>.go` file. As an example, Salesforce defines it configuration under the `providers/salesforce.go` file. The entire set of provider configurations is called as the provider catalog.

- **ProviderInfo**: Each provider's configuration is called it's Provider Info and defines the name, support, authentication scheme, media URLs (to display in the front end UI), any metadata that the connector will need to collect in order to build a request for reading, writing or proxying to the provider. This is critical and must be dealt with very carefully, because all Ampersand components are based on this. Whenever we have metadata that will be injected into the ProviderInfo, we represent it using double curly braces. For example, Salesforce ProviderInfo.BaseURL needs a workspace, and we collect the workspace as input from the user, so we designate it as `{{.workspace}}` in the `providers/salesforce.go`. This is then substituted when we read the ProviderInfo by a helper method which we talk about below in `Reading ProviderInfo for a provider`.

- **Catalog types:** All the catalog types are stored under `providers/types.gen.go` and are generated from https://github.com/amp-labs/openapi/blob/main/catalog/catalog.yaml. When you need to understand the structure of any ProviderInfo, this is the source of truth.

- **Reading ProviderInfo for a provider:** We've defined some utility functions in `providers/utils.go` with functions like `ReadInfo` which accepts the provider, and any metadata variables or substitutions that needs to be made and uses go's `template` library (https://pkg.go.dev/text/template) to do this. For example, Salesforce's base URL in its ProviderInfo has `{{.workspace}}` and when it receives a map for `{"workspace": "test-workspace"}`, the function will substitute the value in the ProviderInfo and we get the complete base URL against which we can now make valid requests.

-   **Proxy Connectors**  (also sometimes called as `auth connectors`): These are basic connectors that simply proxy requests to the provider’s API. They handle authentication and forward requests/responses (with per-request logging and observability) but do not contain custom business logic. A proxy connector mainly needs the provider’s base URL and auth configuration, and is typically defined under the  `providers/`  directory in the `connectors` repository. All proxy connectors reuse a generic proxy implementation that is stored in the `connector` folder by injecting the specific provider info.

-   **Deep Connectors**: These are deeper connectors that implement provider-specific logic for reading data, writing data, and exposing object metadata. Deep connectors handle the quirks of each API – pagination, rate limits, field mappings, non-standard error responses, etc. They reside in their own package (e.g.,  `connectors/salesforce`,  `connectors/hubspot`) separate from the generic proxy connectors. Deep connectors support various interfaces:  **ReadConnector**  (retrieving records, always with pagination, incremental reading (supporting reading from a timestamp) and filtering unless not available),  **Write**  (creating/updating records), and  **ListObjectMetadata**  (providing schema information for objects). Each piece of functionality can be added independently, but a “deep” connector generally aims to support as much as possible (reads, writes, metadata, incremental sync, etc.) for its target provider.

- **Connector interfaces**: All connectors MUST satisfy their respective interfaces which are all stored under `connector.go` in the root directory. So a proxy connector must satisfy the base `Connector` interface, a deep connector which reads MUST satisfy the `ReadConnector` interface, and the same goes for write (`WriteConnector`), delete (`DeleteConnector`), metadata (`ObjectMetadata` and so on.

- **Order of defining a connector:** Usually starts with defining a proxy connector, then adding a package for it under `providers` (eg. `providers/salesforce`), then satisfying the `ObjectMetadata` interface by defining a `Connector` struct and a `ListObjectMetadata` method, then `Read` and `Write`.

- **How to define a deep connector**: Earlier connectors defined their own structure like Salesforce, but that led to a lot of repeated code. For example, salesforce (`providers/salesforce`) would define options as `salesforce.Option` which made it harder to generalize across connectors, or had to define a lot of HTTP request building/parsing in it's read or write methods. To tackle this, we've introduced a `internal/components` package which abstracts away the connector's base implementation. All new connectors here onwards MUST embed and use these components which reduces the lines of code and creates a predictable structure for all connectors. For example, `providers/aha` implements this by embedding the `*components.Connector`, then defines that it requires `AuthenticatedClient` and `Workspace` as inputs to `NewConnector` for proper functioning. When we want to make this a ObjectMetadata connector, we add `components.SchemaProvider` interface to it. Only embed `components.Reader` interface when actually implementing a reader. Do not embed `components.Reader/Writer/SchemaProvider` if you are not implementing them, because that will return errors later if a user tries to read or write using the connector. All in all, the `internal/components` package is the source of truth for deep connectors, and we must try to adopt it everywhere.


### Naming Conventions

-   **Connector Names**: Name connector packages and files after the provider service, using camelcase (e.g.,  `salesforce`,  `google`, `adobeExperience`). Proxy connectors for a provider reuse the `connector` folder, while deep connectors have their own top-level folder under  `connectors/`. Use the provider’s official name (or a reasonable abbreviation if necessary) for naming. Make sure that abbreviations are capitalized (eg. `ironcladEU`, not `ironcladEu`)

-   **Object Names**: Use consistent object names across all operations (read/write/metadata). If a provider has a resource, always try to preserve the same name because that is what our users of connectors will define, and we want to avoid maintaining a mapping of object names. For example, Salesforce contact object can be read from `https://{subdomain}.my.salesforce.com/services/data/v63.0/sobjects/contact/{contactId}`, so the object name is `contact`. If the object name contains slashes, maintain that as well. For example, if stripe has `billing/alerts`, maintain that and don't remap it. In rare cases, if the provider’s API uses different endpoints or names for the same resource (e.g.,  `jobs.list`  vs  `jobs.create`  for reading vs writing jobs), choose a single canonical name (e.g., `jobs`) for that object in our connector. This ensures the same object key is used for reading, writing, etc., in our system. Do not invent new names for objects that the provider already defines — stick to the provider’s terminology as closely as possible (unless consolidation is needed as above). This is to ensure that users of our connectors can simply put in a resource name, and expect to proxy, read, write or inspect the object schema automatically.

-   **Field Names**: Preserve the provider’s field names exactly in our outputs and schema whenever possible. Avoid renaming fields to something custom. For example, if the provider’s API returns a field  `created_at`, our connector should use  `created_at`  (not  `createdDate`  or  `timestamp`) in the data it exposes. This consistency helps users recognize fields and eases maintenance.

-   **Commit & PR Conventions**: When adding or updating connectors, follow our commit message format for clarity. For example, a new connector commit/PR title might be  **`feat: Add XYZ [functionality] Connector`**  (using the provider name in Title Case), where `[functionality]` is read, write, metadata, delete, etc. In case a fix is being made or a minor feature is being added to the connector, clearly mention the connector and the feature/fix, e.g.,  _“[ConnectorName] Add support for incremental sync for <object>”_ where `<object>` is a resource in the API.


### Schema & Metadata (ListObjectMetadata)

-   Implement the  **ListObjectMetadata**  interface for deep connectors to provide schema information for each object (resource) the connector supports. This typically returns a set of fields (and their details) for a given object type. This is very important as it powers all the UI components of Ampersand where in users can select fields that they want to read from a given object, or interact with in other ways.

-   **Metadata Sources**: Whenever possible, retrieve object schema directly from the provider’s API (for example, some APIs have a describe or metadata endpoint for objects. Hubspot has a describe API: https://developers.hubspot.com/docs/reference/api/crm/objects/schemas). This is the preferred source as it’s most precise and up-to-date. If the provider offers metadata discovery endpoints, always use them.

    -   If no API metadata is available (or it’s incomplete), use a  **static schema definition**  (e.g., an OpenAPI/JSON schema file in the connector) as a fallback. Keep this in sync with provider docs, and update it if the API changes.

    -   Another technique is  **schema inference**  by sampling a read response (i.e. parsing a record to infer fields), but this can be unreliable (e.g., empty responses or varying fields). Use inference only as a last resort or to complement other methods.

-   **Metadata Format**: Our platform supports two versions of metadata output:

    -   _Version 1_: A simple map of field API names to display names (deprecated, kept for backward compatibility).

    -   _Version 2_: A rich map of field names to a structure with detailed field properties (data type, possible values, etc.). Always aim to populate the richer  **FieldMetadata**  (V2) for each field, and use provided helpers (e.g.,  `common.NewObjectMetadata()`) to derive V1 from it for legacy support. This includes capturing picklist values (enumerations), data types, whether a field is required or read-only, etc., if the provider provides that info.

-   **Custom Objects & Fields**: If the provider allows user-defined custom objects or custom fields, plan to support that (usually as an extended feature). Often, handling custom objects/fields is complex and done in a separate PR. In such cases, ensure the base connector is designed to accommodate adding custom fields later. For custom fields that appear as identifiers (e.g. field IDs or GUIDs), try to resolve or fetch their human-readable names when presenting metadata or data to the user. Always keep the raw ID in the data (if relevant) but provide a name where possible for usability. Document any custom field handling clearly. As an example, look at `providers/zendesksupport/customFields.go`


### Data Structure –  **Fields**  vs  **Raw**  in Read Results

-   Every read operation (fetching data from the provider) should return results that include two representations of each record:

    -   **Fields**: a processed, flattened representation of the record’s key fields for easy consumption. Our users generally define a list of fields that they are interested in, and we extract them into `Fields` for easier access. This is typically a  `map[string]interface{}`  or similar, containing the main data fields. The connector should remove unnecessary nesting or wrappers that the API might have. For example, if an API wraps all record fields inside an  `{"attrs": { ... }}`  object, our connector’s Fields should pull those out to the top level. But getting rid of unnecessary nesting is the only change we should make, and never touch the actual data. For example, if there is an `address` key which has `city`, `state`, and other fields in it, never flatten that because the user will expect it in a format consistent with the provider's API. The guiding principle is to make Fields as straightforward as possible  **but without losing structure**  that matters. Use provider field names as keys in `Fields` (preserve exact casing/spelling the API uses).

    -   **Raw**: the complete  **unmodified JSON**  (or data structure) exactly as returned from the provider’s API. This should be included for each record (under the key  `raw`  or similar) and should not be altered in any way. Raw is important for debugging and for accessing any data that isn’t flattened into Fields.  **Do not**  modify or omit anything in the raw payload – it must remain a faithful copy of the provider’s response for that record. This is so that users can always fallback to `raw` even if we make a mistake in populating `Fields`.

-   **Consistency between Read and Write**: In general, the principle to follow is that the data shape used in `ReadResult.Fields` should be accepted as input for write operations (`WriteParams.Record`). In other words, a builder should be able to take a  `ReadResult.Fields`  and use it (perhaps with small modifications) as the basis for a  `WriteRequest`. The connector is responsible for translating this user-friendly flat structure back into the provider’s expected format. For example, if the provider’s API expects the payload under an  `"attrs"`  key as in the earlier example, our connector should internally wrap the flat Fields data under  `"attrs"`  when sending a create/update request. This ensures a smooth experience: users don’t need to know the API’s quirks (like wrapper keys or specific nested JSON) – the connector abstracts that away.

-   **No Data Loss or Tampering**: Always include all relevant data from the provider. Do not drop fields from the Raw response, even if they are not understood or not needed for most users; include them in Raw to be safe. Fields can be selective (focusing on commonly used fields), but Raw should be complete. Never modify values (e.g., don’t change date formats or enum codes) in the Raw output – if transformation is needed for ease of use, do it in Fields and possibly document it, but leave Raw untouched.


### Pagination & Data Retrieval

-   Most APIs paginate their responses.  **Deep connectors should handle pagination transparently**. When a user performs a read (for example, “get all contacts”), if the provider limits responses to 100 records per call, the connector must read the `ReadParams`, and look at `Since` or `NextPage` and appropiately read the correct page. Remember that `Read` is a stateless operation. The `NextPage` and `Since` parameters always come from the user, and they are responsible for storing it. Our connectors do not have any memory and that is why they do not store any parameters as such.

-   Implement the appropriate pagination method based on the provider’s API - whether they provide it in the read response body or headers as a token, an integer, a cursor, etc.
-   If the API uses page number & page size, if no `NextPage` is set, read the first page, find the nextPage in the result, and populate it in the `ReadResult` before sending it back.

-   Be mindful of  **rate limits**  and  **performance**: some providers might throttle requests. If a provider has strict rate limits, consider introducing delays or using whatever mechanisms are available to fetch data efficiently without hitting limits (e.g., using bulk endpoints if provided).

-   **No artificial limits**: Do not artificially cap the data retrieved (unless instructed by configuration). If a provider returns 1000 records and it’s within the allowed use, the connector should fetch all 1000 through pagination. If the list endpoint has a limit of maximum 10000 records, always mention it in the pull request or as a comment in the connector's read method.

-   The connector should stop when the provider indicates no more data (e.g., empty page or no next token), i.e. when `NextPage` is empty or `HasMore` is false.

-   Document any pagination specifics for the connector (for example, “API returns a  `nextPageToken`  which we use to loop until empty” or “API does not provide total count, so we fetch pages until an empty result is returned”). This helps future maintainers and users understand how data retrieval works.


### Incremental Sync (Incremental Reads)

-   Wherever supported, implement  **incremental read**  capabilities. Incremental sync means the connector can retrieve only new or updated records since the last sync, rather than pulling all data every time. This sent as `ReadParams.Since` input to the connector's read method.

-   If the provider’s API offers a  **“get changes since X”**  (e.g., a  `updated_after`  query param or a sync token), utilize it to optimize data retrieval. The connector should always accept the `ReadParams.Since` parameter and map it to the particular methodology that the provider exposes to get newer records.

-   For connectors with incremental support, ensure that the logic properly updates the checkpoint/bookmark. For example, after reading incremental data up to a certain timestamp, set the ReadResult's NextPage and Done fields correctly.

-   Indicate clearly in code which objects support incremental sync and which do not. Also note if incremental sync has limitations (e.g., “endpoint only returns changes from last 30 days” or “requires a special permission”). If incremental sync isn’t possible, comment that too.

- This only applies to read connectors.

### Testing Expectations

-   **Unit Tests**: Every deep connector implementing metadata, read, write or other functionality should include unit tests for its core logic. Use mocked API responses to simulate various scenarios. Tests are generally put under the `test` directory in the same package name the deep connector has. For example, Salesforce deep connector which has a package in `providers/salesforce` will have its tests in `tests/salesforce` and will separate the tests inside by functionality. So read tests may go inside `tests/salesforce/read` and so on. There may be a common file in `tests/salesforce` called `connector.go` which instantiates the connector for all the tests.

    -   Normal data retrieval (including multi-page responses, to test pagination).

    -   Edge cases like empty results, partial data, or API errors (e.g., 4xx/5xx responses).

    -   Writing data: test that given a sample input (possibly derived from a read Fields output), the connector produces the correct API request (mock the provider API endpoint and verify the outgoing request’s payload/headers).

    -   Metadata: test that the connector’s ListObjectMetadata returns expected fields and that the formatting (V2) is correct. If using a static schema, test that it’s loaded and returned. If using a discovery API, you might mock its response.

-   **Integration Tests**  (Manual or Automated): For deep connectors, you should run actual calls against the provider’s sandbox or test account to verify end-to-end behavior:

    -   Test a full read of each supported object, ensuring all pages of data are retrieved and that the data in Fields vs Raw is as expected.

    -   Test write operations (create/update/delete if supported) on sample data, and then read back to confirm the write succeeded.

    -   Check that incremental sync works by doing an initial read, then making a change (or waiting for new data) and doing another read with a time filter.

    -   Pay special attention to any provider-specific quirks (like rate limiting, or certain fields that cause errors). If any such issues are discovered, incorporate handling in code and document them.

-   **Repeatability**: Ensure tests can run repeatedly. Clean up any test data created (where possible) to avoid clutter. If the provider’s API has cost or rate considerations, be judicious in test frequency and possibly use recorded responses or a sandbox environment.

-   **Proxy Connectors**  should also be smoke-tested: ensure that a basic proxied request to a known endpoint returns expected data - GET returns data, invalid calls return 4XX errors, etc. Typically, this is just one or two calls since the proxy doesn’t transform data, but it verifies connectivity.

### Error handling
Always test out the provider API to investigate what the error responses from the API are. We want to standardize the error responses. If in case of an error, a provider returns a 200 response code with an error body, we should aim to convert it to a 400 error. Look at `providers/marketo/errors.go` for an implementation of this.

### Do’s and Don’ts (Summary of Best Practices)

**Do:**

-   **Follow provider conventions**: Use the provider’s object names and field names. Maintain the structure of data (nested objects remain nested in Raw, keys remain consistent).

-   **Implement all relevant interfaces**: Deep connectors should implement Read, Write (including update/delete if applicable), and ListObjectMetadata properly. Proxy connectors should implement the proxy interface and necessary auth setup, but  _not_  custom read/write logic.

-   **Handle pagination and incremental logic**  internally to abstract complexity from users. Let the deep connector “just work” by taking care of page limits, offsets, and providing incremental sync where possible.

-   **Keep Raw data intact**: Always include the raw API response data for completeness, and ensure that none of this raw information is lost or altered.

-   **Use robust error handling**: anticipate common API failure modes (network issues, rate limits, auth expiration) and handle them (retry, refresh tokens, etc.) so the connector is reliable.

-   **Write thorough tests**: especially for deep connectors, tests should cover multi-page reads, writes, and edge cases. Use both unit tests with mocks and integration tests with real API calls when possible.

-   **Document everything**: from setup steps to supported operations and any caveats. Assume the reader is a beginner Golang developer who will maintain this connector in the future – give them the context they need.

**Don’t:**

-   **Don’t flatten nested data indiscriminately**. Never flatten the data in the Raw output.


### Connector-Specific Behaviors

-   **Deep Connector Requirements**: A deep connector should implement all relevant methods defined by our connector interface contracts. This usually includes:
    -   `ListObjectMetadata`  (schema for each object type), tries to query objects in parallel if possible while keeping in mind rate limits.
    -   `Read`  (fetch records)

    -   `Write`  (create or update records)

-   **Use of Common Utilities**: Our project likely has common helper libraries in the `common` package and the `internal` packages (for OAuth flows, for HTTP request building, for schema conversions, parsing data, etc.). Use them! Don’t rewrite generic functionality. For example, use the `jquery` helper package in the `internal` package to manage json data. This keeps connectors consistent and reduces bugs.

### Guidelines

Guidelines for copilot:

-   **Follow Project Standards**: Copilot should suggest code that adheres to the naming conventions, patterns, and best practices outlined above. For example, it should use the established method and variable names (like  `ReadResult.Fields`  and  `ReadResult.Raw`  structures) rather than inventing new ones. It should prefer our logging utilities and common helper functions in suggestions. If writing a new connector file, it should place it in the correct directory and follow the general structure of existing connectors.

-   **Adhere to Do’s and Don’ts**: The AI’s suggestions should avoid forbidden patterns. It should  **not**  propose new ways to do things or hallucinate or make up anything if NOT sure. It should internalize the “don’ts” so it doesn’t introduce those errors in its reviews. Conversely, it should actively implement the “do’s” because these are known expectations in our project).

-   **Use Appropriate Tone and Format in Comments**: When Copilot generates comments or documentation, it should maintain a clear and professional tone consistent with our style. For instance, it might generate a comment like:  `// Please do a null check on this variable before using it as a pagination token.`  – which is concise and clear.  Copilot shouldn’t just output code blindly; it should only note discrepencies in the review comments. Responses should remain technical, helpful, and aligned with professional standards. If using GitHub Copilot in code review (or pull request context), it should leverage these instructions to spot deviations from guidelines. For example, it might automatically warn “This code is renaming a field from the provider. This is against our conventions." or “No test coverage for read connector was found – consider adding tests for multi-page responses.” The responses should be polite, specific, and ALWAYS refer back to the standards (as illustrated in the example comments above).