# Campaign Monitor connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Campaign Monitor API environment : v3.3
-------------------------------------------------------------------
| object          |  Resource                              | Method|
| ----------------| ---------------------------------------| ------|
| Clients         | clients.{xml|json}                     | read  |
| Admins          | admins.{xml|json}                      | read  |
|                 | transactional/smartEmail               |       | 
| Transactional   | transactional/classicEmail/groups      | read  |
|                 | transactional/messages                 |       |
--------------------------------------------------------------------

Note: 
 - Not able to check the transactional objects because it requires paid version to access on it.
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
