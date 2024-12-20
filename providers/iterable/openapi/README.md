
# OpenAPI

## Obtain file

You can download the file from https://api.iterable.com/api-docs.

To understand how this was discovered:
1. Navigate to [Iterable API Docs](https://api.iterable.com/api/docs).
2. Open browser network tab
3. Reload the page, and locate the request that retrieves the `api-docs.json` file.

## Modify the File
A known bug in the converter requires a manual adjustment to the `api-docs.json` file.

Instructions:
1. Open the api-docs.json file.
2. Find all occurrences of:
    ```json
    "$ref": "#/definitions/JsValue"
    ```
3. Replace them with
    ```json
    "$ref": "#/components/schemas/JsValue"
    ```
