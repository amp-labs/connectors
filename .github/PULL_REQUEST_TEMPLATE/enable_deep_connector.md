---
template_name: "Enable read and write"
pr_title: "[{{ticket}}] feat({{provider}}): Enable read and write"
priority: 6
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "provider"
    prompt: "Enter provider name"
---
# Conventions
Read more: [Deep Connector Guide](https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist)

## Summary
Enable read and write for **{{provider}}** provider.
Reference ticket [{{ticket}}](https://linear.app/ampersand/issue/{{ticket}}).


# Validation
Read more: https://ampersand.slab.com/posts/deep-connectors-guide-6x4fhxne#ht0ds-reviewer-checklist

## Installation
Mailmonkey using API key/Oauth2 with scopes/Password, etc.

## Read
For each read object:
- [ ] Insert screenshot of metadata fields from read object installation.
  Ex: ![image](https://github.com/user-attachments/assets/ebc027fb-b82b-4505-ac71-f2b61209fa4d)
- [ ] Insert screenshot of Svix destination.
- [ ] Screenshot of operations on dashboard

## Write
For each write object:
- [ ] Insert screenshot of dashboard.
- [ ] Insert screenshot of HTTP call.

## Examples
Link pull request to testdata repo which contains installation file: https://github.com/amp-labs/testdata
