# SnapchatAds connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

SnapchatAds API version : v1

Some of the objects need to be read with an "organization_id" which we accept as metadata in this connector.
These are highlighted below as
- `fundingsources`
- `billingcenters`
- `transactions`
- `adaccounts`
- `members`
- `roles`

---------------------------------------------------------------------------------------------------------
| Object                                       | Resource                                      | Method |
| ---------------------------------------------| ----------------------------------------------| ------ |
| fundingsources                               | fundingsources                                | read   |
| billingcenters                               | billingcenters                                | read   |
| transactions                                 | transactions                                  | read   |
| adaccounts                                   | adaccounts                                    | read   |
| members                                      | members                                       | read   |
| roles                                        | roles                                         | read   |
| targeting/demographics/age_group             | targeting/demographics/age_group              | read   |
| targeting/demographics/gender                | targeting/demographics/gender                 | read   |
| targeting/demographics/languages             | targeting/demographics/languages              | read   |
| targeting/demographics/advanced_demographics | targeting/demographics/advanced_demographics  | read   |
| targeting/device/connection_type             | targeting/device/connection_type              | read   |
| targeting/device/os_type                     | targeting/device/os_type                      | read   | 
| targeting/device/carrier                     | targeting/device/carrier                      | read   |
| targeting/device/marketing_name              | targeting/device/marketing_name               | read   |
| targeting/geo/country                        | targeting/geo/country                         | read   |
| targeting/interests/dlxs                     | targeting/interests/dlxs                      | read   |
| targeting/interests/dlxp                     | targeting/interests/dlxp                      | read   |
| targeting/interests/nln                      | targeting/interests/nln                       | read   |
| targeting/location/categories_loi            | targeting/location/categories_loi             | read   |
---------------------------------------------------------------------------------------------------------
 
Notes:
- The organization_id is retrieved using the postAuthentication method.
- Currently, we only support objects that include the organization_id in the URL path.

