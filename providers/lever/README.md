# Lever connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Lever API version : v1

Below endpoints having url path like opportunities/:opportunity/Object

| Object                  | Resource         | Method       |
| ----------------------- | ---------------- | ------------ |
| feedback                | feedback         | read, write  |
| files                   | files            | read, write  |
| interviews              | interviews       | read, write  | 
| notes                   | notes            | read, write  |
| offers                  | offers           | read         |
| panels                  | panels           | read, write  |
| forms                   | forms            | read, write  |
| referrals               | referrals        | read         |
| resumes                 | resumes          | read         |
| addlinks                | addLinks         | write        |
| removeLinks             | removeLinks      | write        |
| addTags                 | addTags          | write        |
| removeTags              | removeTags       | write        |
| addSources              | addSources       | write        |
| removeSources           | removeSources    | write        |
| stage                   | stage            | write        |
| archived                | archived         | write        |


| Object                  | Resource           | Method       |
| ----------------------- | ------------------ | ------------ |
| archive_reasons         | archive_reasons    | read         |
| audit_events            | audit_events       | read         |
| sources                 | sources            | read         |
| stages                  | stages             | read         |
| tags                    | tags               | read, write  |
| users                   | users              | read. write  |
| feedback_templates      | feedback_templates | read, write  |
| opportunities           | opportunities      | read, write  |
| postings                | postings           | read, write  |
| form_templates          | form_templates     | read, write  |
| requisitions            | requisitions       | read, write  |
| requisition_fields      | requisition_fields | read, write  |
| uploads                 | uploads            | write        |
| contacts                | contacts           | write        |

Below endpoints having url path like users/:userId/Object

| Object                  | Resource         | Method       |
| ----------------------- | ---------------- | ------------ |
| deactivate              | deactivate       | write        |
| reactivate              | reactivate       | write        |

Below endpoints having url path like postings/:postingId/Object

| Object                  | Resource         | Method       |
| ----------------------- | ---------------- | ------------ |
| apply                   | apply            | write        |


Notes:
- Excluded the endpoints /eeo/responses/pii and /eeo/responses because they are not direct endpoints, and their responses are embedded within their respective objectName under data. Other endpoints follow a consistent structure where responses are contained under data.
- Excluded the endpoint /surveys/diversity/:posting because it includes a posting ID in the URL path, only one endpoints with posting in the connector.
- Below delete endpoints cannot be delete that were created within the Lever application. Only endpoints that were created via API can be deleted via API.
    - feedback_templates
	- notes
	- form_templates
    - interviews
