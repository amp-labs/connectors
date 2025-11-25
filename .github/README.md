# Creating Pull Request Templates

This guide explains how to create new PR templates for the project.
These templates are used by the Git pre-push hook to prompt contributors for information and automatically generate PR titles and bodies.

## Template Location

All templates must be placed in:`.github/PULL_REQUEST_TEMPLATE/`.

## YAML Header

Each template starts with a YAML header, enclosed between --- markers at the top of the file. The header defines:

* `template_name` – the human-readable name displayed when selecting a template.
* `pr_title` – the PR title pattern. You can include dynamic fields using `{{field_name}}`.
* `priority` – controls the display order of templates in the selection menu. Lower numbers are displayed first.
* `fields` – a list of prompts for the contributor to fill out. Each field has:
  * `name` – the identifier used in `{{}}` placeholders.
  * `prompt` – the message shown to the contributor when asking for input.

Example Template
```markdown
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
## Summary
Implementation of a `connectors.ObjectMetadataConnector` for **{{provider}}** provider.
Reference ticket [{{ticket}}](https://linear.app/ampersand/issue/{{ticket}}).
```

## How It Works

When a contributor pushes a new branch without a remote, the hook will prompt to select a template.
For the selected template, the script asks for all fields defined in the fields section.
The `pr_title` and **PR body** are automatically populated:
* Any `{{field_name}}` in `pr_title` or the **Markdown body** is replaced with the contributor's input.

The PR body content starts after the YAML header (---), so any Markdown you add below the header will appear in the GitHub PR body.

## Priority and Ordering

Templates are displayed in order based on the priority value.
0 is the highest priority (displayed first).
If two templates have the same priority, they are displayed in alphabetical order by filename.

## Notes

* Keep field names short and descriptive. They are used in the PR title and body.
* All fields in fields are mandatory. Contributors will be prompted for them every time the template is selected.
* Dynamic placeholders in `pr_title` or **body** must match exactly the name in the field definition (`{{name}}`).
