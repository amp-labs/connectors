# Metadata

The static file `schemas.json` is embedded by `metadata.go` and defines metadata objects for the Livestorm deep connector. This schema was authored manually from the public Livestorm API documentation:

- [List events](https://developers.livestorm.co/reference/get_events)
- [List people](https://developers.livestorm.co/reference/get_people)
- [List people attributes](https://developers.livestorm.co/reference/get_people-attributes)
- [List chat messages from a session](https://developers.livestorm.co/reference/get_sessions-id-chat-messages)

Bulk session registrants (`POST …/sessions/{id}/people/bulk`) are not exposed on Write; a future Bulk-style integration would live outside this deep connector’s write path (see Salesforce bulk-write in-repo).
