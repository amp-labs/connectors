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
| Inbox-placement-analytics | inbox-placement-analytics  | read   |
| Inbox-placement-reports   | inbox-placement-reports    | read   |
| Api-keys                  | api-keys                   | read   |
| Background-jobs           | background-jobs            | read   |
| Custom-tags               | custom-tags                | read   |
| Block-lists-entries       | block-lists-entries        | read   |
| Lead-labels               | lead-labels                | read   |
| Workspace-group-members   | workspace-group-members    | read   |
| Workspace-members         | workspace-members          | read   |
| Subsequences              | subsequences               | read   |
| Lead                      | leads/list                 | read   |
| Campaigns analytics       | campaigns/analytics        | read   |
| Campaigns analytics daily | campaigns/analytics/daily  | read   |
| Campaigns analytics steps | campaigns/analytics/steps  | read   |
| Email service provider    | inbox-placement-tests/email| read   |
                              -service-provider-options  
------------------------------------------------------------------

The "Lead" object supports the POST API method instead of GET, while the Campaigns Analytics, Daily, Steps, and Email Service Provider objects return a direct response, meaning the response data is not nested within an object like "items" or "data." The remaining objects' responses are nested within an "items" object.
