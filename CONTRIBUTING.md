# Contributing to Connectors

## Development Environment Setup

```
make install/dev
```

## Prerequisites

1. To test connectors, you'll need your own developer account and instance with the provider you're testing (like
   Salesforce or Hubspot). Follow the provider guide on how to do
   this [here.](https://docs.withampersand.com/provider-guides/overview)

2. You'll need to obtain the credentials (using oAuth2.0 or apiKey) to link the connectors to the providers.

3. Add a `creds.json` at the root of the connectors repo. Since each provider has different credential parameters and
   everyone has different environments, we've made a `CredentialRegistry` and `Reader` interface to make credential
   loading consistent. Please use these. We currently support `JSONReader`, `EnvVarReader`, and `ValueReader`.

| Credential Type                    | Description                         | Example `creds.json`                                                                                                                                                                                           |
|------------------------------------|-------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| OAuth 2.0                          | Authentication using OAuth 2.0 flow | ```{ "id": "someID" "projectId": "someProjectId",     "providerApp": {        "provider": "salesforce",        "clientId": "YOUR_CLIENT_ID",        "clientSecret": "YOUR_CLIENT_SECRET",  "scopes": null}}``` |
| API key                            | Api key based authentication        | ```{ "provider": "salesforce", "apiKey": "<api-key-here>", "metadata": { "workspace": "test-instance"  } }```                                                                                                  |
| Basic Authentication               | Basic authentication                | ```{"provider":"salesforce","username": "<username-here>", "password": "<password-here>","metadata": { "workspace": "test-instance",}}```                                                                      | 
| OAuth 2.0 Client Credentials Grant | Authentication using OAuth 2.0 flow | ```{ "provider": "salesforce", "clientId": "<salesforce-provider-clientId-here>","clientSecret": "<salesforce-provider-clientSecret-here>","scopes": "","metadata": {"workspace": "test-instance",}}```        |

## Adding a Proxy Connector

1. Add a new file `connectors/providers/<PROVIDER>.go` for eg: `connectors/providers/linkedIn.go` (camelcase always!)
2. Update the file depending on the connector's auth type.

### Provider Authentication Types

Here's a reference table for implementing different authentication types in your provider.go file:

| Auth Type                            | Description                             | Key Configuration Fields                                                                | Example Provider                                  |
|--------------------------------------|-----------------------------------------|-----------------------------------------------------------------------------------------|---------------------------------------------------|
| OAuth2 Authorization Code (3-legged) | OAuth flow requiring user authorization | ```AuthType: OAuth2, OauthOpts: { GrantType: AuthorizationCode, AuthURL, TokenURL }```  | See [LinkedIn Provider](/providers/linkedIn.go)   |
| OAuth2 Client Credentials (2-legged) | OAuth flow using client credentials     | ```AuthType: OAuth2, OauthOpts: { GrantType: ClientCredentials, TokenURL }```           | See [Marketo Provider](/providers/marketo.go)     |
| API Key                              | Authentication using an API key         | ```AuthType: ApiKey, ApiKeyOpts: { Type: InHeader/InQuery, HeaderName, ValuePrefix }``` | See example in [Monday](/providers/monday.go)     |
| Basic Auth                           | Username/password authentication        | ```AuthType: Basic, BaseURL```                                                          | See [Insightly Provider](/providers/insightly.go) |

Additional configuration notes:

- For workspace-specific providers, use `{{.workspace}}` in BaseURL
- Set `ExplicitScopesRequired: true` if the provider requires explicit scope definition
- Use `PostAuthInfoNeeded: true` if additional information is needed after authentication
- Configure `Support` struct to specify provider capabilities (BulkWrite, Proxy, Read, Subscribe, Write)

### Testing your proxy connector

Proxy server script

`go run scripts/proxy/proxy.go`

This will run the proxy server at port `:4444`

You can now make requests against `localhost:4444` and they will be routed to the provider you're working on.

> Note: For OAuth connectors, you'd need to run `go run scripts/oauth/token.go` to go through the OAuth flow first and
> get the credentials to save them into `creds.json`.

Once the proxy endpoint is tested with different kind of API calls, <b>you can make a PR. </b>

## Adding a Deep Connector

Ensure you have made a proxy connector PR first and is merged and only then <b>work on the following changes in
order</b>.



> Refer the `smartleadV2` connector as a reference. It might be the case that you may have to add other files or
> functions for a specific connector.


> Please make separate PRs for each of the steps below.

### 1. Add list metadata functionality:

Files to add:

1. For functionality add, `connectors/<PROVIDER>/metadata.go`.
2. Implement the `SchemaProvider` interface in `connectors/<PROVIDER>/connector.go`
3. For testing add, `connectors/test/<PROVIDER>/metadata/metadata.go`

Run your tests:

`go run ./test/<PROVIDER>/metadata`

### 2. Add read action support for the objects agreed upon.

Files to add or update:

1. For functionality add, `connectors/<PROVIDER>/handlers.go`. Here we need to implement the interface
   `buildReadRequest` and `parseReadResponse`.
2. Implement the `Reader` interface in `connectors/<PROVIDER>/connector.go`
3. For testing add, `connectors/<PROVIDER>/read.go`

Run your tests:

`go run ./test/<PROVIDER>/read`

### 3. Add write action support.

Files to update:

1. For functionality add, `connectors/<PROVIDER>/handlers.go`. Here we need to implement the interface
   `buildWriteRequest` and `parseWriteResponse`.
2. Implement the `Writer` interface in `connectors/<PROVIDER>/connector.go`
3. For testing add, `connectors/<PROVIDER>/write.go`

Run your tests:

`go run ./test/<PROVIDER>/write`

### 4. Add delete support

Files to update:

1. For functionality add, `connectors/<PROVIDER>/handlers.go`. Here we need to implement the interface
   `buildDeleteRequest` and `parseDeleteResponse`.
2. Implement the `Deleter` interface in `connectors/<PROVIDER>/connector.go`
3. For testing add, `connectors/<PROVIDER>/delete.go`

Run your tests:

`go run ./test/<PROVIDER>/delete`

<Note>
Refer the `smartleadV2` connector as a reference. 
</Note>
