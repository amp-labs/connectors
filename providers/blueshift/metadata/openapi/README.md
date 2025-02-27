# OpenAPI
The OpenAPI is available at  [Developer blueshift](https://developer.blueshift.com/openapi)

## Modification

1. campaigns
The `campaigns` object has response key `campaigns` in the OpenAPI file, but when the API is called, there is no response key. However, when the pagination is used, the response key is `results`. So the OpenAPI should be modified to replace the `campaigns` key with `results`.
 The response is as follows:
```json
{
    "total_count": 24.0,
    "total_pages": 0,
    "per_page": 500,
    "page": 0,
    "results": [
        {...}
    ]
}
```

2. segments/list

The `segments/list` object has response key `segments` in the OpenAPI file, but when the API is called, there is no response key. So the OpenAPI should be modified to remove the `segments` key.
 The response is as follows:
```json
[
    {
        "name": "All customers with email",
        "status": "active",
        "created_at": "2023-10-31T12:54:30.000Z",
        "updated_at": "2024-05-20T18:43:34.000Z",
        "approxusers": 0,
        "email_users": 0,
        "push_users": 0,
        "uuid": "2dd093e6-8c6f-4e14-8a2c-744477e1fdef",
        "approxusers_updated_at": null,
        "sms_users": 0,
        "version": 1,
        "mixin_key": null
    }
]
```

3. email_templates, external_fetches, push_templates
The `email_templates`, `external_fetches`, `push_templates` objects have a response key named `template`s` when no pagination is used. However, when the API is called with pagination, the response key changes to `results`, which is nested under the `templates` key. The OpenAPI specification should be updated to replace the templates key with results in the response.

 The response is as follows:
```json
{
    "templates": {
        "total_count": 2.0,
        "total_pages": 0,
        "per_page": 2,
        "page": 0,
        "results": [
            { ... },
        ]
    }
}
```