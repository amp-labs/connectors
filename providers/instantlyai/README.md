# InstantlyAI connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

InstantlyAI API version: v2
------------------------------------------------------------------------------
| Object                     | Resource                    | Method          |
| Accounts                   | accounts                    | read and write  |
| Campaigns                  | campaigns                   | read and write  |
| Emails                     | emails                      | read            |
| Lead-lists                 | lead-lists                  | read and write  |
| Inbox-placement-tests      | inbox-placement-tests       | read and write  |
| Background-jobs            | background-jobs             | read            |
| Custom-tags                | custom-tags                 | read and write  |
| Block-lists-entries        | block-lists-entries         | read and write  |
| Lead-labels                | lead-labels                 | read and write  |
| Workspace-group-members    | workspace-group-members     | read and write  |
| Workspace-members          | workspace-members           | read and write  |
| Lead                       | leads/list                  | read            |
| Campaigns analytics        | campaigns/analytics         | read            |
| Campaigns analytics daily  | campaigns/analytics/daily   | read            | 
| Campaigns analytics steps  | campaigns/analytics/steps   | read            |
| Email service provider     | inbox-placement-tests/email | read            |
                              -service-provider-options  
| Audit logs                 | audit-logs                  | read            |                   
| Email verification         | email-verification          | write           |
| Email reply                | emails/reply                | write           |
| Leads merge                | leads/merge                 | write           |
| Update interest status     | leads/update-interest-status| write           | 
| Leads subsequence remove   | leads/subsequence/remove    | write           |
| Leads move                 | leads/move                  | write           |
| Leads export               | leads/export                | write           |
| Leads subsequence move     | leads/subsequence/move      | write           |
| Custom-tags toggle-resource| custom-tags/toggle-resource | write           |
| Workspaces current         | workspaces/current          | write           |
------------------------------------------------------------------------------

Note:
 - The "Lead" object uses the POST API method to retrieve metadata, unlike the GET method typically used for such read actions.
 - Among the supported objects, Campaigns Analytics, Daily, Steps, and Email Service Provider return their response data directly at the root level of the JSON, without nesting.
 - In contrast, the other objects return their response data nested within an "items" field.