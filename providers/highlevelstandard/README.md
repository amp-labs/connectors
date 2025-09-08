# HighLevelStandard connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

| object                          | Resource                        | Method       |
| --------------------------------| --------------------------------| -------------|
| businesses                      | businesses                      | read, write  |
| calendars                       | calendars                       | read, write  |
| calendars/groups                | calendars/groups                | read, write  |
| campaigns                       | campaigns                       | read         |
| conversations/search            | conversations/search            | read         |
| emails/schedule                 | emails/schedule                 | read         |
| forms/submissions               | forms/submissions               | read         |
| forms                           | forms                           | read         |
| invoices                        | invoices                        | read, write  |
| invoices/template               | invoices/template               | read, write  |
| invoices/schedule               | invoices/schedule               | read, write  |
| invoices/estimate/list          | invoices/estimate/list          | read         |
| invoices/estimate/template      | invoices/estimate/template      | read, write  |
| links                           | links                           | read, write  |
| blogs/authors                   | blogs/authors                   | read         |
| blogs/categories                | blogs/categories                | read         |
| funnels/lookup/redirect/list    | funnels/lookup/redirect/list    | read         |
| funnels/funnel/list             | funnels/funnel/list             | read         |
| opportunities/pipelines         | opportunities/pipelines         | read         |
| payment/orders                  | payment/orders                  | read         |
| payments/transactions           | payments/transactions           | read         |
| payments/subscriptions          | payments/subscriptions          | read         |
| payments/coupon/list            | payments/coupon/list            | read         |
| products                        | products                        | read, write  |
| products/inventory              | products/inventory              | read         |
| products/collections            | products/collections            | read, write  |
| products/reviews                | products/reviews                | read         |
| proposals/document              | proposals/document              | read         |
| proposals/templates             | proposals/templates             | read         |
| store/shipping-zone             | store/shipping-zone             | read, write  |
| store/shipping-carrier          | store/shipping-carrier          | read         |
| store/store-setting             | store/store-setting             | read         |
| surveys                         | surveys                         | read         |
| users                           | users                           | read, write  |
| workflows                       | workflows                       | read         |
| locations/search                | locations/search                | read         |
| custom-menus                    | custom-menus                    | read, write  |
| calendars/events/appointments   | calendars/events/appointments   | write        |
| calendars/events/block-slots    | calendars/events/block-slots    | write        |
| contacts                        | contacts                        | write        |
| objects                         | objects                         | write        |
| associations                    | associations                    | write        |
| associations/relations          | associations/relations          | write        |
| custom-fields                   | custom-fields                   | write        |
| custom-fields/folder            | custom-fields/folder            | write        |
| conversations                   | conversations                   | write        |
| conversations/messages          | conversations/messages          | write        |
| conversations/messages/inbound  | conversations/messages/inbound  | write        |
| conversations/messages/outbound | conversations/messages/outbound | write        |
| conversations/messages/upload   | conversations/messages/upload   | write        |
| emails/builder                  | emails/builder                  | write        |
| invoices/text2pay               | invoices/text2pay               | write        |
| invoices/estimate               | invoices/estimate               | write        |
| locations                       | locations                       | write        |
| blogs/posts                     | blogs/posts                     | write        |
| funnels/lookup/redirect         | funnels/lookup/redirect         | write        |
| opportunities                   | opportunities                   | write        |
| payments/coupon                 | payments/coupon                 | write        |

Note:
 - For single-segment paths (e.g., "businesses"), the URL must have a trailing slash at the end.
   Example: https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location
 - For multi-segment paths (e.g., "calendars/groups"), the URL does not require a trailing slash.
   Example: https://highlevel.stoplight.io/docs/integrations/89e47b6c05e67-get-groups
 -From the above endpoints, some require locationId, altId, altType, limit, and offset as query parameters.
 - Some endpoints use different pagination parameters, such as offset and skip, but share the same limit parameter.
 For more Ref: https://ampersand.slab.com/posts/highlevel-connector-0naldomk#hefaa-supported-metadata-objects.