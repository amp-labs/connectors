# Jump Metadata

`schemas.json` is generated from the Jump Enterprise GraphQL schema (SDL) by
`scripts/openapi/jump/metadata`. Do not edit it by hand.

Jump does not publish an OpenAPI specification, and GraphQL introspection is disabled for API keys.
The SDL is the documented alternative: it is downloaded with the same API key and stored at
`graphql/schema.graphql` in this directory.
API reference (login required): https://my.jumpapp.com/enterprise/documentation

To refresh after Jump changes its schema, run from the repo root:

```sh
curl -H "Authorization: Bearer <token>" https://my.jumpapp.com/enterprise/graphql/schema \
  -o providers/jump/metadata/graphql/schema.graphql
go run ./scripts/openapi/jump/metadata
```

Every list query returning a `<Type>Connection` becomes an object, and the fields of its `items`
type become the object fields. Scalars map to their value types, enums become select fields with
their possible values, and nested objects and the `Json` scalar are typed as `other`.

Single-record queries (`contact`, `meeting`, ...) and aggregates (`trendingTopics`,
`trendingObjections`) are not objects.
