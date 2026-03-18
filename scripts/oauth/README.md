# Description

The `token.go` script acquires access and refresh tokens using OAuth flows:

- Authorization Code
- Authorization Code with PKCE

# Running instructions

You must provide the credentials and configurations for a provider by creating a `creds.json` file in the root of the
project (see below).

Then, execute the following command in this directory:

```bash
go run token.go
```

### CLI Flags

- `-port`: Port to listen on (default: `8080`)
- `-sslcert`: Path to SSL certificate (default: `.ssl/server.crt`)
- `-sslkey`: Path to SSL key (default: `.ssl/server.key`)
- `-proto`: Protocol to use, `http` or `https` (default: `http`)
- `-callback`: The full OAuth callback path (default: `/callbacks/v1/oauth`)
- `-writeCreds`: If set to `true`, the script updates `creds.json` and the corresponding provider-specific file with the
  new token information.

# Credential Updates

When the `-writeCreds` flag is set to `true`, the script updates:

1. `creds.json` at the project root.
2. The corresponding provider-specific file (e.g., `google-creds.json`).

If a `module` name is specified in the `creds.json` file, the script will update the module-specific file (e.g.,
`google-calendar-creds.json` otherwise will fall back to `google-creds.json`).

# File Content

The `creds.json` file at the root should follow this format for OAuth:

```json
{
  "provider": "google",
  "module": "gmail",
  "clientId": "...",
  "clientSecret": "...",
  "scopes": "...",
  "metadata": {
    "workspace": "..."
  }
}
```

### Examples of provider-specific file

| Provider       | Module   | File Name                  |
|----------------|----------|----------------------------|
| dynamicsCRM    |          | dynamics-crm-creds.json    |
| zendeskSupport |          | zendesk-support-creds.json |
| google         | calendar | google-calendar-creds.json |

If `-writeCreds` is enabled, these files will be updated with `accessToken`, `refreshToken`, and `expiry` information
once the OAuth flow completes.
