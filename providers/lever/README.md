# Lever connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Lever API version : v1

| Object                  | Resource           | Method       |
| ----------------------- | ------------------ | ------------ |
| archive_reasons         | archive_reasons    | read         |
| audit_events            | audit_events       | read         |
| sources                 | sources            | read         |
| stages                  | stages             | read         |
| tags                    | tags               | read         |
| users                   | users              | read, write  |
| feedback_templates      | feedback_templates | read, write  |
| opportunities           | opportunities      | read, write  |
| postings                | postings           | read, write  |
| form_templates          | form_templates     | read, write  |
| requisitions            | requisitions       | read, write  |
| requisition_fields      | requisition_fields | read, write  |
| uploads                 | uploads            | write        |
| contacts                | contacts           | write        |


Notes:
- Excluded the endpoints /eeo/responses/pii and /eeo/responses because they are not direct endpoints, and their responses are embedded within their respective objectName under data. Other endpoints follow a consistent structure where responses are contained under data.
- Excluded the endpoint /surveys/diversity/:posting because it includes a posting ID in the URL path, only one endpoints with posting in the connector.
- Currently we do not support below endpoints because they requires an shared ID in the URL path.
  - opportunity ID
    - feedback
    - files
    - interviews
    - notes
    - offers
    - panels
    - forms
    - referrals
    - resumes
  - posting ID
    - apply
  - user ID
    - deactivate
    - reactivate 
- The delete endpoints below cannot be removed if they were created within the Lever application. Only endpoints created through the API are eligible for deletion via API.
    - feedback_templates
	- form_templates
