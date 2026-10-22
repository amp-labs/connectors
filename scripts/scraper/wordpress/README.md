# WordPress Object Metadata Scraper

Scripts for collecting WordPress REST API object metadata and converting it into
the connector's standard `ListObjectMetadata` format.

The scraper discovers WordPress.com `/wp/v2/` collection endpoints, retrieves
each endpoint's JSON Schema through an HTTP `OPTIONS` request, and transforms
the results into the metadata format consumed by the connector.

## Background

WordPress.com exposes several REST API namespaces and URL follows this format:

```text
https://public-api.wordpress.com/{namespace}/{version}/{endpoint}
```

This scraper targets the `/wp/v2/` namespace, which follows the WordPress core
REST API specification.

- `v1.1` is the legacy WordPress.com API. Its object schemas are documented in
  HTML and would need to be scraped from [the documentation site](https://developer.wordpress.com/docs/api/rest-api-reference).
- `/wp/v2/` exposes machine-readable JSON Schemas through HTTP `OPTIONS`
  requests, making it suitable for automated metadata collection.

For example:

```bash
curl --location --request OPTIONS 'https://public-api.wordpress.com/wp/v2/sites/{site_id}/posts'
```

returns the JSON Schema describing the fields, types, formats, enums, and
contexts available for `post` objects.

## Workflow

```text
/wp/v2 index
    │
    ▼
Manually curated collection endpoints
    │
    ▼
1-discovery-combine
    │
    ▼
internal/files/combined-responses.json
    │
    ▼
2-metadata
    │
    ▼
providers/wordpress/internal/metadata/objectsMetadata.json.gz
```

## Prerequisites

The scripts send requests through a local proxy that forwards traffic to WordPress.com.
Assume the proxy is listening on `http://localhost:4444`
All requests in the scripts should therefore target this local address rather
than `https://public-api.wordpress.com` directly.

## Execution

### 1. Discover available endpoints

Request the `/wp/v2/` API index:

```bash
curl http://localhost:4444/wp/v2
```

The response contains the routes available for the namespace. Review the routes
and manually identify the endpoints that represent collections, such as:

- `posts`
- `pages`
- `media`
- `comments`
- `tags`
- `users`

Add the selected collection endpoints to the Go list in `internal/files/file.go`
Only collection endpoints should be included.

### 2. Collect and combine schemas

Run:

```bash
go run ./scripts/scraper/wordpress/1-discovery-combine
```

This script:

1. Sends an `OPTIONS` request to every endpoint listed in `internal/files/file.go`.
2. Stores the raw schema response for each endpoint.
3. Combines all responses into a single JSON object indexed by endpoint URL.
4. Writes the result to `internal/files/combined-responses.json`

### 3. Generate connector metadata

Run:

```bash
go run ./scripts/scraper/wordpress/2-metadata
```

This script:

1. Reads `internal/files/combined-responses.json`.
2. Transforms each WordPress JSON Schema into the connector's standard object metadata format.
3. Writes the compressed output to: `providers/wordpress/internal/metadata/objectsMetadata.json.gz`

The generated file is consumed by the connector's `ListObjectMetadata` implementation.

## Output files

| File                                                            | Produced by           | Purpose                                                |
|-----------------------------------------------------------------|-----------------------|--------------------------------------------------------|
| `internal/files/file.go`                                        | Manual                | Curated list of `/wp/v2/` collection endpoints         |
| `internal/files/combined-responses.json`                        | `1-discovery-combine` | Raw `OPTIONS` schema responses indexed by endpoint URL |
| `providers/wordpress/internal/metadata/objectsMetadata.json.gz` | `2-metadata`          | Final connector object metadata                        |

## Notes

- The `/wp/v2/` API index lists available routes but does not identify which
  routes represent collections. Endpoint selection is therefore manual.
- The generated metadata reflects the WordPress.com `/wp/v2/` implementation.
  Do not assume it is identical to the schema returned by a self-hosted
  WordPress installation.
