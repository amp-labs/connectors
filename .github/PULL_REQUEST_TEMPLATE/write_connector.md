---
template_name: "Impl Write Connector"
pr_title: "[{{ticket}}] feat({{provider}}): WriteConnector"
priority: 5
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "provider"
    prompt: "Enter provider name"
---
# Conventions
Read more: [Deep Connector Guide](https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist)

## Summary
Implementation of a `connectors.WriteConnector` for **{{provider}}** provider.
Reference ticket [{{ticket}}](https://linear.app/ampersand/issue/{{ticket}}).

## Checklist
- [ ] Connector uses `internal/components`
- [ ] Write payloads should accept what `ReadResults.Fields` is returning. Any unnecessary nesting around the input is removed.
- [ ] Provider errors are mapped if non-standard (errors with 200 response code are converted to 4XX) 
- [ ] Unit tests document expected connector behavior using sample provider responses (connectors/provider/<provider>)
- [ ] Live tests cover write/delete logic (placed in connectors/tests/<provider>)
- [ ] Appropriate object names are used. Objects need to be resources, not actions (`jobs` and not `jobs.list`).
- [ ] Connector metadata required for initialization is listed in provider catalog inside `ProviderMetadata` (connectors/providers/<provider>.go)
