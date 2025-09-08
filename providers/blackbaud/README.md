# Blackbaud connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

| Object                                 | Resource                                      | Method        |
| ---------------------------------------| --------------------------------------------- |---------------|
| crm-adnmg/batches                      | crm-adnmg/batches                             | write         |
| crm-adnmg/batches/revenue              | crm-adnmg/batches/revenue                     | write         |
| crm-adnmg/batchtemplates               | crm-adnmg/batchtemplates/list                 | read          |
| crm-adnmg/businessprocess/launch       | crm-adnmg/businessprocess/launch              | write         |
| crm-adnmg/businessprocessinstances     | crm-adnmg/businessprocessinstances/list       | read          |
| crm-adnmg/businessprocessparameterset  | crm-adnmg/businessprocessparameterset/search  | read          |
| crm-adnmg/businessprocessstatus        | crm-adnmg/businessprocessstatus/list          | read          |
| crm-adnmg/currencies                   | crm-adnmg/currencies/list                     | read          |
| crm-adnmg/notifications                | crm-adnmg/notifications                       | write         |
| crm-adnmg/sites                        | crm-adnmg/sites/search                        | read          |
| crm-conmg/addresses                    | crm-conmg/addresses                           | write         |
| crm-conmg/alternatelookupids           | crm-conmg/alternatelookupids                  | write         |
| crm-conmg/constituentappeals           | crm-conmg/constituentappeals                  | write         |
| crm-conmg/constituentappealresponses   | crm-conmg/constituentappealresponses          | write         |
| crm-conmg/constituentattributes        | crm-conmg/constituentattributes               | write         |
| crm-conmg/constituentcorrespondencecode| crm-conmg/constituentcorrespondencecode       | write         |
| crm-conmg/constituentnotes             | crm-conmg/constituentnotes                    | write         |
| crm-conmg/constituents                 | crm-conmg/constituents                        | write         |
| crm-conmg/educationalhistories         | crm-conmg/educationalhistories                | write         |
| crm-conmg/emailaddresses               | crm-conmg/emailaddresses                      | write         |
| crm-conmg/fundraisers                  | crm-conmg/fundraisers                         | write         |
| crm-conmg/individuals                  | crm-conmg/individuals                         | write         |
| crm-conmg/interaction                  | crm-conmg/interaction                         | write         |
| crm-conmg/mergetwoconstituents         | crm-conmg/mergetwoconstituents                | write         |
| crm-conmg/organizations                | crm-conmg/organizations                       | write         |
| crm-conmg/phones                       | crm-conmg/phones                              | write         |
| crm-conmg/relationshipjobsinfo         | crm-conmg/relationshipjobsinfo                | write         |
| crm-conmg/solicitcodes                 | crm-conmg/solicitcodes                        | write         |
| crm-conmg/tribute                      | crm-conmg/tribute                             | write         |
| crm-evtmg/events                       | crm-evtmg/events/search                       | read,write    |
| crm-evtmg/locations                    | crm-evtmg/locations/list                      | read,write    |
| crm-evtmg/registrants                  | crm-evtmg/registrants/search                  | read          |
| crm-evtmg/registrants                  | crm-evtmg/registrants                         | write         |
| crm-evtmg/registrationoptions          | crm-evtmg/registrationoptions                 | write         |
| crm-evtmg/registrationtypes            | crm-evtmg/registrationtypes/search            | read          |
| crm-evtmg/registrationtypes            | crm-evtmg/registrationtypes                   | write         |
| crm-fndmg/designations/hierarchies     | crm-fndmg/designations/hierarchies/list       | read          |
| crm-fndmg/educationalhistory           | crm-fndmg/educationalhistory/search           | read          |
| crm-fndmg/fundraisingpurposes          | crm-fndmg/fundraisingpurposes/search          | read,write    |
| crm-fndmg/fundraisingpurposerecipients | crm-fndmg/fundraisingpurposerecipients/search | read,write    |
| crm-fndmg/fundraisingpurposetypes      | crm-fndmg/fundraisingpurposetypes/search      | read          |
| crm-mktmg/appeals                      | crm-mktmg/appeals/list                        | read,write    |
| crm-mktmg/correspondencecodes          | crm-mktmg/correspondencecodes/list            | read,write    |
| crm-mktmg/responsecategories           | crm-mktmg/responsecategories                  | write         |
| crm-mktmg/segments                     | crm-mktmg/segments                            | write         |
| crm-mktmg/segments/recordsources       | crm-mktmg/segments/recordsources              | read          |
| crm-mktmg/solicitcodes                 | crm-mktmg/solicitcodes/list                   | read          |
| crm-prsmg/prospectcontactreports       | crm-prsmg/prospectcontactreports              | write         |
| crm-prsmg/prospectmanagers             | crm-prsmg/prospectmanagers/search             | read          |
| crm-prsmg/prospectopportunities        | crm-prsmg/prospectopportunities/search        | read,write    |
| crm-prsmg/prospectplans                | crm-prsmg/prospectplans                       | write         |
| crm-prsmg/prospects                    | crm-prsmg/prospects/search                    | read,write    |
| crm-prsmg/prospectsconstituency        | crm-prsmg/prospectsconstituency               | write         |
| crm-prsmg/prospectsegmentations        | crm-prsmg/prospectsegmentations               | write         |
| crm-prsmg/prospectsteps                | crm-prsmg/prospectsteps                       | write         |
| crm-prsmg/stewardshipplans             | crm-prsmg/stewardshipplans                    | write         |
| crm-prsmg/stewardshipplansteps         | crm-prsmg/stewardshipplansteps/search         | read          |
| crm-prsmg/stewardshipplansteps         | crm-prsmg/stewardshipplansteps                | write         |
| crm-revmg/payments                     | crm-revmg/payments/search                     | read,write    |
| crm-revmg/recurringgifts               | crm-revmg/recurringgifts                      | write         |
| crm-revmg/revenuenotes                 | crm-revmg/revenuenotes                        | write         |
| crm-revmg/revenuetransactions          | crm-revmg/revenuetransactions/search          | read          |
| crm-volmg/jobs                         | crm-volmg/jobs/search                         | read,write    |
| crm-volmg/occurrences                  | crm-volmg/occurrences/search                  | read,write    |
| crm-volmg/volunteerassignments         | crm-volmg/volunteerassignments/search         | read,write    |
| crm-volmg/volunteers                   | crm-volmg/volunteers/search                   | read,write    |
| crm-volmg/volunteerschedules           | crm-volmg/volunteerschedules                  | write         |

In this connector, there are nearly eight solutions. As of now, we concentrate on CRM solutions. Within the CRM solution, there are approximately nine modules, each with a different segment in the base URL. The following are listed below.

| Module                 |                Link                     |
| -----------------------|-----------------------------------------|
| CRM Administration     | https://api.sky.blackbaud.com/crm-adnmg |
| CRM Analysis           | https://api.sky.blackbaud.com/crm-anamg |
| CRM Constituent        | https://api.sky.blackbaud.com/crm-conmg |
| CRM Event              | https://api.sky.blackbaud.com/crm-evtmg |
| CRM Fundraising        | https://api.sky.blackbaud.com/crm-fndmg |
| CRM Marketing          | https://api.sky.blackbaud.com/crm-mktmg |
| CRM Prospect           | https://api.sky.blackbaud.com/crm-prsmg |
| CRM Revenue            | https://api.sky.blackbaud.com/crm-revmg |
| CRM Volunteer          | https://api.sky.blackbaud.com/crm-volmg |

Note: 
- We need to provide endpoints along with their module segments, because all the modules share the same base URL and only differ in the module segment. Therefore, I decided to combine the module segment with the object name.  For example, use the object name as crm-adnmg/batchtemplates, crm-conmg/addresses.
- Most metadata endpoints end with list or search, and we should ignore these suffixes, using the appropriate object name instead. For instance, use crm-adnmg/batchtemplates rather than crm-adnmg/batchtemplates/list.
