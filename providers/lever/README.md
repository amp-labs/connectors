# Lever connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Lever API version : v1

Below endpoints having url path like opportunities/:opportunity/Object

| Object                  | Resource         | Method       |
| ----------------------- | ---------------- | ------------ |
| feedback                | feedback         | read         |
| files                   | files            | read         |
| interviews              | interviews       | read         | 
| notes                   | notes            | read         |
| offers                  | offers           | read         |
| panels                  | panels           | read         |
| forms                   | forms            | read         |
| referrals               | referrals        | read         |
| resumes                 | resumes          | read         |

| Object                  | Resource           | Method       |
| ----------------------- | ------------------ | ------------ |
| archive_reasons         | archive_reasons    | read         |
| audit_events            | audit_events       | read         |
| sources                 | sources            | read         |
| stages                  | stages             | read         |
| tags                    | tags               | read         |
| users                   | users              | read         |
| feedback_templates      | feedback_templates | read         |
| opportunities           | opportunities      | read         |
| postings                | postings           | read         |
| form_templates          | form_templates     | read         |
| requisitions            | requisitions       | read         |
| requisition_fields      | requisition_fields | read         |


Notes:
- Excluded the endpoints /eeo/responses/pii and /eeo/responses because they are not direct endpoints, and their responses are embedded within their respective objectName under data. Other endpoints follow a consistent structure where responses are contained under data.
- Excluded the endpoint /surveys/diversity/:posting because it includes a posting ID in the URL path, only one endpoints with posting in the connector.
