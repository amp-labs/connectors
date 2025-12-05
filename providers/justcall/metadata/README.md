# Metadata

The static file `schemas.json` was manually created based on the JustCall API v2.1 documentation and live API responses, as JustCall does not provide a publicly accessible OpenAPI spec.

## Supported Objects

The schema includes definitions for the following core objects:

* **Users** (`/v2.1/users`) - Team members and agents
* **Calls** (`/v2.1/calls`) - Call logs and recordings
* **Contacts** (`/v2.1/contacts`) - Contact directory
* **Texts** (`/v2.1/texts`) - SMS and MMS messages
* **Phone Numbers** (`/v2.1/phone-numbers`) - JustCall phone numbers
* **Webhooks** (`/v2.1/webhooks`) - Webhook subscriptions

## API Notes

* **Authentication**: API Key + Secret in `Authorization` header (format: `api_key:api_secret`)
* **Base URL**: `https://api.justcall.io`
* **API Version**: v2.1 (current), v2.0 (deprecated)
* **Rate Limits**: Varies by plan (Team: 1800/hr, Pro: 3600/hr, Business: 5400/hr)
* **Pagination**: Uses `page` and `per_page` query parameters
* **Date Format**: `yyyy-mm-dd hh:mm:ss` or `yyyy-mm-dd`

## References

* [JustCall API Reference](https://developer.justcall.io/reference)
* [JustCall Authentication](https://developer.justcall.io/reference/authentication)
