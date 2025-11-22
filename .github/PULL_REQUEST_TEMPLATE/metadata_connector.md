---
template_name: "Impl Metadata Connector"
pr_title: "[{{ticket}}] feat({{provider}}): ObjectMetadataConnector"
priority: 3
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "provider"
    prompt: "Enter provider name"
---
# Conventions
Read more: [Deep Connector Guide](https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist)

## Summary
Implementation of a `connectors.ObjectMetadataConnector` for **{{provider}}** provider.
Reference ticket [{{ticket}}](https://linear.app/ampersand/issue/{{ticket}}).

## Checklist
- [ ] Connector uses `internal/components`
- [ ] Metadata uses V2 metadata format
- [ ] Raw response is returned as is for **METADATA**, no formatting or flattening is performed.
- [ ] Provider errors are mapped if non-standard (errors with 200 response code are converted to 4XX)
- [ ] Custom fields, if not human-readable names, are resolved to readable names. 
- [ ] Unit tests document expected connector behavior using sample provider responses (connectors/provider/<provider>)
- [ ] Live tests cover metadata logic (placed in connectors/tests/<provider>/<module-if-any>/metadata)
- [ ] Appropriate object names are used. Objects need to be resources, not actions (`jobs` and not `jobs.list`).
- [ ] Modules are only being added because:
  - [ ] They share the same authentication scheme
  - [ ] The same base URL cannot be used to make a proxy call to objects in all modules
  - [ ] Different base URLs (drive.google.com vs google.com)
  - [ ] Object name collisions (same object name exists in two or more modules)
- [ ] Connector metadata required for initialization is listed in provider catalog inside `ProviderMetadata` (connectors/providers/<provider>.go)
