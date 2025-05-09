# InstantlyAI connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

InstantlyAI API version: v2
------------------------------------------------------------------------------
| Object                     | Resource                    | Method          |
| Accounts                   | accounts                    | read and write  |
| Campaigns                  | campaigns                   | read and write  |
| Emails                     | emails                      | read and write  |
| Lead-lists                 | lead-lists                  | read and write  |
| Inbox-placement-tests      | inbox-placement-tests       | read and write  |
| Inbox-placement-analytics  | inbox-placement-analytics   | read            |
| Inbox-placement-reports    | inbox-placement-reports     | read            |
| Api-keys                   | api-keys                    | read and write  |
| Background-jobs            | background-jobs             | read            |
| Custom-tags                | custom-tags                 | read and write  |
| Block-lists-entries        | block-lists-entries         | read and write  |
| Lead-labels                | lead-labels                 | read and write  |
| Workspace-group-members    | workspace-group-members     | read and write  |
| Workspace-members          | workspace-members           | read and write  |
| Subsequences               | subsequences                | read and write  |
| Lead                       | leads/list                  | read            |
| Campaigns analytics        | campaigns/analytics         | read            |
| Campaigns analytics daily  | campaigns/analytics/daily   | read            | 
| Campaigns analytics steps  | campaigns/analytics/steps   | read            |
| Email service provider     | inbox-placement-tests/email | read            |
                              -service-provider-options  
| Email verification         | email-verification          | write           |
| Leads merge                | leads/merge                 | write           |
| Update interest status     | leads/update-interest-status| write           | 
| Leads subsequence remove   | leads/subsequence/remove    | write           |
| Leads move                 | leads/move                  | write           |
| Leads export               | leads/export                | write           |
| Leads subsequence move     | leads/subsequence/move      | write           |
| Custom-tags toggle-resource| custom-tags/toggle-resource | write           |
| Workspaces current         | workspaces/current          | write           |
------------------------------------------------------------------------------

The "Lead" object supports the POST API method instead of GET, while the Campaigns Analytics, Daily, Steps, and Email Service Provider objects return a direct response, meaning the response data is not nested within an object like "items" or "data." The remaining objects' responses are nested within an "items" object.
