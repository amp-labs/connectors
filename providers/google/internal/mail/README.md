# Read

Objects suitable for reading from Gmail APIs:
* [Drafts](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.drafts/list) - paginated
* [Filters](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.filters/list)
* [Labels](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.labels/list)
* [Messages](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.messages/list) - paginated
* [SendAs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.sendAs/list)
* [Threads](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.threads/list) - paginated

Unusual endpoints that return empty body:
* [ForwardingAddresses](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.forwardingAddresses/list) - 204 NoContent

Endpoints that need extra setup, otherwise they return 403 Forbidden:
* [Delegates](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.delegates/list)
* [Identities](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.cse.identities/list)
* [KeyPairs](https://developers.google.com/workspace/gmail/api/reference/rest/v1/users.settings.cse.keypairs/list)
