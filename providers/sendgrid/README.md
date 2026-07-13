# SendGrid Connector

Deep connector for [Twilio SendGrid](https://www.twilio.com/docs/sendgrid). Proxy support already exists via `providers/sendGrid.go`.

## Auth

API key attached as `Authorization: Bearer <api_key>`. See [API getting started](https://www.twilio.com/docs/sendgrid/for-developers/sending-email/api-getting-started).

Base URL: `https://api.sendgrid.com` (paths in schemas are relative to `/v3`).

## Connector-level metadata

**None required.**

Per [How to determine what connector metadata needs to be collected](https://ampersand.slab.com/posts/mu7822bu):

| Candidate | Used across endpoints? | Source | Decision |
| --- | --- | --- | --- |
| Workspace / subdomain | No — host is fixed (`api.sendgrid.com`) | N/A | Not metadata |
| Account / user id | Not required on request paths | Token / profile APIs | Not metadata |
| Subuser (`on-behalf-of`) | Optional header for a subset of calls | Client-supplied | Keep at request level (Proxy), not connector metadata |
| Object ids (list id, template id, …) | Narrow / per-call | Request params | Not metadata |

## Object metadata (`ListObjectMetadata`)

Static V2 schemas from [SendGrid OpenAPI](https://github.com/twilio/sendgrid-oai), focused on marketing + suppressions + webhooks:

| Object | Path (under `/v3`) | Response key |
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
