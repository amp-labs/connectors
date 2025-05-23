# Installation
Mailmonkey using API key/Oauth2 with scopes/Password, etc.

# Conventions
 - Connector uses `internal/components`
 - Metadata uses V2 metadata format
 - Read supports pagination and incremental sync
 - Raw response is returned as is, no formatting done
 - Provider errors are mapped if non-standard (errors with 200 response code are converted to 4XX)
 - Unit tests cover read/write/metadata logic (placed in /tests/<provider>)

# Read
For each read object:
- [ ] Insert screenshot of metadata fields from read object installation.
Ex: ![image](https://github.com/user-attachments/assets/ebc027fb-b82b-4505-ac71-f2b61209fa4d)
- [ ] Insert screenshot of Svix destination.
- [ ] Screenshot of operations on dashboard

# Write
For each write object:
- [ ] Insert screenshot of dashboard.
- [ ] Insert screenshot of HTTP call.

# Examples
Link pull request to testdata repo which contains installation file: https://github.com/amp-labs/testdata
