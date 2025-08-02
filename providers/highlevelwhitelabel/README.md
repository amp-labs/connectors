# HighLevelWhilteLabel connector


## Supported Objects 
Below is an exhaustive list of objects & methods supported on the objects

| object                       | Resource                     | Method|
| -----------------------------| -----------------------------| ------|
| businesses                   | businesses                   | read  |
| calendars                    | calendars                    | read  |
| calendars/groups             | calendars/groups             | read  |
| campaigns                    | campaigns                    | read  |
| conversations/search         | conversations/search         | read  |
| emails/schedule              | emails/schedule              | read  |
| forms/submissions            | forms/submissions            | read  |
| forms                        | forms                        | read  |
| invoices                     | invoices                     | read  |
| invoices/template            | invoices/template            | read  |
| invoices/schedule            | invoices/schedule            | read  |
| invoices/estimate/list       | invoices/estimate/list       | read  |
| invoices/estimate/template   | invoices/estimate/template   | read  |
| links                        | links                        | read  |
| blogs/authors                | blogs/authors                | read  |
| blogs/categories             | blogs/categories             | read  |
| funnels/lookup/redirect/list | funnels/lookup/redirect/list | read  |
| funnels/funnel/list          | funnels/funnel/list          | read  |
| opportunities/pipelines      | opportunities/pipelines      | read  |
| payment/orders               | payment/orders               | read  |
| payments/transactions        | payments/transactions        | read  |
| payments/subscriptions       | payments/subscriptions       | read  |
| payments/coupon/list         | payments/coupon/list         | read  |
| products                     | products                     | read  |
| products/inventory           | products/inventory           | read  |
| products/collections         | products/collections         | read  |
| products/reviews             | products/reviews             | read  |
| proposals/document           | proposals/document           | read  |
| proposals/templates          | proposals/templates          | read  |
| store/shipping-zone          | store/shipping-zone          | read  |
| store/shipping-carrier       | store/shipping-carrier       | read  |
| store/store-setting          | store/store-setting          | read  |
| surveys                      | surveys                      | read  |
| users                        | users                        | read  |
| workflows                    | workflows                    | read  |
| locations/search             | locations/search             | read  |
| custom-menus                 | custom-menus                 | read  |

Note:
 - For single-segment paths (e.g., "businesses"), the URL must have a trailing slash at the end.
   Example: https://highlevel.stoplight.io/docs/integrations/a8db8afcbe0a3-get-businesses-by-location
 - For multi-segment paths (e.g., "calendars/groups"), the URL does not require a trailing slash.
   Example: https://highlevel.stoplight.io/docs/integrations/89e47b6c05e67-get-groups
 -From the above endpoints, some require locationId, altId, altType, limit, and offset as query parameters.
 - Some endpoints use different pagination parameters, such as offset and skip, but share the same limit parameter.
 For more Ref: https://ampersand.slab.com/posts/highlevel-connector-0naldomk#hefaa-supported-metadata-objects.