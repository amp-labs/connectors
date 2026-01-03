---
template_name: "New Connector in catalog"
pr_title: "[{{ticket}}] feat({{provider}}): Add Provider to catalog"
priority: 1
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "provider"
    prompt: "Enter provider name"
---
# Configuration
<Any special connector notes>

# Conventions
- [ ] Provider name is camelcase (`goTo` and not `goto`)
- [ ] Should cover all modules within the connector (ex, `goTo` has modules `webinar` and `meeting` or `google` has modules `drive` and `calendar`)
- [ ] Base URLs do NOT have version information
- [ ] DocsURLs actually link to user-friendly documentation (do not link to very technical documentation)
- [ ] All required metadata variables are templated (`{{.var}}`) and defined in `ProviderInfo.Metadata`
- [ ] If OAuth2 connector, if `workspace` is required, `Oauth2Opts.ExplicitWorkspaceRequired` is ALSO set to true
- [ ] Basic smoke tests added (valid request succeeds, invalid request fails)
- [ ] Docs and logos attached or linked
- [ ] Modules are only being added because:
  - [ ]  They share the same authentication scheme
  - [ ] The same base URL cannot be used to make a proxy call to objects in all modules
  - [ ] Different base URLs (drive.google.com vs google.com)
  - [ ] Object name collisions (same object name exists in two or more modules)

## Testing
### GET
URL: <http://localhost:4444/v2/some-api-call>
Postman screenshot (must show the request URL, the response status code & body clearly)

### POST
URL: <http://localhost:4444/v2/some-api-call>
Postman screenshot (must show the request URL, the response status code & body clearly)

<Please add the same information for all other methods available - PUT, PATCH, DELETE, etc>

## Pagination
Please add screenshots that show successful pagination using the connector.

## Raw token response
<In case of Oauth2 auth connector paste token response>

# Logo & Icons
<Add screenshot of what icons, logos where selected>
