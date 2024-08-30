# Description

This folder contains **scripts that run against provider APIs**.

These files provide examples of how to instantiate and use a connector.
Here, you can find scripts for `Read`, `Write`, `ListObjectMetadata`, `Delete`, and other unique operations used by connectors.

Every operation that exists on a connector has a corresponding script to support that functionality.

# Structure

Scripts will be located under `test/<PROVIDER_NAME>/<OPERATION>/main.go`.

They import a test connector that is instantiated under `test/<PROVIDER_NAME>/connector.go`.

# Running instructions

You must provide the credentials and configurations of a provider by creating a JSON file (see below).
Then, execute the following command in the project root directory:

```
go run ./test/salesforce/read-write/
```
This will work assuming you have `salesforce-creds.json` under `connectors` root.

## File location

By default, JSON file is expected to be at the root of the project `connectors`.
The file should be named `"<PROVIDER_NAME | to_kebab_case>" + "-creds.json"`.
Alternatively, you can specify a different file path and name the file as you wish.
To apply this, set the environment variable `"<PROVIDER_NAME | to_upper_snake>" + "_CRED_FILE"` to the desired file location.

**Examples:**

| Provider       | File Name                  | Env holding file path     |
|----------------|----------------------------|---------------------------|
| dynamicsCRM    | dynamics-crm-creds.json    | DYNAMICS_CRM_CRED_FILE    |
| zendeskSupport | zendesk-support-creds.json | ZENDESK_SUPPORT_CRED_FILE |
| anthropic      | anthropic-creds.json       | ANTHROPIC_CRED_FILE       |


## File Content
The file must have a `provider` field. Other required fields can be inferred by checking the constructor in `test/<PROVIDER_NAME>/connector.go`.
If any required fields are missing, an error message will indicate the missing fields.

File formats can be categorized into common authentication types. Below, you will find examples for each authentication category.

**API Key**
```json
{
  "provider": "anthropic",
  "apiKey": "..."
}
```

**Basic auth with Password**
```json
{
  "provider": "braintree",
  "username": "...",
  "password": "..."
}
```

**Oauth2**
```json
{
  "provider": "dynamicsCRM",
  "clientId": "...",
  "clientSecret": "...",
  "accessToken": "...",
  "refreshToken": "...",
  "substitutions": {
    "workspace": "..."
  },
  "expiry": "2024-03-26T16:22:32.768450621+02:00",
  "expiryFormat": "RFC3339Nano",
  "state": "vvqv3VAe6lLW3nYHdcOQNA",
  "scopes": "..."
}
```
