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
| age_group               | age_group              | read   |
| gender                  | gender                 | read   |
| languages               | languages              | read   |
| advanced_demographics   | advanced_demographics  | read   |
| connection_type         | connection_type        | read   |
| os_type                 | os_type                | read   | 
| carrier                 | carrier                | read   |
| marketing_name          | marketing_name         | read   |
| country                 | country                | read   |
| dlxs                    | dlxs                   | read   |
| dlxp                    | dlxp                   | read   |
| nln                     | nln                    | read   |
| categories_loi          | categories_loi         | read   |
-------------------------------------------------------------
 
Notes:
- The organization_id is retrieved using the postAuthentication method.
- Currently, we only support objects that include the organization_id in the URL path.

