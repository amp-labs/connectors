# OpenAPI

## Obtain file

You can download the file from the JustCall API documentation.

To understand how this was discovered:
1. Navigate to [JustCall API Docs](https://developer.justcall.io/reference).
2. Open browser network tab
3. Reload the page, and locate the request that retrieves the `api-v21.json` file.

## Modifications

The original OpenAPI file has documentation bugs for several endpoints where schemas describe single item responses instead of array responses wrapped in a data envelope.

**What OpenAPI describes:**
```json
{"id": 1, "name": "John", "email": "..."}
```

**What API actually returns:**
```json
{"status": "success", "count": 1, "data": [{"id": 1, "name": "John", ...}]}
```

The following endpoints were modified to wrap the schema in the correct response structure:

1. `/v2.1/users` - Changed response schema to array wrapped in `data` key
2. `/v2.1/calls` - Changed response schema to array wrapped in `data` key
3. `/v2.1/calls_ai` - Changed response schema to array wrapped in `data` key
4. `/v2.1/meetings_ai` - Changed response schema to array wrapped in `data` key
5. `/v2.1/sales_dialer/calls` - Changed response schema to array wrapped in `data` key

This allows the api3 tool to correctly extract object schemas from the OpenAPI specification.
