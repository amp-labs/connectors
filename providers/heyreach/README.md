# Heyreach connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Heyreach API environment : public
---------------------------------------------------------------------------------
| Module          | object               | Resource                     | Method|
| ----------------| ---------------------| -----------------------------| ------|
| Campaigns       | Get All              | campaign/GetAll              | read  |
|                 | AddLeadsToCampaignV2 | campaign/AddLeadsToCampaignV2| write |
|-------------------------------------------------------------------------------|
| LinkedInAccount | Get All              | li_account/GetAll            | read  |
|-------------------------------------------------------------------------------|
| List            | Get All              | list/GetAll                  | read  |
|                 | AddLeadsToListV2     | li_account/AddLeadsToListV2  | write |
|                 | CreateEmptyList      | list/CreateEmptyList         | write |
|-------------------------------------------------------------------------------|
| Inbox           | SendMessage          | inbox/SendMessage            | write |
---------------------------------------------------------------------------------

Heyreach connector offers API for:
  - PublicAuthentication
      - CheckApiKey
  - PublicCampaigns
      - GetAll
      - GetById
      - Resume
      - Pause
      - AddLeadsToCampaign
      - AddLeadsToCampaignV2
      - StopLeadInCampaign
      - GetLeadsFromCampaign
      - GetCampaignsForLead
  - PublicInbox
      - GetConversations
      - GetConversationsV2
      - GetChatroom
      - SendMessage
  - PublicLinkedInAccount
      - GetAll
      - GetById
  - PublicList
      - GetAll
      - GetById
      - GetLeadsFromList
      - DeleteLeadsFromList
      - DeleteLeadsFromListByProfileUrl
      - GetCompaniesFromList
      - AddLeadsToList
      - AddLeadsToListV2
      - GetListForLead
      - CreateEmptyList
  - PublicStats
      - Get Overall Stats
  - PublicLead
      - GetLead
  - PublicMyNetwork
      - GetMyNetworkForSender

# Getting Metadata and Read
Supported objects for metadata are PublicCampaigns, PublicLinkedInAccount, and PublicList. The remaining objects do not have a GetAll endpoint. 

Reason for unsupported object:
1. PublicInbox - This endpoint requires campaignIds and linkedInAccountIds in body.
2. PublicStats - This endpoint requires accountIds, campaignIds, startDate and endDate in body.
3. PublicMyNetwork - This endpoint requires senderID in body.

Read functionality uses Post method instead of Get method.

# Write Functions
Supported objects for write are AddLeadsToCampaignV2,AddLeadsToListV2,CreateEmptyList and SendMessage. The remaining objects do have a Post API.

There is no API for updating the objects.
