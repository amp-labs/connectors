---
template_name: "Impl Read Connector"
pr_title: "[{{ticket}}] feat({{provider}}): ReadConnector"
priority: 4
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "provider"
    prompt: "Enter provider name"
---
# Conventions
Read more: [Deep Connector Guide](https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist)

## Summary
Implementation of a `connectors.ReadConnector` for **{{provider}}** provider.
Reference ticket [{{ticket}}](https://linear.app/ampersand/issue/{{ticket}}).

## Checklist
- [ ] Connector uses `internal/components`
- [ ] Read supports pagination and incremental sync
- [ ] Raw response is returned as is for **READ**, no formatting or flattening is performed.
- [ ] Provider errors are mapped if non-standard (errors with 200 response code are converted to 4XX)
- [ ] Custom fields, if not human-readable names, are resolved to readable names. 
- [ ] Unit tests document expected connector behavior using sample provider responses (connectors/provider/<provider>)
- [ ] Live tests cover read logic (placed in connectors/tests/<provider>)
- [ ] Appropriate object names are used. Objects need to be resources, not actions (`jobs` and not `jobs.list`).
- [ ] Connector metadata required for initialization is listed in provider catalog inside `ProviderMetadata` (connectors/providers/<provider>.go)
