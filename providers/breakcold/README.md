# Breakcold connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

| Object          | Resource       | Method        |
| ----------------| ---------------| --------------|
| status          | status         | read          |
| workspaces      | workspaces     | read          |
| members         | members        | read          |
| leads           | leads          | read          |
| tags            | tags           | read          |
| lists           | lists          | read          |
| notes           | notes          | read          |
| reminders       | reminders      | read          |

- The endpoints below use the POST method instead of the GET method and use appropriate object name.

| Original objectname | Changed objectname |
| --------------------| -------------------|
| leads/list          | leads              |          
| notes/list          | notes              |
| reminders/list      | reminders          |

- 