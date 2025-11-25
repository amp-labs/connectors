---
template_name: "️🗃️ Default PR"
pr_title: "[{{ticket}}] {{intent}}({{provider}}): {{message}}"
priority: 0
fields:
  - name: "ticket"
    prompt: "Enter Linear ticket number"
  - name: "intent"
    prompt: "Enter one from (feat/tidy/chore/test)"
  - name: "provider"
    prompt: "Enter provider name"
  - name: "message"
    prompt: "Message telling what changed (Ex: Add search API)"
---
Enter a description for your PR here.
