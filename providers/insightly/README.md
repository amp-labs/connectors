# Objects API Coverage

The Professional plan does not provide access to all API endpoints available in Enterprise Insightly plan.

Below is a list of objects that are either:

* ❌ 404 Not Found: The object is not available in the Professional plan.
* 🔒 403 Forbidden. Unknown reason, could be due to the plan but why the message differs.

Although these objects are supported in the connector implementation, their functionality cannot be validated at this
time due to access restrictions.

| Object                   | Status | Message                  |
|--------------------------|--------|--------------------------|
| CommunityComments        | 🔒403  | user doesn't have access |
| CommunityForums          | 🔒403  | user doesn't have access |
| CommunityPosts           | 🔒403  | user doesn't have access |
| ForumCategories          | 🔒403  | user doesn't have access |
| KnowledgeArticle         | 🔒403  | user doesn't have access |
| KnowledgeArticleCategory | 🔒403  | user doesn't have access |
| KnowledgeArticleFolder   | 🔒403  | user doesn't have access |
| OpportunityLineItem      | ❌404   | N/A                      |
| Pricebook                | ❌404   | N/A                      |
| PricebookEntry           | ❌404   | N/A                      |
| Product                  | ❌404   | N/A                      |
| Prospect                 | 🔒403  | user doesn't have access |
| Quotation                | ❌404   | N/A                      |
| QuotationLineItem        | ❌404   | N/A                      |
| Ticket                   | 🔒403  | user doesn't have access |
