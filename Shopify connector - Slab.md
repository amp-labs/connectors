**Shopify** is a leading **e-commerce platform** that allows individuals and businesses (merchants) to create, manage, and grow online stores. It provides the full infrastructure to sell products online, in person, and across multiple channels, without requiring merchants to manage their own servers or complex backend systems.

# Provider Terminologies 

| **Term** | **Definition** |
| --- | --- |
| **Shopify / Store** | A merchant’s account/store on Shopify where products are sold. |
| **Merchant** | The owner of a Shopify store. |
| **Admin API** | API for managing store data like products, orders, customers, and inventory. |
| **Dev Dashboard** | Shopify’s interface for developers to manage apps, stores, and API credentials. |
| **Custom App** | A Shopify app created directly by a merchant in their store (not in the public app store). |
| **Public App** | An app listed in Shopify App Store and installable by any merchant. |
| **Embedded App** | An app that runs inside the Shopify admin UI (Shopify Admin embedded iframe). |
| **Customer** | A person who buys or interacts with a merchant’s store. |
| **Order** | A transaction created when a customer purchases products from a store. |
| **Product** | An item available for sale in a Shopify store. |

# API Usage

Shopify offers a suite of APIs that allow developers to extend the platform’s built-in features. These APIs allow app developers to read and write app user data, interoperate with other systems and platforms, and add new functionality to Shopify.

Requirements for using shopify APIs:

- All APIs are subject to the [Shopify API License and Terms of Use](https://www.shopify.com/legal/api-terms).
- All APIs are subject to rate limits but limits are different for each type of the APIs.
- All APIs require developers to authenticate.
- Some APIs are versioned.

# API Pagination

Shopify's GraphQL API uses cursor-based pagination that allows to fetch max `250 records` per request.

Pagination ref link: [https://shopify.dev/docs/api/usage/pagination-graphql](https://shopify.dev/docs/api/usage/pagination-graphql)

Shopify allows up to `250 records` per request, but 100 is chosen as a balanced default to avoid rate limiting while maintaining reasonable performance. But it allows to override the limit from the pageSize param.

## How It Works?

1. **Request Variables**
    - **first:** Number of records to fetch (default: 100, max: 250)
    - **after:** Cursor pointing to the last item of the previous page (optional for first request)
    - **query:** Optional filter string for incremental sync
1. **Response Structure**

```json
 {
   "data": {
     "products": {
       "nodes": [...],
       "pageInfo": {
         "hasNextPage": true,
         "endCursor": "eyJsYXN0X2lkIjoxMjM0NTY3ODl9"
       }
     }
   }
 }
```

1. **Pagination Flow**

```
 Page 1: first=100, after=null     → returns endCursor="abc123"
 Page 2: first=100, after="abc123" → returns endCursor="def456"
 Page 3: first=100, after="def456" → returns hasNextPage=false (done)
```

# API Rate Limit

Shopify's GraphQL APIs use a calculated query cost per request method to rate limit the APIs.

Rate limit ref link: [https://shopify.dev/docs/api/usage/limits](https://shopify.dev/docs/api/usage/limits)

All Shopify APIs use a [leaky bucket algorithm](https://en.wikipedia.org/wiki/Leaky_bucket) to manage requests. Rate limit is defined based on the query cost and points/seconds.

**_Each page of 100 records ≈ 102 points_**

| API | [Rate-limiting method](https://shopify.dev/docs/api/usage/limits#rate-limiting-methods) | Standard limit | Advanced Shopify limit | Shopify Plus limit | Shopify for enterprise (Commerce Components) |
| --- | --- | --- | --- | --- | --- |
| [GraphQL Admin API](https://shopify.dev/docs/api/admin-graphql) | Calculated query cost | 100 points/second | 200 points/second | 1000 points/second | 2000 points/second |
| [Storefront API](https://shopify.dev/docs/api/storefront) | None | None | None | None | None |
| [Payments Apps API](https://shopify.dev/docs/api/payments-apps/) ([GraphQL](https://shopify.dev/docs/api/payments-apps)) | Calculated query cost | 27300 points/second | 27300 points/second | 54600 points/second | 109200 points/second |
| [Customer Account API](https://shopify.dev/docs/api/customer) | Calculated query cost | 100 points/second | 200 points/second | 200 points/second | 400 points/second |

# Scopes and Permissions

- Depending on how you're [distributing your app](https://ampersand.slab.com/posts/shopify-connector-cd708bwy#h5vvp-app-distribution), you might need to request certain permissions or access scopes when users install your app.
- With a few APIs, you’ll need to request access from Shopify and be approved before you can start making calls.

# Provider documentation 

Shopify API Documentation [link](https://shopify.dev/docs/api)

# Provider app creation

There are different types of Shopify apps:

| **App Type** | **Detail** |
| --- | --- |
| **Embedded App** | An app that is designed to run within the Shopify admin dashboard using an iframe and **must authenticate incoming requests using session tokens**. For token acquisition, embedded apps **should use the recommended Token Exchange flow** for an improved user experience. |
| **Non-embedded App** | An app that runs outside the Shopify admin (e.g., a standalone website or mobile app) and must **implement its own authentication method** for incoming requests. These apps acquire the necessary access tokens using the **Authorization Code Grant flow**. |
| **Admin-created Custom App** | An app built specifically for a single merchant's store, created and **authenticated directly within the Shopify admin**. The access token is typically **generated upon app generation** in the Shopify admin, bypassing OAuth flows. |

Starting January 1, 2026, Shopify is **deprecating** merchant-created (**Admin-created**) custom apps where merchants can create app, choose scope and get the access to the token tied to that app.

[https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/generate-app-access-tokens-admin](https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/generate-app-access-tokens-admin)

We will be using **Non-embedded App**  to get the access token with given scope permissions that will be used to access the APIs in the Ampersand.

Shopify provides two methods of creating the OAuth app:

- Using shopify CLI
- Manual / UI screen

We will be using the manual/UI approach for this purpose.

- Navigate to shopify [partners portal](https://partners.shopify.com) and go to app distribution:

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/lPVnXKNObB3hlzVgiggLNNcp.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

- And click **Visit Dev Dashboard** that will redirect you to a new page listing all the apps:

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/UI0SG4VLDuVr7TOC5PtPCsDr.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

- Click on the **Create App** 

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/r7EW1X6yqbCeAAI92KcDDH2H.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

Give your app name and click create, it will open a full form to fill app information

- Select access **scope** and **optional scopes** from the list.
- In the **Redirect URIs** input field, enter Ampersand Redirect URLs (All of them) as a comma-separated list.

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/AFHparqLQ0fq5m7oBLdprhlI.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/wP20cY2rqgXspqRdjyefN59T.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/v8bthKETQuRfrTA65UqG0xLJ.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

- Finally click **Release** and it will ask for the release version name and message (optional)

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/YaF6tOHThmXFnvhCKn2xo6uD.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/ggVjM7wHmyGqRnpX9CGcrzoz.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

- Go to **Home** page and from the right side sections click on the **Distribution** and select distribution method.

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/dgerG9RVztstZABYhSNcDngv.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/qiNcVNNNb7IO834K0fQJ9xIk.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

- Go to settings page from **Dev** dashboard page and access **Client ID** and **Client Secret**

![](https://slabstatic.com/prod/assets/bcf9q7xo/post/cd708bwy/preimages/KgE54Y6J7Sq7MFEiMwZF1MW6.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

# Authentication types

All Shopify apps, other than apps created in the Shopify admin, needs to obtain authorization using the OAuth 2.0 specification to use Shopify’s API resources.

**Using the OAuth Authorization code grant flow:**

## Prerequisites (One-Time)

Before starting, you must already have:

1. A **Shopify App** created (via Partner Dashboard)
1. From the app settings:
    - **Client ID (API Key)**
    - **Client Secret**
1. A **Redirect URI** configured in the app
    - Example:

```
https://example.com/callback
```

    - (It does NOT need to be live; it just needs to match exactly)
1. The **Shop name**

```
{store-name}.myshopify.com
```

**TODO**

Update steps as client now need to generate the token using the Postman API calls.

This flow is defined for the custom backend service.

**Step 1:**

When a user installs the app through the Shopify App Store or using an installation link, your app receives a `GET` request to the **App URL** path that you specify in the Partner Dashboard. The request includes the `shop`, `timestamp`, and `hmac` query parameters. You need to verify the authenticity of these requests using the provided `hmac` parameter.

To verify the request, you need to remove the `hmac` parameter from the query string and process it through an HMAC-SHA256 hash function. For a request to be valid, the `hmac` parameter must match the HMAC-SHA256 hash of the remaining parameters in the query string. These parameters are subject to change, so don't hard code them inside your verification code.

Example:

Before HMAC removal

```bash
"code=0907a61c0c8d55e99db179b68161bc00
&hmac=700e2dadb827fcc8609e9d5ce208b2e9cdaab9df07390d2cbca10d7c328fc4bf
&shop={shop}.myshopify.com
&state=0.6784241404160823&timestamp=1337178173"
```

After HMAC removal

```bash
"code=0907a61c0c8d55e99db179b68161bc00
&shop={shop}.myshopify.com
&state=0.6784241404160823
&timestamp=1337178173"
```

Remember the remaining parameters must be sorted alphabetically as strings, in the format `"parameter_name=parameter_value"`.

We can process the string through an `HMAC-SHA256` hash function using your client secret. The message is authentic if the generated hexdigest is equal to the value of the `hmac` parameter.

**Step 2:**

Requesting the authorization code. This step involves redirecting the user to the consent screen and obtaining the code. Note that the app should redirect the user to the consent screen when:

- App doesn't have a token for that shop.
- App uses online tokens and the token for that shop has expired.
- App has a token for that shop, but it was created before you rotated the app's secret
- App has a token for that shop, but your app now requires scopes that differ from the scopes granted with that token

If your app is never embedded, then perform a 3xx redirect to the grant screen.

- If your app can be embedded:
    - Check whether the app is being rendered in an iframe by checking the `embedded` parameter.
    - If the app is being rendered in an iframe, then escape the iframe using a Shopify App Bridge redirect action that redirects back to the same URL.
    - Perform a 3xx redirect to the grant screen.

OAuth details:

- AuthURL: https://{shop}.myshopify.com/admin/oauth/authorize
- TokenURL:  https://{shop}.myshopify.com/admin/oauth/access_token

For the Authorizaion the following query parameters are expected:

Shopify's **OAuth2 flow** provides two types of access tokens: **Online-Token** & **Offline-Token**. _Online access tokens_ expire every 24 hours, and the app needs to re-authenticate the user (as no refresh token is provided). _Offline access tokens_ are like API keys — they never expire, and that is what we will be using for the Shopify connector.

**Sopify docs:**

[https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/authorization-code-grant](https://shopify.dev/docs/apps/build/authentication-authorization/access-tokens/authorization-code-grant)

**Community questions explaining refresh token issue:**

[https://community.shopify.com/t/can-not-refresh-access-tokens-of-our-users-via-admin-api-for-our-nonembedded-puclic-app/274314](https://community.shopify.com/t/can-not-refresh-access-tokens-of-our-users-via-admin-api-for-our-nonembedded-puclic-app/274314)

[https://community.shopify.com/t/why-shopify-authentication-does-provide-the-refresh-token/99577](https://community.shopify.com/t/why-shopify-authentication-does-provide-the-refresh-token/99577)

| Query parameter | Description |
| --- | --- |
| `{shop}` | The name of the user's shop. |
| `{client_id}` | The client ID for the app. |
| `{scopes}` | A comma-separated list of scopes. For example, to write orders and read customers, use `scope=write_products,read_shipping`. You should include every scope your app needs, regardless of any previously requested scopes. Any permission to write a resource includes the permission to read it. Be careful about what scopes you request. Some data is considered protected customer data, and places additional requirements on your app. For more information, refer to [Protected customer data](https://shopify.dev/docs/apps/store/data-protection/protected-customer-data). This parameter should be omitted if you've pushed requested API access scopes with the [`TOML` file](https://shopify.dev/docs/apps/structure#root-configuration-files). |
| `{redirect_uri}` | The URL to which a user is redirected after authorizing the app. The complete URL specified here must be added to your app as an allowed redirection URL, as defined in the Partner Dashboard. |
| `{nonce}` | A randomly selected value provided by your app that is unique for each authorization request. During the OAuth callback, your app must check that this value matches the one you provided during authorization. This mechanism is important for [the security of your app](https://tools.ietf.org/html/rfc6819#section-3.6). |
| `{access_mode}` | Sets the access mode. For an [online access token](https://shopify.dev/docs/apps/auth/access-token-types/online), set to `per-user`. For an [offline access token](https://shopify.dev/docs/apps/auth/access-token-types/offline) omit this parameter. |

Not that when we use the offline access token, the token does not expire unless revoked. No need to refresh the token.

**step 3:**

If you aren't using a library, then make sure that you verify the following:

- The `nonce` is the same one that your app provided to Shopify when [asking for permission](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started#ask-for-permission). Additionally, the signed cookie that you set when [asking for permission](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started#ask-for-permission) is present and its value equals the `nonce` value in the `state` parameter.
- The `hmac` is valid and [signed by Shopify](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started#step-1-verify-the-installation-request).
- The `shop` parameter is a valid shop hostname, ends with `myshopify.com`, and doesn't contain characters other than letters (a-z), numbers (0-9), periods, and hyphens.

You can use a regular expression to confirm that the hostname is valid. In the following example, the regular expression matches the hostname form of `https://{shop}.myshopify.com/`:

`/^https?\:\/\/[a-zA-Z0-9][a-zA-Z0-9\-]*\.myshopify\.com\/?/`

**step 4:**

If all the security check passes, you can now exchange the code for an access token by sending the request to the tokenURL. The following parameters must be provided on the request

| Parameter | Description |
| --- | --- |
| `client_id` | The client ID for the app, as defined in the Partner Dashboard. |
| `client_secret` | The client secret for the app, as defined in the Partner Dashboard. |
| `code` | The authorization code provided in the redirect. |

**Making Authenticated  Requests:**

Ideally **OAuth2** flow returns a refresh token for the first time or when the permissions are revoked for that application, which is then used to get the new authentication token after it gets expired. However, in Shopify's OAuth2 flow we don't receive it. So, in the case of OAuth2 with an **online-token**, if the token expires, we need to re-authenticate the user to regain permissions.

After obtaining an API access token, we can make authenticated requests to the [Admin API](https://shopify.dev/docs/api/admin). These requests are accompanied with a header `X-Shopify-Access-Token: {access_token}` where `{access_token}` is replaced with the access token.

**Scopes**

An app can request authenticated or unauthenticated access scopes.

| Type of access scopes | Description | Example use cases |
| --- | --- | --- |
| [Authenticated](https://shopify.dev/docs/api/usage/access-scopes#authenticated-access-scopes) | Controls access to resources in the [REST Admin API](https://shopify.dev/docs/api/admin-rest), [GraphQL Admin API](https://shopify.dev/docs/api/admin-graphql), and [Payments Apps API](https://shopify.dev/docs/api/payments-apps).<br>Authenticated access is intended for interacting with a store on behalf of a user. | - Creating products<br>- Managing discount codes |
| [Unauthenticated](https://shopify.dev/docs/api/usage/access-scopes#unauthenticated-access-scopes) | Controls an app's access to [Storefront API](https://shopify.dev/docs/api/storefront) objects.<br>Unauthenticated access is intended for interacting with a store on behalf of a customer. | - Viewing products<br>- Initiating a checkout |
| [Customer](https://shopify.dev/docs/api/usage/access-scopes#customer-access-scopes) | Controls an app's access to [Customer Account API](https://shopify.dev/docs/api/customer) objects.<br>Customer access is intended for interacting with data that belongs to a customer. | - Viewing orders<br>- Updating customer details |

**Authenticated access scopes**

An app can request the following authenticated access scopes:

| Scope | Access |
| --- | --- |
| `read_all_orders` | All relevant [orders](https://shopify.dev/docs/api/admin-graphql/latest/objects/Order) rather than the default window of orders created within the last 60 dayspermissions required<br>This access scope is used in conjunction with existing order scopes, for example `read_orders` or `write_orders`.<br>You need to [request permission for this access scope](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started#orders-permissions) from your Partner Dashboard before adding it to your app. |
| `read_assigned_fulfillment_orders`,<br>`write_assigned_fulfillment_orders` | [FulfillmentOrder](https://shopify.dev/docs/api/admin-rest/latest/resources/assignedfulfillmentorder) resources assigned to a location managed by your [fulfillment service](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillmentservice) |
| `read_cart_transforms`,<br>`write_cart_transforms` | Manage [Cart Transform](https://shopify.dev/docs/api/admin-graphql/unstable/objects/CartTransform) objects to sell [bundles.](https://shopify.dev/docs/apps/selling-strategies/bundles/add-a-customized-bundle) |
| `read_checkouts`,<br>`write_checkouts` | [Checkouts](https://shopify.dev/docs/api/admin-rest/latest/resources/checkout) |
| `read_checkout_branding_settings`,<br>`write_checkout_branding_settings` | [Checkout branding](https://shopify.dev/docs/api/admin-graphql/latest/queries/checkoutBranding) |
| `read_content`,<br>`write_content` | [Article](https://shopify.dev/docs/api/admin-rest/latest/resources/article), [Blog](https://shopify.dev/docs/api/admin-rest/latest/resources/blog), [Comment](https://shopify.dev/docs/api/admin-rest/latest/resources/comment), [Page](https://shopify.dev/docs/api/admin-rest/latest/resources/page), [Redirects](https://shopify.dev/docs/api/admin-rest/latest/resources/redirect), and [Metafield Definitions](https://shopify.dev/docs/api/admin-graphql/latest/queries/metafieldDefinitions) |
| `read_customer_merge`,<br>`write_customer_merge` | [CustomerMergePreview](https://shopify.dev/docs/api/admin-graphql/unstable/objects/CustomerMergePreview) and [CustomerMergeRequest](https://shopify.dev/docs/api/admin-graphql/unstable/objects/CustomerMergeRequest) |
| `read_customers`,<br>`write_customers` | [Customer](https://shopify.dev/docs/api/admin-rest/latest/resources/customer) and [Saved Search](https://shopify.dev/docs/api/admin-graphql/latest/objects/savedsearch) |
| `read_customer_payment_methods` | [CustomerPaymentMethod](https://shopify.dev/docs/api/admin-graphql/latest/objects/customerpaymentmethod)permissions required<br>You need to [request permission for this access scope](https://shopify.dev/docs/api/usage/access-scopes#subscription-apis-permissions) from your Partner Dashboard before adding it to your app. |
| `read_discounts`,<br>`write_discounts` | GraphQL Admin API [Discounts features](https://shopify.dev/docs/apps/selling-strategies/discounts/) |
| `read_draft_orders`,<br>`write_draft_orders` | [Draft Order](https://shopify.dev/docs/api/admin-rest/latest/resources/draftorder) |
| `read_files`,<br>`write_files` | GraphQL Admin API [GenericFile](https://shopify.dev/docs/api/admin-graphql/latest/objects/genericfile) object and [fileCreate](https://shopify.dev/docs/api/admin-graphql/latest/mutations/filecreate), [fileUpdate](https://shopify.dev/docs/api/admin-graphql/latest/mutations/fileupdate), and [fileDelete](https://shopify.dev/docs/api/admin-graphql/latest/mutations/filedelete) mutations |
| `read_fulfillments`,<br>`write_fulfillments` | [Fulfillment Service](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillmentservice) |
| `read_gift_cards`,<br>`write_gift_cards` | [Gift Card](https://shopify.dev/docs/api/admin-rest/latest/resources/gift-card) |
| `read_inventory`,<br>`write_inventory` | [Inventory Level](https://shopify.dev/docs/api/admin-rest/latest/resources/inventorylevel) and [Inventory Item](https://shopify.dev/docs/api/admin-rest/latest/resources/inventoryitem) |
| `read_legal_policies` | GraphQL Admin API [Shop Policy](https://shopify.dev/docs/api/admin-graphql/latest/objects/shoppolicy) |
| `read_locales`,<br>`write_locales` | GraphQL Admin API [Shop Locale](https://shopify.dev/docs/api/admin-graphql/latest/objects/shoplocale) |
| `write_locations` | GraphQL Admin API [locationActivate](https://shopify.dev/docs/api/admin-graphql/latest/mutations/locationActivate), [locationAdd](https://shopify.dev/docs/api/admin-graphql/latest/mutations/locationAdd), [locationDeactivate](https://shopify.dev/docs/api/admin-graphql/latest/mutations/locationDeactivate), [locationDelete](https://shopify.dev/docs/api/admin-graphql/latest/mutations/locationDelete), and [locationEdit](https://shopify.dev/docs/api/admin-graphql/latest/mutations/locationEdit) mutations. |
| `read_locations` | [Location](https://shopify.dev/docs/api/admin-rest/latest/resources/location) |
| `read_markets`,<br>`write_markets` | [Market](https://shopify.dev/docs/api/admin-graphql/latest/objects/market) |
| `read_metaobject_definitions`,<br>`write_metaobject_definitions` | [MetaobjectDefinition](https://shopify.dev/docs/api/admin-graphql/latest/objects/metaobjectdefinition) |
| `read_metaobjects`,<br>`write_metaobjects` | [Metaobject](https://shopify.dev/docs/api/admin-graphql/latest/objects/metaobject) |
| `read_marketing_events`,<br>`write_marketing_events` | [Marketing Event](https://shopify.dev/docs/api/admin-rest/latest/resources/marketingevent) |
| `read_merchant_approval_signals` | [MerchantApprovalSignals](https://shopify.dev/docs/api/admin-graphql/latest/objects/merchantapprovalsignals) |
| `read_merchant_managed_fulfillment_orders`,<br>`write_merchant_managed_fulfillment_orders` | [FulfillmentOrder](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillmentorder) resources assigned to merchant-managed locations |
| `read_orders`,<br>`write_orders` | [Abandoned checkouts](https://shopify.dev/docs/api/admin-rest/latest/resources/abandoned-checkouts), [Customer](https://shopify.dev/docs/api/admin-rest/latest/resources/customer), [Fulfillment](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillment), [Order](https://shopify.dev/docs/api/admin-rest/latest/resources/order), and [Transaction](https://shopify.dev/docs/api/admin-rest/latest/resources/transaction) resources |
| `read_payment_mandate`,<br>`write_payment_mandate` | [PaymentMandate](https://shopify.dev/docs/api/admin-graphql/latest/objects/PaymentMandate) |
| `read_payment_terms`,<br>`write_payment_terms` | GraphQL Admin API [PaymentSchedule](https://shopify.dev/docs/api/admin-graphql/latest/objects/paymentschedule) and [PaymentTerms](https://shopify.dev/docs/api/admin-graphql/latest/objects/paymentterms) objects |
| `read_price_rules`,<br>`write_price_rules` | [Price Rules](https://shopify.dev/docs/api/admin-rest/latest/resources/pricerule) |
| `read_products`,<br>`write_products` | [Product](https://shopify.dev/docs/api/admin-rest/latest/resources/product), [Product Variant](https://shopify.dev/docs/api/admin-rest/latest/resources/product-variant), [Product Image](https://shopify.dev/docs/api/admin-rest/latest/resources/product-image), [Collect](https://shopify.dev/docs/api/admin-rest/latest/resources/collect), [Custom Collection](https://shopify.dev/docs/api/admin-rest/latest/resources/customcollection), and [Smart Collection](https://shopify.dev/docs/api/admin-rest/latest/resources/smartcollection) |
| `read_product_listings` | [Product Listing](https://shopify.dev/docs/api/admin-rest/latest/resources/productlisting) and [Collection Listing](https://shopify.dev/docs/api/admin-rest/latest/resources/collectionlisting) |
| `read_publications`,<br>`write_publications` | [Product publishing](https://shopify.dev/docs/api/admin-graphql/latest/mutations/productpublish) and [Collection publishing](https://shopify.dev/docs/api/admin-graphql/latest/mutations/collectionpublish) |
| `read_purchase_options`,<br>`write_purchase_options` | [SellingPlan](https://shopify.dev/docs/api/admin-graphql/latest/objects/SellingPlan) |
| `read_reports`,<br>`write_reports` | [Reports](https://shopify.dev/docs/api/admin-rest/latest/resources/report) |
| `read_resource_feedbacks`,<br>`write_resource_feedbacks` | [ResourceFeedback](https://shopify.dev/docs/api/admin-rest/latest/resources/resourcefeedback) |
| `read_script_tags`,<br>`write_script_tags` | [Script Tag](https://shopify.dev/docs/api/admin-rest/latest/resources/scripttag) |
| `read_shipping`,<br>`write_shipping` | [Carrier Service](https://shopify.dev/docs/api/admin-rest/latest/resources/carrierservice), [Country](https://shopify.dev/docs/api/admin-rest/latest/resources/country), and [Province](https://shopify.dev/docs/api/admin-rest/latest/resources/province) |
| `read_shopify_payments_disputes` | Shopify Payments [Dispute](https://shopify.dev/docs/api/admin-rest/latest/resources/dispute) resource |
| `read_shopify_payments_payouts` | Shopify Payments [Payout](https://shopify.dev/docs/api/admin-rest/latest/resources/payout), [Balance](https://shopify.dev/docs/api/admin-rest/latest/resources/balance), and [Transaction](https://shopify.dev/docs/api/admin-rest/latest/resources/transaction) resources |
| `read_store_credit_accounts` | [StoreCreditAccount](https://shopify.dev/docs/api/admin-graphql/unstable/objects/StoreCreditAccount) |
| `read_store_credit_account_transactions`,<br>`write_store_credit_account_transactions` | [StoreCreditAccountDebitTransaction](https://shopify.dev/docs/api/admin-graphql/unstable/objects/StoreCreditAccountDebitTransaction) and [StoreCreditAccountCreditTransaction](https://shopify.dev/docs/api/admin-graphql/unstable/objects/StoreCreditAccountCreditTransaction) |
| `read_own_subscription_contracts`,<br>`write_own_subscription_contracts` | [SubscriptionContract](https://shopify.dev/docs/api/admin-graphql/latest/objects/SubscriptionContract)permissions required<br>You need to [request permission for these access scopes](https://shopify.dev/docs/api/usage/access-scopes#subscription-apis-permissions) from your Partner Dashboard before adding them to your app. |
| `read_returns`,<br>`write_returns` | [Return](https://shopify.dev/docs/api/admin-graphql/unstable/objects/Return) object |
| `read_themes`,<br>`write_themes` | [Asset](https://shopify.dev/docs/api/admin-rest/latest/resources/asset) and [Theme](https://shopify.dev/docs/api/admin-rest/latest/resources/theme) |
| `read_translations`,<br>`write_translations` | GraphQL Admin API [Translatable](https://shopify.dev/docs/api/admin-graphql/latest/queries/translatableresource) object |
| `read_third_party_fulfillment_orders`,<br>`write_third_party_fulfillment_orders` | [FulfillmentOrder](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillmentorder) resources assigned to a location managed by any [fulfillment service](https://shopify.dev/docs/api/admin-rest/latest/resources/fulfillmentservice) |
| `read_users` | [User](https://shopify.dev/docs/api/admin-rest/latest/resources/user) and [StaffMember](https://shopify.dev/docs/api/admin-graphql/latest/objects/staffmember)shopify plus |
| `read_order_edits`,<br>`write_order_edits` | GraphQL Admin API [OrderStagedChange](https://shopify.dev/docs/api/admin-graphql/latest/unions/OrderStagedChange) types and [order editing](https://shopify.dev/docs/apps/fulfillment/order-management-apps/order-editing) features |
| `write_payment_gateways` | Payments Apps API [paymentsAppConfigure](https://shopify.dev/docs/api/payments-apps/latest/mutations/paymentsAppConfigure) |
| `write_payment_sessions` | Payments Apps API [Payment](https://shopify.dev/docs/api/payments-apps/latest/objects/PaymentSession), [Capture](https://shopify.dev/docs/api/payments-apps/latest/objects/CaptureSession), [Refund](https://shopify.dev/docs/api/payments-apps/latest/objects/RefundSession) and [Void](https://shopify.dev/docs/api/payments-apps/latest/objects/VoidSession) |
| `write_pixels`,<br>`read_customer_events` | [Web Pixels API](https://shopify.dev/docs/api/web-pixels-api) |
| `write_privacy_settings`,<br>`read_privacy_settings` | [Privacy API](https://shopify.dev/docs/api/admin-graphql/unstable/mutations/dataSaleOptOut) |
| `read_validations`,<br>`write_validations` | GraphQL Admin API [`Validation`](https://shopify.dev/docs/api/admin-graphql/latest/objects/Validation) object |

**Orders permissions**

By default, you have access to the last 60 days' worth of orders for a store. To access all the orders, you need to request access to the `read_all_orders` scope from the user:

1. From the Partner Dashboard, go to **Apps**.
1. Click the name of your app.
1. Click **API access**.
1. In the **Access requests** section, on the **Read all orders scope** card, click **Request access**.
1. On the **Orders** page that opens, describe your app and why you’re applying for access.
1. Click **Request access**.

If Shopify approves the request, then you can add the `read_all_orders` scope to your app along with `read_orders` or `write_orders`.

**Subscription APIs permissions**

Subscription apps let users sell subscription products that generate multiple orders on a specific billing frequency.

With subscription products, the app user isn't required to get customer approval for each subsequent order after the initial subscription purchase. As a result, your app needs to request the required protected access scopes to use Subscription APIs from the app user:

1. From the Partner Dashboard, go to **Apps**.
1. Click the name of your app.
1. Click **API access**.
1. In the **Access requests** section, on the **Access Subscriptions APIs** card, click **Request access**.
1. On the **Subscriptions** page that opens, describe why you’re applying for access.
1. Click **Request access**.

If Shopify approves your request, then you can add the `read_customer_payment_methods` and `write_own_subscription_contracts` scopes to your app.

**Protected customer data permissions**

By default, apps don't have access to any protected customer data. To access protected customer data, you must meet our [protected customer data requirements](https://shopify.dev/docs/apps/store/data-protection/protected-customer-data#requirements). You can add the relevant scopes to your app, but the API won't return data from non-development stores until your app is configured and approved for protected customer data use.

Example of installing the app:

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/-gQVQT7cOnvDRXFLku22-fhM.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

Raw token form, generated by the script.

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/G5OrRDJ3cGbpnQcx_VSEJ1fG.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

# List of API modules, objects or API categories

Shopify APIs are categorized as follows:

1. Apps

- GraphQL Admin API
- Rest Admin API
- Partners API
- Payment API
- Shopify functions API

2. Custom Storepoints

- Storefront API
- Customer Account API

3.Themes

- Ajax API
- Liquid

The Admin API is the core of the shopify APIs.And it is available in both GraphQL and Rest.It provides data on products, customers, orders, inventory, fulfilment and more.

Some newer platform features may only be available in GraphQL.

**Details about the Admin API**

- The Admin API supports both GraphQL and REST.
- This is a versioned API. Updates are released quarterly and supported API versions are listed in the [release notes](https://shopify.dev/docs/api/release-notes).
- Apps must explicitly request the relevant access scopes from the user during installation.
- Apps must authenticate to interact with the Admin API.
- The Admin API enforces rate limits on all requests. Note that there are different rate-limiting methods for GraphQL and REST. All apps connecting to the Admin API are subject to Shopify’s API Terms of Service.

## GraphQL Admin API

Public and custom apps created in the Partner Dashboard generate tokens using OAuth, and custom apps made in the Shopify admin are authenticated in the Shopify admin.

Include your token as a `X-Shopify-Access-Token` header on all API queries. Using Shopify’s supported client libraries can simplify this process.

GraphQL queries are executed by sending POST HTTP requests to the endpoint:

[`https://{store_name}.myshopify.com/admin/api/2024-04/graphql.json`](https://{store_name}.myshopify.com/admin/api/2024-04/graphql.json)

Queries begin with one of the objects listed under QueryRoot. The QueryRoot is the schema’s entry-point for queries.

Queries are equivalent to making a GET request in REST. The example shown is a query to get the ID and title of the first three products.

An example of GraphQL request.

```bash
# Get the ID and title of the three most recently added products
curl -X POST   https://{store_name}.myshopify.com/admin/api/2024-04/graphql.json \
  -H 'Content-Type: application/json' \
  -H 'X-Shopify-Access-Token: {access_token}' \
  -d '{
  "query": "{
    products(first: 3) {
      edges {
        node {
          id
          title
        }
      }
    }
  }"
}'
```

**Status And Error Codes**

GraphQL HTTP status codes are different from REST API status codes. Most importantly, the GraphQL API can return a `200 OK` response code in cases that would typically produce 4xx or 5xx errors in REST.

**Error handling**

The response for the errors object contains additional detail to help you debug your operation.

The response for mutations contains additional detail to help debug your query. To access this, you must request `userErrors`.

## REST Admin API

Admin REST API endpoints are organized by resource type. You’ll need to use different endpoints depending on your app’s requirements.

All Admin REST API endpoints follow this pattern:

`https://{store_name}.myshopify.com/admin/api/2024-04/{resource}.json`

The BaseURL:  [https://{store_name}.myshopify.com](https://{store_name}.myshopify.com/admin)

All REST endpoints support cursor-based pagination.

The Resources list below:

| **Resouce** | **Endpoint** |
| --- | --- |
| Access | /admin/oauth/access_scopes.json |
| Inventory | /admin/api/2024-04/inventory_items/{inventory_id}.json |
| MarketingEvents | /admin/api/2024-04/marketing_events.json |
| Order | /admin/api/2024-04/orders.json |
| Transaction | /admin/api/2024-04/orders/{order_id}/transactions.json |
| Product | /admin/api/2024-04/products.json |
| Payment | admin/api/2024-04/checkouts/{token}/payments.json |
| Fullfilment | /admin/api/2024-04/fulfillments.json |
| TenderTransaction | /admin/api/2024-04/tender_transactions.json |
| Webhook | /admin/api/2024-04/webhooks.json |

# Rate limiting

Shopify APIs use several different rate-limiting methods. They’re described in more detail below, but these are the key figures in brief:

| API | Rate-limiting method | Standard limit | Advanced Shopify limit | Shopify Plus limit | Shopify for enterprise (Commerce Components) |
| --- | --- | --- | --- | --- | --- |
| [Admin API](https://shopify.dev/docs/api/admin) ([GraphQL](https://shopify.dev/docs/api/admin-graphql)) | Calculated query cost | 100 points/second | 200 points/second | 1000 points/second | 20X |
| [Admin API](https://shopify.dev/docs/api/admin) ([REST](https://shopify.dev/docs/api/admin-rest)) | Request-based limit | 2 requests/second | 4 requests/second | 20 requests/second | 20X |
| [Storefront API](https://shopify.dev/docs/api/storefront) | None | None | None | None | None |
| [Payments Apps API](https://shopify.dev/docs/api/payments-apps) ([GraphQL](https://shopify.dev/docs/api/payments-apps)) | Calculated query cost | 910 points/second | 910 points/second | 1820 points/second | 20X |
| [Customer Account API](https://shopify.dev/docs/api/customer) | Calculated query cost | 100 points/second | 200 points/second | 200 points/second | 20X |

All Shopify APIs use a leaky bucket algorithm to manage requests.

## Rate limiting methods

Shopify uses two different methods for managing rate limits. Different APIs use different methods depending on use case, so it's useful to understand the various types of rate limits your apps will encounter:

### Request-based limits

Apps can make a maximum **number** of requests per minute.

This method is used by the REST Admin API.

### Calculated query costs

Apps can make requests that cost a maximum number of **points** per minute.

This method is used by the GraphQL API

## GraphQL Admin API rate limits

Calls to the GraphQL Admin API are limited based on calculated query costs, which means you should consider the _cost_ of requests over time, rather than the _number_ of requests.

GraphQL Admin API rate limits are based on the combination of the app and store. This means that calls from one app don't affect the rate limits of another app, even on the same store. Similarly, calls to one store don't affect the rate limits of another store, even from the same app.

Each combination of app and store is given a bucket size and restore rate based on API and plan tier. By making simpler, lower-cost queries, you can maximize your throughput and make more queries over time.

### Cost calculation

Every field in the schema has an integer cost value assigned to it. The cost of a query is the maximum of possible fields selected. Running a query is the best way to find out its true cost.

By default, a field's cost is based on what the field returns:

| **Field returns** | **Cost value** |
| --- | --- |
| Scalar | 0 |
| Enum | 0 |
| Object | 1 |
| Interface | Maximum of possible selections |
| Union | Maximum of possible selections |
| Connection | Sized by `first` and `last` arguments |
| Mutation | 10 |



Although these default costs are in place, Shopify also reserves the right to set manual costs on fields.

### Requested and actual cost

Shopify calculates the cost of a query both before and after execution.

- The **requested cost** is based on the composition of fields selected in the request.
- The **actual cost** is based on the query results, and may be lower than requested cost due to the actual objects returned or connections that return fewer edges than requested.

Rate limits use a combination of the requested and actual query cost. Before execution begins, an app’s bucket must have enough capacity for the requested cost of a query. When execution is complete, the bucket is refunded the difference between the requested cost and the actual cost of the query.

### Single query limit

A single query may not exceed a cost of 1,000 points, regardless of plan limits. This limit is enforced before a query is executed based on the query’s requested cost.

### Maximum input array size limit

Input arguments that accept an array have a maximum size of 250. Queries and mutations return an error if an input array exceeds 250 items.

### GraphQL response

The response includes information about the cost of the request and the state of the throttle. This data is returned under the `extensions` key:

```
"extensions": {
    "cost": {
      "requestedQueryCost": 101,
      "actualQueryCost": 46,
      "throttleStatus": {
        "maximumAvailable": 1000,
        "currentlyAvailable": 954,
        "restoreRate": 50
      }
    }
  }
```

To get a detailed breakdown of how each field contributes to the requested cost, you can include the header `'X-GraphQL-Cost-Include-Fields': true` in your request.

```
"extensions": {
    "cost": {
      "requestedQueryCost": 101,
      "actualQueryCost": 46,
      "throttleStatus": ...,
      "fields": [
        {
          "path": [
            "shop"
          ],
          "definedCost": 1,
          "requestedTotalCost": 101,
          "requestedChildrenCost": 100
        },
        ...
      ]
    }
  }
```

### Bulk operations

To query and fetch large amounts of data, you should use bulk operations instead of single queries. Bulk operations are designed for handling large amounts of data, and they don't have the max cost limits or rate limits that single queries have.



## REST Admin API rate limits

Calls to the REST Admin API are governed by request-based limits, which means you should consider the total _number_ of API calls your app makes. In addition, there are resource-based rate limits and throttles.

REST Admin API rate limits are based on the combination of the app and store. This means that calls from one app don't affect the rate limits of another app, even on the same store. Similarly, calls to one store don't affect the rate limits of another store, even from the same app.

Limits are calculated using the leaky bucket algorithm. All requests that are made after rate limits have been exceeded are throttled and an HTTP `429 Too Many Requests` error is returned. Requests succeed again after enough requests have emptied out of the bucket. You can see the current state of the throttle for a store by using the rate limits header.

The _bucket size_ and _leak rate_ properties determine the API’s burst behavior and request rate.

The default settings are as follows:

- **Bucket size**: `40 requests/app/store`
- **Leak rate**: `2/second`

The bucket size and leak rate is increased by a factor of 10 for [Shopify Plus stores](https://www.shopify.com/plus):

- **Bucket size**: `400 requests/app/store`
- **Leak rate**: `20/second`

If the bucket size is exceeded, then an HTTP `429 Too Many Requests` error is returned. The bucket empties at a leak rate of two requests per second. To avoid being throttled, you can build your app to average two requests per second. The throttle is a pass or fail operation. If there is available capacity in your bucket, then the request is executed without queueing or processing delays. Otherwise, the request is throttled.

There is an additional rate limit for GET requests. When the value of the `page` parameter results in an offset of over 100,000 of the requested resource, a `429 Too Many Requests` error is returned. For example, a request to `GET /admin/collects.json?limit=250&page=401` would generate an offset of 100,250 (250 x 401 = 100,250) and return a 429 response.

**Caution**

Page-based pagination was deprecated in the Admin API with version 2019-07. Use [cursor-based pagination](https://shopify.dev/docs/api/usage/pagination-rest) instead.

### Rate limits header

You can check how many requests you’ve already made using the Shopify `X-Shopify-Shop-Api-Call-Limit` header that was sent in response to your API request. This header lists how many requests you’ve made for a particular store. For example:

```xml
X-Shopify-Shop-Api-Call-Limit: 32/40
```

In this example, `32` is the current request count and `40` is the bucket size. The request count decreases according to the leak rate over time. For example, if the header displays `39/40` requests, then after a wait period of ten seconds, the header displays `19/40` requests.

### Retry-After header

When a request goes over a rate limit, a `429 Too Many Requests` error and a `Retry-After` header are returned. The `Retry-After` header contains the number of seconds to wait until you can make a request again. Any request made before the wait time has elapsed is throttled.

```xml
X-Shopify-Shop-Api-Call-Limit: 40/402  Retry-After: 2.0
```

# Batch endpoints

## Exports and Queries

Bulk operations are only available through the [GraphQL Admin API](https://shopify.dev/docs/api/admin-graphql). You can't perform bulk operations with the REST Admin API or the Storefront API.

Bulk Operations are: [`bulkOperationRunMutation`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunmutation)  [`bulkOperationRunQuery`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunquery) but you can also run `bulkOperationCancel`

**Limitations**

- You can run only one bulk operation of each type ([`bulkOperationRunMutation`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunmutation) or [`bulkOperationRunQuery`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunquery)) at a time per shop.
- The bulk query operation has to complete within 10 days. After that it will be stopped and marked as `failed`.
- When your query runs into this limit, consider reducing the query complexity and depth.

On waiting for the bulk opefration to finish, it is recommended to subscribe to an event rather than redundant API calls.

**Bulk query workflow**

Below is the high-level workflow for creating a bulk query:

1. Identify a potential bulk operation.
1. You can use a new or existing query, but it should potentially return a lot of data. Connection-based queries work best.
1. Test the query by using the Shopify GraphiQL app.
1. Write a new mutation document for `bulkOperationRunQuery`.
1. Include the query as the value for the `query` argument in the mutation.
1. Run the mutation.
1. Wait for the bulk operation to finish by either:
    1. Subscribing to a webhook topic that sends a webhook payload when the operation is finished.
    1. Polling the bulk operation until the `status` field shows that the operation is no longer running.
1. You can check the operation's progress using the `objectCount` field in `currentBulkOperation`.
1. Download the JSONL file at the URL provided in the `url` field.

**Rate limits**

You can run only one bulk operation of each type (`bulkOperationRunMutation` or `bulkOperationRunQuery`) at a time per shop. This limit is in place because operations are asynchronous and long-running. To run a subsequent bulk operation for a shop, you need to either cancel the running operation or wait for it to finish.

Bulk operations have some additional restrictions:

- Maximum of five total connections in the query.
- Connections must implement the `Node` interface
- The top-level `node` and `nodes` fields can't be used.
- Maximum of two levels deep for nested connections. For example, the following is invalid because there are three levels of nested connections:

## Import and Mutations

Using the GraphQL Admin API, you can bulk import large volumes of data asychronously. When the operation is complete, the results are delivered in a JSON Lines (JSONL) file that Shopify makes available at a URL.

**Limitations**

- You can run only one bulk operation of each type ([`bulkOperationRunMutation`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunmutation) or [`bulkOperationRunQuery`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/bulkoperationrunquery)) at a time per shop.
- The bulk mutation operation has to complete within 24 hours. After that it will be stopped and marked as `failed`.
- When your import runs into this limit, consider reducing the input size.
- You can supply only one of the supported GraphQL API mutations to the `bulkOperationRunMutation` at a time:
    - [`collectionCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/collectioncreate)
    - [`collectionUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/collectionupdate)
    - [`customerCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/customercreate)
    - [`customerUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/customerupdate)
    - [`customerPaymentMethodRemoteCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/customerpaymentmethodremotecreate)
    - [`giftCardCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/giftcardcreate)
    - [`giftCardUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/giftcardupdate)
    - [`marketingActivityUpsertExternal`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/marketingActivityUpsertExternal)
    - [`marketingEngagementCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/marketingEngagementCreate)
    - [`metafieldsSet`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/metafieldsset)
    - [`metaobjectUpsert`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/metaobjectupsert)
    - [`priceListFixedPricesAdd`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/pricelistfixedpricesadd)
    - [`priceListFixedPricesDelete`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/pricelistfixedpricesdelete)
    - [`privateMetafieldUpsert`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/privatemetafieldupsert)
    - [`productCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/productcreate)
    - [`productUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/productupdate)
    - [`productUpdateMedia`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/productupdatemedia)
    - [`productVariantUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/productvariantupdate)
    - [`publishablePublish`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/publishablePublish)
    - [`publishableUnpublish`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/publishableUnpublish)
    - [`publicationUpdate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/publicationUpdate)
    - [`subscriptionBillingAttemptCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/subscriptionbillingattemptcreate)
    - [`subscriptionContractAtomicCreate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/subscriptioncontractatomiccreate)
    - [`subscriptionContractProductChange`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/subscriptioncontractproductchange)
    - [`subscriptionContractSetNextBillingDate`](https://shopify.dev/docs/api/admin-graphql/latest/mutations/subscriptioncontractsetnextbillingdate)
- The mutation that's passed into `bulkOperationRunMutation` is limited to one connection field, which is defined by the GraphQL Admin API schema.
- The size of the JSONL file cannot exceed 20MB.

# Pagination

**Pagination with GraphQL**

You can retrieve up to a maximum of 250 resources. If you need to paginate larger volumes of data, then you can perform a bulk query operation using the GraphQL Admin API.

In the GraphQL Admin API, each connection returns a [`PageInfo`](https://shopify.dev/docs/api/admin-graphql/latest/objects/PageInfo) object that assists in cursor-based pagination. The `PageInfo` object is composed of the following fields:

| Field | Type | Description |
| --- | --- | --- |
| `hasPreviousPage` | Boolean | Whether there are results in the connection before the current page. |
| `hasNextPage` | Boolean | Whether there are results in the connection after the current page. |
| `startCursor` | string | The cursor of the first node in the `nodes` list. |
| `endCursor` | string | The cursor of the last node in the `nodes` list. |

All connections in Shopify's APIs provide forward pagination. This is achieved with the following connection variables:

| Field | Type | Description |
| --- | --- | --- |
| `first` | integer | The requested number of `nodes` for each page. |
| `after` | string | The cursor to retrieve `nodes` after in the connection. Typically, you should pass the `endCursor` of the previous page as `after`. |

Examples:

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/41ip0t-H1oneRIq3d949f3Pt.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/RV0RRb1zJufoCvc_YHeDk5pS.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

Some connections in Shopify's APIs also provide backward pagination. This is achieved with the following connection variables:

| Field | Type | Description |
| --- | --- | --- |
| `last` | integer | The requested number of `nodes` for each page. |
| `before` | string | The cursor to retrieve `nodes` before in the connection. Typically, you should pass the `startCursor` of the previous page as `before`. |

Example:

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/WNjAnvyfHfb2mH0kfqkUq-tH.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

![](https://slabstatic.com/prod/uploads/bcf9q7xo/posts/images/preload/I2TyQzXR0QjHnWPKMC1pqgMJ.png?jwt=eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJodHRwczovL3NsYWJzdGF0aWMuY29tIiwiZXhwIjoxNzY5MTcwOTQ5LCJpYXQiOjE3Njc5NjEzNDksImlzcyI6Imh0dHBzOi8vYXBwLnNsYWIuY29tIiwianRpIjoiMzI0Z3V1b2Fnb2xhbzBmM2R2cWppMmU0IiwibmJmIjoxNzY3OTYxMzQ5LCJwYXRoIjoicHJvZC9hc3NldHMvYmNmOXE3eG8vcG9zdC9jZDcwOGJ3eSJ9.JIKbwbmc9i8KddoVFlIeNOpPc3tzvySU_ToPsvUkt6FY3Kv9SO3CJtveVw_nH2iHsiyN_YzMmEyoQs7yWqYHSa9OAUzOVrNzqEC5Inng4PBBsxHEr3axOHtMcnuTvXZLxOF06DvgXB5za9r6s0NFmsO8PdFDON4RdKwPmyNzHnk3DJjUd4fA_KT1piF1zSEESwF746HOBPxZZG8bSSCPOVyr9OS1dugm8B9HMBu1Ua1uREVWDCX2fEo8d0GgtB2BuRqyDpmOOimx4MdfoubOcJiIv93THuP_K2r16uwDF1tP0o5ECsrzaMJTtQLxTK13MZnJqBv2v9xH4eudod7ITQ)

**Pagination with Rest API**

REST endpoints support cursor-based pagination. If the response is paginated  then a response header returns links to the next page and the previous page of results.The   `link` header will have the next page.

The URL in the link header can include the following parameters:

| Parameter | Description |
| --- | --- |
| `page_info` | A unique ID used to access a certain page of results. The `page_info` parameter can't be modified and must be used exactly as it appears in the link header URL. |
| `limit` | The maximum number of results to show on the page:<br>- The default `limit` value is `50`.<br>- The maximum `limit` value is `250`. |
| `fields` | A comma-separated list of which fields to show in the results. This parameter only works for some endpoints. |

## Limitations and considerations

- A request that includes the `page_info` parameter can't include any other parameters except for `limit` and `fields` (if it applies to the endpoint). If you want your results to be filtered by other parameters, then you need to include those parameters in the first request you make.
- The link header URLs are temporary and we don't recommend saving them to use later. Use link header URLs only while working with the request that generated them.
- Any request that sends the `page` parameter will return an error.

# Errors

| 400 Bad Request | The request wasn't understood by the server, generally due to bad syntax or because the `Content-Type` header wasn't correctly set to `application/json`.<br>This status is also returned when a [token exchange](https://shopify.dev/docs/apps/auth/get-access-tokens/token-exchange) request includes an expired or otherwise invalid session token.<br>This status is also returned when the request provides an invalid `code` parameter during [authorization code grant](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started). |
| --- | --- |
| 401 Unauthorized | The necessary [authentication credentials](https://shopify.dev/docs/apps/auth) are not present in the request or are incorrect. |
| 402 Payment Required | The requested shop is currently frozen. The shop owner needs to log in to the shop's admin and pay the outstanding balance to unfreeze the shop. |
| 403 Forbidden | The server is refusing to respond to the request. This status is generally returned if you haven't [requested the appropriate scope](https://shopify.dev/docs/apps/auth/get-access-tokens/authorization-code-grant/getting-started#ask-for-permission) for this action. |
| 404 Not Found | The requested resource was not found but could be available again in the future. |
| 405 Method Not Allowed | The server recognizes the request but rejects the specific HTTP method. This status is generally returned when a client-side error occurs. |
| 406 Not Acceptable | The requested resource is only capable of generating content not acceptable according to the Accept headers sent in the request. |
| 409 Resource Conflict | The requested resource couldn't be processed because of conflict in the request. For example, the requested resource might not be in an expected state, or processing the request would create a conflict within the resource. |
| 414 URI Too Long | The server is refusing to accept the request because the Uniform Resource Identifier (URI) provided was too long. |
| 415 Unsupported Media Type | The server is refusing to accept the request because the payload format is in an unsupported format. |
| 422 Unprocessable Entity | The request body was well-formed but contains semantic errors. A `422` error code can be returned from a variety of scenarios including, but not limited to:<br>- Incorrectly formatted input<br>- Checking out products that are out of stock<br>- Canceling an order that has fulfillments<br>- Creating an order with tax lines on both line items and the order<br>- Creating a customer without an email or name<br>- Creating a product without a title<br>The response body provides details in the `errors` or `error` parameters. |
| 423 Locked | The requested shop is currently locked. Shops are locked if they repeatedly exceed their API request limit, or if there is an issue with the account, such as a detected compromise or fraud risk.<br>[Contact support](https://help.shopify.com/en/questions#/contact) if your shop is locked. |
| 429 Too Many Requests | The request was not accepted because the application has exceeded the rate limit. Learn more about [Shopify’s API rate limits](https://shopify.dev/docs/api/usage/rate-limits). |
| 430 Shopify Security Rejection | The request was not accepted because the request might be malicious, and Shopify has responded by rejecting it to protect the app from any possible attacks. |
| 500 Internal Server Error | An internal error occurred in Shopify. Simplify or retry your request. If the issue persists, then please record any error codes, timestamps and [contact Partner Support](https://help.shopify.com/en/questions/partners) so that Shopify staff can investigate. |
| 501 Not Implemented | The requested endpoint is not available on that particular shop, e.g. requesting access to a Shopify Plus–only API on a non-Plus shop. This response may also indicate that this endpoint is reserved for future use. |
| 502 Bad Gateway | The server, while acting as a gateway or proxy, received an invalid response from the upstream server. A 502 error isn't typically something you can fix. It usually requires a fix on the web server or the proxies that you're trying to get access through. |
| 503 Service Unavailable | The server is currently unavailable. Check the [Shopify status page](https://www.shopifystatus.com/) for reported service outages. |
| 504 Gateway Timeout | The request couldn't complete in time. Shopify waits up to 10 seconds for a response. Try breaking it down in multiple smaller requests. |
| 530 Origin DNS Error | Cloudflare can't resolve the requested DNS record. Check the [Shopify status page](https://www.shopifystatus.com/) for reported service outages. |
| 540 Temporarily Disabled | The requested endpoint isn't currently available. It has been temporarily disabled, and will be back online shortly. |
| 783 Unexpected Token | The request includes a JSON syntax error, so the API code is failing to convert some of the data that the server returned. |



# Miscellanous

## App distribution

The way you choose to distribute your app depends on its purpose and your audience. **You can't change the distribution method after you select it,** so make sure that you understand the different capabilities and requirements of each type.

| Distribution model | Number of stores | App type | Authorization or authentication method | Approval required | Limitations |
| --- | --- | --- | --- | --- | --- |
| Public distribution | Can be installed on multiple Shopify stores | Public | If embedded, token exchange and session tokens<br>If not embedded, authorization code grant | Yes | Must sync certain data with Shopify |
| Custom distribution | Installed on a single Shopify store or on multiple stores that belong to the same Plus organization | Custom | If embedded, token exchange and session tokens<br>If not embedded, authorization code grant | No | Can't use the Billing API to charge merchants |
| Shopify admin | Installed on a single Shopify store | Custom | Authenticate in the Shopify admin | No | Can't use Shopify App Bridge to display as an embedded app<br>Can't use app extensions<br>Can't use the Billing API to charge merchants |

API Release [notes](https://shopify.dev/docs/api/release-notes)

Webhook [configuration ](https://shopify.dev/docs/apps/webhooks/configuration)
