# Metadata

The static file `schemas.json` is embedded by `metadata.go` and defines list/read objects aligned with the [FastSpring API reference](https://developer.fastspring.com/reference/getting-started-with-your-api); it was **authored manually** from those public docs (paths, response keys, and field lists), not machine-generated from a checked-in OpenAPI artifact.

- **Commerce:** accounts, orders, products, subscriptions  
- **Events:** [processed](https://developer.fastspring.com/reference/list-all-processed-events) (`events-processed`, `GET /events/processed`) and [unprocessed](https://developer.fastspring.com/reference/list-all-unprocessed-events) (`events-unprocessed`, `GET /events/unprocessed`), both using response key `events` on the list payload.
