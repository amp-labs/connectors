# Metadata

The static file `schemas.json` was created manually from the [Slack API documentation](https://docs.slack.dev/reference/methods), since Slack's OpenAPI spec is archived and out of date.

It only covers write-only objects: objects that can be created or updated but have no list call, so there is no record to sample. Each one lists the arguments of its create method. Every other object's metadata is sampled from the API.
