# Metadata

The static file `schemas.json` was created by inspecting the [SendGrid OpenAPI specs](https://github.com/twilio/sendgrid-oai) and [SendGrid API reference](https://www.twilio.com/docs/sendgrid/api-reference), then translating the supported objects into our static schema format. Paths in `schemas.json` are relative to `/v3` (prepended at read time against `https://api.sendgrid.com`).

## Supported objects

| Object | Path | Response key |
| --- | --- | --- |
| `contacts` | `/marketing/contacts` | `result` |
| `lists` | `/marketing/lists` | `result` |
| `segments` | `/marketing/segments` | `results` |
| `singlesends` | `/marketing/singlesends` | `result` |
| `templates` | `/templates` | `result` |
| `field_definitions` | `/marketing/field_definitions` | `custom_fields` |
| `verified_senders` | `/verified_senders` | `results` |
| `senders` | `/marketing/senders` | `results` |
| `bounces` | `/suppression/bounces` | root array |
| `blocks` | `/suppression/blocks` | root array |
| `spam_reports` | `/suppression/spam_reports` | root array |
| `unsubscribes` | `/suppression/unsubscribes` | root array |
| `invalid_emails` | `/suppression/invalid_emails` | root array |
| `asm_groups` | `/asm/groups` | root array |
| `categories` | `/categories` | root array |
| `subusers` | `/subusers` | root array |
| `event_webhook_settings` | `/user/webhooks/event/settings/all` | `webhooks` |
| `parse_webhook_settings` | `/user/webhooks/parse/settings` | `result` |

Mail Send (`POST /v3/mail/send`) is write-only and is not modeled as a listable object.

## Connector-level metadata

None required. The API host is fixed (`api.sendgrid.com`) and auth is an API key. Optional subuser impersonation (`on-behalf-of`) is request-scoped and is not connector metadata.
