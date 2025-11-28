# Braintree connector

## Supported Objects
Below is an exhaustive list of objects & methods supported on the objects

| Object            | Resource          | Method |
|-------------------|-------------------|--------|
| customers         | customers         | read   |
| transactions      | transactions      | read   |
| refunds           | refunds           | read   |
| disputes          | disputes          | read   |
| verifications     | verifications     | read   |
| merchant_accounts | merchant_accounts | read   |

## Incremental Read
Below objects support incremental read via `createdAt` filter:
- customers
- transactions
- refunds
- disputes
- verifications

Note: `merchant_accounts` does not support time-based filtering.

## Known Limitations
- **Incremental reads use `createdAt`**: Braintree's GraphQL Search API only exposes `createdAt` as a timestamp filter. Updates to existing records will not be captured via incremental reads.
- **`payment_methods` not readable**: Cannot be searched independently in Braintree's GraphQL API. Must be accessed via a Customer's paymentMethods connection.
