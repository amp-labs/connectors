# LinkedIn connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Note: 
- In this connector, there is no GetAll endpoint; it only has search endpoints. Currently this connector not supported metadata and read operation because  endpoints requires query params and shared ID in the URL path. Refer the link to know what the endpoints
https://ampersand.slab.com/posts/linked-in-connector-cw5tqsrr#hklvo-deep-connector.
- This connector supports write operation for some other objects. 
- For creating the objects, it return a 201 Created HTTP status code and return shared ID in the x-restli-id response header instead of response body.
- For updating the objects, it return a 204 No Content HTTP status code instead of response body.

linkedIn API Environment: rest

LinkedIn API version: v2
---------------------------------------------------------------
| Object                  | Resource                | Method  |
| adAccounts              | adAccounts              | write   |
| adTargetTemplates       | adTargetTemplates       | write   |
| adPublisherRestrictions | adPublisherRestrictions | write   |
| inmailContents          | inMailContents          | write   |
| conversationAds         | conversationAds         | write   |
| adLiftTests             | adLiftTests             | write   |
| adExperiments           | adExperiments           | write   |
| conversions             | conversions             | write   |
| thirdPartyTrackingTags  | thirdPartyTrackingTags  | write   |
| events                  | events                  | write   |
| insightTags             | insightTags             | write   |
| adPageSets              | adPageSets              | write   |
| dmpSegments             | dmpSegments             | write   |
| leadForms               | leadForms               | write   |
| posts                   | posts                   | write   |
| creatives               | creatives               | write   |
---------------------------------------------------------------

Note: In this connector, lot of endpoints having same name but differ in the body parameter which are Events, posts, creatives. Go through the Advertising & Targeting section in the link https://learn.microsoft.com/ru-ru/linkedin/marketing/overview?view=li-lms-2025-04.