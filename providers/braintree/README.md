# Braintree connector

## Supported Objects
Below is an exhaustive list of objects & methods supported on the objects

| Object                          | Resource                        | Method |
|---------------------------------|---------------------------------|--------|
| customers                       | customers                       | read   |
| transactions                    | transactions                    | read   |
| refunds                         | refunds                         | read   |
| disputes                        | disputes                        | read   |
| verifications                   | verifications                   | read   |
| merchantAccounts                | merchantAccounts                | read   |
| inStoreLocations                | inStoreLocations                | read   |
| inStoreReaders                  | inStoreReaders                  | read   |
| businessAccountCreationRequests | businessAccountCreationRequests | read   |
| payments                        | payments                        | read   |

## Incremental Read
Below objects support incremental read via `createdAt` filter:
- customers
- transactions
- refunds
- disputes
- verifications
- payments

Note: The following objects do not support time-based filtering:
- `merchantAccounts`
- `inStoreLocations`
- `inStoreReaders`
- `businessAccountCreationRequests`

## Known Limitations
- **Incremental reads use `createdAt`**: Braintree's GraphQL Search API only exposes `createdAt` as a timestamp filter. Updates to existing records will not be captured via incremental reads.
- **`paymentMethods` not readable**: Cannot be searched independently in Braintree's GraphQL API. Must be accessed via a Customer's paymentMethods connection.
- **`payments` is a unified view**: The `payments` object returns both transactions and refunds. Consider using `transactions` and `refunds` separately for more specific queries.
