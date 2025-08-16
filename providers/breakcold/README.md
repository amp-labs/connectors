# Breakcold connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

| Object          | Resource       | Method        |
| ----------------| ---------------| --------------|
| status          | status         | read, write   |
| workspaces      | workspaces     | read          |
| members         | members        | read          |
| leads           | leads          | read          |
| tags            | tags           | read, write   |
| lists           | lists          | read, write   |
| notes           | notes          | read, write   |
| reminders       | reminders      | read, write   |
| attribute       | attribute      | write         |
| lead            | lead           | write         |
| leads/all-list  | leads/all-list | write         |

- The endpoints below use the POST method instead of the GET method and use appropriate object name.

| Original objectname | Changed objectname |
| --------------------| -------------------|
| leads/list          | leads              |          
| notes/list          | notes              |
| reminders/list      | reminders          |

- POST objects (attribute, lead) use singular names, while PATCH and DELETE use plural (attributes, leads).
- Users can provide singular object names (attribute, lead) for POST, PATCH, and DELETE operations. The code internally converts them to plural (attributes, leads) for PATCH and DELETE requests.

- 