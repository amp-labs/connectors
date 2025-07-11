# Facebook connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Facebook API version : v19.0

Below object having the urlpath format: act_{ad_account_id}/objectName
------------------------------------------------------------------------------
| Object                       | Resource                    | Method        |
| ---------------------------- | ----------------------------| ------------- |
|  Users                       | users                       | read          |
|  Ad_place_page_sets          | ad_place_page_sets          | read, write   |
|  Adrules_library             | adrules_library             | read, write   |
|  Adplayables                 | adplayables                 | read, write   |
|  Adlabels                    | adlabels                    | read, write   |
|  Adimages                    | adimages                    | read          |
|  Account_controls            | account_controls            | read, write   |
|  Ads                         | ads                         | read, write   |
|  Adsets                      | adsets                      | read, write   |
|  Advertisable_applications   | advertisable_applications   | read, write   |
|  Advideos                    | advideos                    | read          |  
|  Applications                | applications                | read          |
|  Broadtargetingcategories    | broadtargetingcategories    | read          |
|  Customaudiencestos          | customaudiencestos          | read, write   |
|  Customconversions           | customconversions           | read, write   |
|  Deprecatedtargetingadsets   | deprecatedtargetingadsets   | read          |
|  Dsa_recommendations         | dsa_recommendations         | read          |
|  Impacting_ad_studies        | impacting_ad_studies        | read          |
|  Minimum_budgets             | minimum_budgets             | read          |
|  Promote_pages               | promote_pages               | read          |
|  Publisher_block_lists       | publisher_block_lists       | read, write   |
|  Reachfrequencypredictions   | reachfrequencypredictions   | read, write   |
|  Saved_audiences             | saved_audiences             | read          |
|  Subscribed_apps             | subscribed_apps             | read, write   |
|  Targetingbrowse             | targetingbrowse             | read          |
|  Tracking                    | tracking                    | read, write   |
|  Adcreatives                 | adcreatives                 | read, write   |
|  Campaigns                   | campaigns                   | read, write   |
|  Customaudiences             | customaudiences             | read, write   |
|  Assigned_users              | assigned_users              | write         |
------------------------------------------------------------------------------


Below object having the urlpath format: business_id/objectName
----------------------------------------------------------------------------------------------------------
| Object                                     | Resource                                  | Method        |
| -------------------------------------------| ------------------------------------------| ------------- |
| Ad_studies                                 | ad_studies                                | read, write   |
| Adnetworkanalytics_results                 | adnetworkanalytics_results                | read          |
| Adspixels                                  | adspixels                                 | read, write   |
| Business_invoices                          | business_invoices                         | read          |
| Business_users                             | business_users                            | read, write   |
| Client_pages                               | client_pages                              | read          |
| Client_pixels                              | client_pixels                             | read          |
| Client_product_catalogs                    | client_product_catalogs                   | read          |
| Client_whatsapp_business_accounts          | client_whatsapp_business_accounts         | read          |
| Clients                                    | clients                                   | read          |
| Collaborative_ads_collaboration_requests   | collaborative_ads_collaboration_requests  | read          |
| Collaborative_ads_suggested_partners       | collaborative_ads_suggested_partners      | read          |
| Event_source_groups                        | event_source_groups                       | read, write   |
| Extendedcredits                            | extendedcredits                           | read          |
| Initiated_audience_sharing_requests        | initiated_audience_sharing_requests       | read          |
| Managed_partner_ads_funding_source_details | managed_partner_ads_funding_source_details| read          |
| Owned_apps                                 | owned_apps                                | read, write   |
| Owned_businesses                           | owned_businesses                          | read, write   |
| Owned_pages                                | owned_pages                               | read, write   |
| Owned_pixels                               | owned_pixels                              | read          |
| Owned_product_catalogs                     | owned_product_catalogs                    | read, write   |
| Owned_whatsapp_business_accounts           | owned_whatsapp_business_accounts          | read          |
| Pending_client_ad_accounts                 | pending_client_ad_accounts                | read          |
| Pending_client_apps                        | pending_client_apps                       | read          |
| Pending_client_pages                       | pending_client_pages                      | read          |
| Pending_owned_ad_accounts                  | pending_owned_ad_accounts                 | read          |
| Pending_owned_pages                        | pending_owned_pages                       | read          |
| Pending_users                              | pending_users                             | read          |
| Received_audience_sharing_requests         | received_audience_sharing_requests        | read          |
| System_users                               | system_users                              | read, write   |
| Adaccount                                  | adaccount                                 | write         |
| China_business_onboarding_attributions     | china_business_onboarding_attributions    | write         |
| Claim_custom_conversions                   | claim_custom_conversions                  | write         |
| Client_apps                                | client_apps                               | write         |
| Adnetworkanalytics                         | adnetworkanalytics                        | write         |
| Owned_ad_accounts                          | owned_ad_accounts                         | write         |
----------------------------------------------------------------------------------------------------------

Note: 
- Checked only few more endpoints due to unsupported scopes and upgraded plan to access them.