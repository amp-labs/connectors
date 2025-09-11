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

Objects that support creation.
* [Draft](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/create)
* [Label](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.labels/create)
* [Filters](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.filters/create)
* [SendAs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.sendAs/create)

Sending messages:
https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/send

## Messages

The `raw` property must contain the entire email message, encoded using Base64url.

The message itself must be formatted according to
[RFC 2822](https://www.rfc-editor.org/rfc/rfc2822.html) (for basic headers and body).
For attachments and rich content, it should follow 
[MIME](https://www.rfc-editor.org/rfc/rfc2045) conventions, which extend RFC 2822.
