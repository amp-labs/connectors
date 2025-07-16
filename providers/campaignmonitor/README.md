# Campaign Monitor connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Campaign Monitor API environment : v3.3
--------------------------------------------------------------------------
| object          |  Resource                              | Method      |
| ----------------| ---------------------------------------| ------------|
| Clients         | clients.{xml|json}                     | read,write  |
| Admins          | admins.{xml|json}                      | read,write  |
--------------------------------------------------------------------------

Note: 
 - Currently we do not support below endpoints because they requires an shared ID in the URL path.
   - clientid
      - lists       
      - segments
      - suppressionlist
      - templates    
      - people  
      - tags    
      - campaigns
      - scheduled
      - drafts  
      - sendingdomains
      - journeys
    - campaignid
      - emailclientusage
      - recipients
      - bounces
      - opens
      - clicks
      - unsubscribes
      - spam
    - emailId
       - recipients
       - opens.
       - clicks
       - bounces
       - unsubscribes
    - listId
       - customfields
       - segments
       - active
       - unconfirmed
       - unsubscribed
       - bounced
    - segmentID
       - active
 - The following endpoints require clientId as a query parameter. Although the documentation marks it as optional, it’s only optional when the user is using a specific client API key — but since we use an OAuth token, it is required. Refer to: https://www.campaignmonitor.com/api/v3-3/transactional/#smart-email-listing.
    - transactional/smartEmail
    - transactional/classicEmail/groups
    - transactional/messages