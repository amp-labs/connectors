# Campaign Monitor connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

Campaign Monitor API environment : v3.3
---------------------------------------------------------------------------
| object          |  Resource                              | Method       |
| ----------------| ---------------------------------------| -------------|
| Clients         | clients.{xml|json}                     | read, write  |
| Admins          | admins.{xml|json}                      | read, write  |
|                 | transactional/smartEmail               |              | 
| Transactional   | transactional/classicEmail/groups      | read         |
|                 | transactional/messages                 |              |
| lists           | clients/{clientid}/lists.json          | read         |
|                 | lists/{clientid}.json                  | write        |
| Segments        | clients/{clientid}/segments.json       | read         |
| Suppressionlist | clients/{clientid}/suppressionlist.json| read         |
| Templates       | clients/{clientid}/templates.json      | read         |
|                 | templates/{clientid}.json              | write        |
| People          | clients/{clientid}/people.json         | read         |
| Tags            | clients/{clientid}/tags.json           | read         |
| Campaigns       | clients/{clientid}/campaigns.json      | read         |
|                 | campaigns/{clientid}.json              | write        |
| Scheduled       | clients/{clientid}/scheduled.json      | read         |
| Drafts          | clients/{clientid}/drafts.json         | read         |
| Sendigndomains  | clients/{clientid}/sendingdomains.json | read         |
| Journeys        | clients/{clientid}/journeys.json       | read         |
| Suppress        | clients/{clientid}/suppress.json       | write        |
| Credits         | clients/{clientid}/credits.json        | write        |
| People          | clients/{clientid}/people.json         | write        |
| Sendingdomains  | clients/{clientid}/sendingdomains.json | write        |
---------------------------------------------------------------------------

Note: 
 - The connector now supports objects with a client ID in the URL, which is passed in the JSON and retrieved in the code.
 - Ignored the remaining objects with shared IDs like listId, campaignId, and segmentId, as they can be obtained after using the object with clientId.
 - Not able to check the transactional objects because it requires paid version to access on it.