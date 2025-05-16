# InstantlyAI connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

InstantlyAI API version: v2
------------------------------------------------------------------
| Object                    | Resource                   | Method |
| Accounts                  | accounts                   | read   |
| Campaigns                 | campaigns                  | read   |
| Emails                    | emails                     | read   |
| Lead-lists                | lead-lists                 | read   |
| Inbox-placement-tests     | inbox-placement-tests      | read   |
| Background-jobs           | background-jobs            | read   |
| Custom-tags               | custom-tags                | read   |
| Block-lists-entries       | block-lists-entries        | read   |
| Lead-labels               | lead-labels                | read   |
| Workspace-group-members   | workspace-group-members    | read   |
| Workspace-members         | workspace-members          | read   |
| Lead                      | leads/list                 | read   |
| Campaigns analytics       | campaigns/analytics        | read   |
| Campaigns analytics daily | campaigns/analytics/daily  | read   |
| Campaigns analytics steps | campaigns/analytics/steps  | read   |
| Email service provider    | inbox-placement-tests/email| read   |
                              -service-provider-options  
| Audit logs                | audit-logs                 | read   |                              
------------------------------------------------------------------

Note:
 - The "Lead" object uses the POST API method to retrieve metadata, unlike the GET method typically used for such read actions.
 - Among the supported objects, Campaigns Analytics, Daily, Steps, and Email Service Provider return their response data directly at the root level of the JSON, without nesting.
 - In contrast, the other objects return their response data nested within an "items" field.