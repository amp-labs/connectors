# Heyreach connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Heyreach API environment : public

| Object | Resource | Method
| :-------- | :------- | 
| campaign | read |
| li_account | read |
| list | read |

Heyreach connector offers API are:
  - PublicAuthentication
  - PublicCampaigns
  - PublicInbox
  - PublicLinkedInAccount
  - PublicList
  - PublicStats
  - PublicLead
  - PublicMyNetwork

# Getting Metadata and Read
Supported objects for metadata are PublicCampaigns, PublicLinkedAccount, and PublicList. The remaining objects do not have a GetAll endpoint. 

Reason for unsupported object:
1. PublicInbox - This endpoint requires campaignIds and linkedInAccountIds in body.
2. PublicStats - This endpoint requires accountIds, campaignIds, startDate and endDate in body.
3. PublicMyNetwork - This endpoint requires senderID in body.