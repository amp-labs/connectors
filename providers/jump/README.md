# Jump connector

Jump exposes a single GraphQL endpoint: `https://my.jumpapp.com/enterprise/graphql`.
API reference (login required): https://my.jumpapp.com/enterprise/documentation

## Supported Objects
Below is an exhaustive list of objects & methods supported on the objects

| Object            | Resource                                              | Method      |
|-------------------|-------------------------------------------------------|-------------|
| contacts          | contacts, createContact, updateContact, deleteContact | read, write |
| documents         | documents                                             | read        |
| integrations      | integrations                                          | read        |
| meetingPreps      | meetingPreps                                          | read        |
| meetings          | meetings                                              | read        |
| notes             | notes                                                 | read        |
| pulses            | pulses                                                | read        |
| scorecards        | scorecards                                            | read        |
| signalDefinitions | signalDefinitions                                     | read        |
| tasks             | tasks                                                 | read        |
| users             | users                                                 | read        |

## Incremental Read
All objects support incremental read via the `updatedAfter` / `updatedBefore` query arguments.
Jump compares timestamps to the microsecond and both bounds are inclusive, so `Since` and `Until`
are sent with their fractional seconds.

## Write and Delete
Only `contacts` can be written, through the `createContact`, `updateContact` and `deleteContact` mutations.
The remaining mutations are actions rather than record writes: `uploadDocument` returns a signed upload URL
and creates the document asynchronously, `sendNotetaker` and `deactivateUser` act on meetings and users.

- Writes accept the mutation arguments `firstName`, `lastName` and `email`, which differ from the read fields:
  reads return the combined `name`. Record fields are sent as GraphQL variables.
- `createContact` requires `firstName`, `lastName` and `email`. `updateContact` only changes the fields provided.
- Account-scoped API keys must name the user each mutation acts for. The connector sends the `userId` connector
  metadata, collected at installation, with every mutation. Reads are not scoped to that user.
- Jump documents `updateContact` for Jump native contacts. `deleteContact` also disconnects the contact from
  linked meeting participants.

## Metadata
Jump does not publish an OpenAPI specification and disables GraphQL introspection for API keys.
Object metadata is generated from the published GraphQL schema, see `metadata/README.md`.
Nested relations (for example `meetings.transcript`) are only queried when requested.

## Query Complexity
Jump rejects queries above 1000 complexity points. A list query costs `first × (1 + selected fields)`,
and some fields have fixed costs (for example `Meeting.rawTranscript` is 20). The page size is
therefore computed per object from the requested fields, see `objects.go` and `read.go`.
The per-record costs were measured against the live API and must be updated if query templates change.

## Known Limitations
- **Contacts reject every pagination cursor**: the `contacts` query returns `INVALID_PAGINATION` for any
  `after` value, including the `endCursor` it returned itself. Contacts are returned in ascending
  `insertedAt` order, so subsequent pages are requested with the inclusive `insertedAfter` argument and
  records already returned are dropped by id. If more contacts than fit on a page share the same `insertedAt`,
  pagination stops with an error instead of looping.
- **`updateContact` does not change `updatedAt`**: a contact updated through the API keeps its previous
  `updatedAt`, so incremental reads of `contacts` do not return it.
- **`users` requires an account-scoped API key**: OAuth tokens cannot query users.
