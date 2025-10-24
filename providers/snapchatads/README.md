# SnapchatAds connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

SnapchatAds API version : v1

Below objects having url like v1/{organization_id}/objectName

-------------------------------------------------------------
| Object                  | Resource               | Method |
| ----------------------- | ---------------------- | ------ |
| fundingsources          | fundingsources         | read   |
| billingcenters          | billingcenters         | read   |
| transactions            | transactions           | read   |
| adaccounts              | adaccounts             | read   |
| members                 | members                | read   |
| roles                   | roles                  | read   |
-------------------------------------------------------------
 
Notes:
- The organization_id is retrieved using the postAuthentication method.
- Currently, we only support objects that include the organization_id in the URL path.
- Incremental read is support for one object - "transactions".

