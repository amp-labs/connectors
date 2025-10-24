# SnapchatAds connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

SnapchatAds API version : v1

Below objects having url like v1/{organization_id}/objectName

-------------------------------------------------------------------
| Object                  | Resource               | Method       |
| ----------------------- | ---------------------- | -------------|
| fundingsources          | fundingsources         | read         |
| billingcenters          | billingcenters         | read, write  |
| transactions            | transactions           | read         |
| adaccounts              | adaccounts             | read, write  |
| members                 | members                | read, write  |
| roles                   | roles                  | read, write  |
-------------------------------------------------------------------
 
Notes:
- The organization_id is retrieved using the postAuthentication method.
- Currently, we only support objects that include the organization_id in the URL path.
- Incremental read is support for one object - "transactions".
- The API documentation mentions that we must use the POST method to update billingcenters, but it turns out that the PUT method works too.
- For the roles object, there's no mention of updating organization-based member roles in the documentation, yet the update operation still goes through.
- This connector doesn't add the record ID to the URL path. Instead, it's expected to be passed in the request body. So, when params.RecordId is set in the WriteParams struct, the system treats the request as an update.
