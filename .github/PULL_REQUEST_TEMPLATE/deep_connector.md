# Installation
Mailmonkey using API key/Oauth2 with scopes/Password, etc.

# Conventions
Read more: https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist

- [ ] Connector uses `internal/components`
- [ ] Metadata uses V2 metadata format
- [ ] Read supports pagination and incremental sync
- [ ] Raw response is returned as is for **READ** & **METADATA**, no formatting or flattening is performed.
- [ ] Write payloads should accept what `ReadResults.Fields` is returning. Any unnecessary nesting around the input is removed.
- [ ] Provider errors are mapped if non-standard (errors with 200 response code are converted to 4XX)
- [ ] Custom fields, if not human readable names, are resolved to readable names.
- [ ] Unit tests cover read/write/metadata logic (placed in /tests/<provider>)
- [ ] Appropriate object names are used. Objects need to be resources, not actions (`jobs` and not `jobs.list`).
- [ ] Modules are only being added because:
  - [ ] They share the same authentication scheme
  - [ ] The same base URL cannot be used to make a proxy call to objects in all modules
  - [ ] Different base URLs (drive.google.com vs google.com)
  - [ ] Object name collisions (same object name exists in two or more modules)

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
