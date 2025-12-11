# LinkedIn connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Note: 
- In this connector, there is no GetAll endpoint; it only has search endpoints.
- This connector supports write operation for some other objects. 
- For creating the objects, it return a 201 Created HTTP status code and return shared ID in the X-Restli-Id response header instead of response body.
- For updating the objects, it return a 204 No Content HTTP status code instead of response body.

---------------------------------------------------------------------
| Object                  | Resource                | Method        |
| adTargetingFacets       | adTargetingFacets       | read          |
| dmpEngagementSourceTypes| dmpEngagementSourceTypes| read          |
| adCampaignGroups        | adCampaignGroups        | read, write   |
| adCampaigns             | adCampaigns             | read, write   |
| adAccounts              | adAccounts              | read, write   |
| adTargetTemplates       | adTargetTemplates       | write         |
| adPublisherRestrictions | adPublisherRestrictions | write         |
| inmailContents          | inMailContents          | write         |
| conversationAds         | conversationAds         | write         |
| adLiftTests             | adLiftTests             | write         |
| adExperiments           | adExperiments           | write         |
| conversions             | conversions             | write         |
| thirdPartyTrackingTags  | thirdPartyTrackingTags  | write         |
| events                  | events                  | write         |
| insightTags             | insightTags             | write         |
| adPageSets              | adPageSets              | write         |
| dmpSegments             | dmpSegments             | read, write   |
| leadForms               | leadForms               | write         |
| posts                   | posts                   | write         |
| creatives               | creatives               | write         |
| adAnalytics             | adAnalytics             | read          |
---------------------------------------------------------------------

Note: 
- In this connector, lot of endpoints having same name but differ in the body parameter which are Events, posts, creatives. Go through the Advertising & Targeting section in the link https://learn.microsoft.com/ru-ru/linkedin/marketing/overview?view=li-lms-2025-04.
- using openAPI to getting metadata for adAnalytics object and refer the link https://learn.microsoft.com/en-us/linkedin/marketing/integrations/ads-reporting/ads-reporting?view=li-lms-2025-09&tabs=http#metrics-available to know the list of fields.
- We separate the linkedIn connector into two modules:
  - Platform (regular LinkedIn consumer API)
  - Ads API
