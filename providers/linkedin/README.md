# LinkedIn connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Note: 
- In this connector, there is no GetAll endpoint; it only has search endpoints.
- For creating the objects, it return a 201 Created HTTP status code and return shared ID in the Location response header instead of response body.
- For updating the objects, it return a 204 No Content HTTP status code instead of response body.

linkedIn API Environment: rest

LinkedIn API version: v2
---------------------------------------------------------------------
| Object                  | Resource                | Method        |
| adTargetingFacets       | adTargetingFacets       | read          |
| dmpEngagementSourceTypes| dmpEngagementSourceTypes| read          |
| adCampaignGroups        | adCampaignGroups        | read          |
| adCampaigns             | adCampaigns             | read          |
| AdAccounts              | adAccounts              | read, write   |
| AdTargetTemplates       | adTargetTemplates       | write         |
| AdPublisherRestrictions | adPublisherRestrictions | write         |
| InmailContents          | inMailContents          | write         |
| ConversationAds         | conversationAds         | write         |
| AdLiftTests             | adLiftTests             | write         |
| AdExperiments           | adExperiments           | write         |
| Conversions             | conversions             | write         |
| ThirdPartyTrackingTags  | thirdPartyTrackingTags  | write         |
| Events                  | events                  | write         |
| InsightTags             | insightTags             | write         |
| ConversionEvents        | conversionEvents        | write         |
| AdPageSets              | adPageSets              | write         |
| DmpSegments             | dmpSegments             | write         |
| LeadForms               | leadForms               | write         |
| UgcPosts                | ugcPosts                | write         | 
| Posts                   | posts                   | write         |
| Creatives               | creatives               | write         |
---------------------------------------------------------------------

Note: In this connector, lot of endpoints having same name but differ in the body parameter which are Events, posts, creatives. Go through the Advertising & Targeting section in the link https://learn.microsoft.com/ru-ru/linkedin/marketing/overview?view=li-lms-2025-04.