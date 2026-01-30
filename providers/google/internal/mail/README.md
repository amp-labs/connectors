# Read

This section lists Gmail API endpoints currently supported for reading.  
Endpoints are categorized by pagination, unusual responses, and extra setup requirements.

---

## Paginated Endpoints
These endpoints return multiple items and support pagination via `pageToken` and `maxResults`:

* [Drafts](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list)
* [Messages](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list)
* [Threads](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list)

---

## Non-Paginated Endpoints
These endpoints return a single set of items:

* [Filters](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.filters/list)
* [Labels](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.labels/list)
* [SendAs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.sendAs/list)


## Unusual Endpoints
These endpoints may return an empty body (**204 No Content**):
* [ForwardingAddresses](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.forwardingAddresses/list) - 204 NoContent

---

## Endpoints Requiring Extra Setup
These endpoints require special configuration, otherwise they return and error (**403 Forbidden**):
* [Delegates](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.delegates/list)
  > This method is only available to service account clients that have been delegated domain-wide authority.
* [Identities](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.cse.identities/list)
  > For users managing their own identities and keypairs, requests require hardware key encryption turned on and configured.
* [KeyPairs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.cse.keypairs/list)
  > For users managing their own identities and keypairs, requests require hardware key encryption turned on and configured.

# Write

This section lists Gmail API endpoints currently supported for creating objects or sending data.

## Creatable Objects
The following endpoints allow creating resources:

* [Drafts](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/create)
* [Labels](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.labels/create)
* [Filters](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.filters/create)
* [SendAs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.sendAs/create)

---

## Sending Messages
The [Messages.send](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/send) endpoint is used to send email messages.

### Requirements

* The `raw` property must contain the entire email message, **Base64url-encoded**.
* The message must follow **RFC 2822** formatting for headers and body:
  * [RFC 2822](https://www.rfc-editor.org/rfc/rfc2822.html) — basic headers and message body
* For attachments or rich content, messages should follow **MIME conventions**:
  * [RFC 2045](https://www.rfc-editor.org/rfc/rfc2045) — MIME format extensions

---

## Notes

* Only the endpoints listed above are currently supported for write operations.
* Sending messages requires proper encoding and header formatting to comply with Gmail API expectations.
* Other write-related endpoints (e.g., modifying labels) may be added as needed.
