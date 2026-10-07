# Wrike connector

Deep connector for the [Wrike API v4](https://developers.wrike.com/api/v4/).
Authentication is OAuth 2.0 (authorization code). The catalog base URL is
`https://www.wrike.com/api`; object paths carry the `/v4` prefix.

## Supported objects

All 27 objects are account-level lists whose response is Wrike's standard
envelope `{"kind": "...", "data": [...]}`; records live under `data`.

| Object | Resource path | Scopes |
|---|---|---|
| `access_roles` | `/v4/access_roles` | `amReadOnlyAccessRole` |
| `approvals` | `/v4/approvals` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `attachments` | `/v4/attachments` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `audit_log` | `/v4/audit_log` | `amReadOnlyAuditLog` |
| `bookings` | `/v4/bookings` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `comments` | `/v4/comments` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `contacts` | `/v4/contacts` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `custom_item_types` | `/v4/custom_item_types` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `customfields` | `/v4/customfields` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `folder_blueprints` | `/v4/folder_blueprints` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `folders` | `/v4/folders` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `groups` | `/v4/groups` | `amReadOnlyGroup`, `amReadWriteGroup` |
| `invitations` | `/v4/invitations` | `amReadOnlyInvitation`, `amReadWriteInvitation` |
| `jobroles` | `/v4/jobroles` | `amReadOnlyUser`, `amReadWriteUser` |
| `placeholders` | `/v4/placeholders` | `amReadOnlyUser`, `amReadWriteUser` |
| `request_forms` | `/v4/request_forms` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `spaces` | `/v4/spaces` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `task_blueprints` | `/v4/task_blueprints` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `tasks` | `/v4/tasks` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `timelog_categories` | `/v4/timelog_categories` | `Default`, `amReadOnlyTimelogCategory`, `amReadWriteTimelogCategory`, `wsReadOnly`, `wsReadWrite` |
| `timelogs` | `/v4/timelogs` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `timesheet_submission_rules` | `/v4/timesheet_submission_rules` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `timesheets` | `/v4/timesheets` | `Default`, `wsReadOnly`, `wsReadWrite` |
| `user_schedule_exclusions` | `/v4/user_schedule_exclusions` | `amReadOnlyWorkSchedule`, `amReadWriteWorkSchedule` |
| `user_types` | `/v4/user_types` | `amReadOnlyAccessRole` |
| `workflows` | `/v4/workflows` | `Default`, `amReadOnlyWorkflow`, `amReadWriteWorkflow`, `wsReadOnly`, `wsReadWrite` |
| `workschedules` | `/v4/workschedules` | `amReadOnlyWorkSchedule`, `amReadWriteWorkSchedule` |

Object names are the last URL segment, which is also how Wrike's reference
names each resource (`customfields`, `jobroles`, `timelogs` are spelled
that way by Wrike).

## Metadata

`ListObjectMetadata` is served from the static `metadata/schemas.json`,
generated from Wrike's official OpenAPI spec. For every object the fields are
taken verbatim from the spec's list-endpoint response schema (the item schema
of `data`).

- Spec: `scripts/openapi/wrike/internal/files/wrike_api_v4_ver154.json`,
  downloaded from <https://developers.wrike.com/openapi/> (the file is served
  with a `.yaml` extension but is JSON, OpenAPI 3.0.1). Stored with Git LFS;
  run `git lfs pull` after cloning before regenerating.
- Generator: `go run ./scripts/openapi/wrike/metadata`. Objects are an
  explicit allow-list in the script. The spec loads without any rewriting.
- Field types come from the spec: enums become `singleSelect` with their
  values (for example `tasks.importance`, `tasks.status`), numbers map to
  `other` because the shared converter only maps integer, boolean and string.
- `scripts/openapi/wrike/internal/files/live-fields.json` adds fields seen
  in live responses but missing from the spec (`approvals.reviewId`).
- Some fields are returned by the API only when requested through the
  `fields` query parameter (for example `tasks.recurrent`,
  `tasks.attachmentCount`, `contacts.metadata`). The metadata lists them
  because the spec declares them on the record; the read connector decides
  which to request.

Refresh the spec with:

```sh
curl -L -o scripts/openapi/wrike/internal/files/wrike_api_v4_ver154.json \
  https://developers.wrike.com/openapi/wrike_api_v4_ver154.yaml
go run ./scripts/openapi/wrike/metadata
```

## Read

Implemented in `read.go` with the per-object table in `objects.go`; all 27
objects are readable.

- **Records.** Every response is the `{kind, data}` envelope; records come
  from `data` (the `responseKey` in `schemas.json`) and are returned as-is.
- **Pagination.** `tasks`, `approvals`, `timelogs`, `audit_log`,
  `request_forms`, `task_blueprints` and `groups` page with `pageSize` and a
  token (`nextPageToken`; `pageToken` for groups). The next-page value is the
  full next URL, so the token travels with the original filters and `fields`
  selection, which Wrike does not keep across pages (verified live). The last
  page omits the token. Page size defaults to 100, capped at 1000 (100 for
  `request_forms`). Every other object, including `folders` (which rejects
  `pageSize`), returns the whole collection in one response.
- **Optional fields.** Wrike returns some fields only when they are named in
  its `fields` query parameter (JSON array; unknown names are rejected). The
  allowed names per object are listed in `objects.go` from the spec; the ones
  present in `ReadParams.Fields` are sent, so fields such as
  `tasks.description` or `contacts.customFields` are actually returned.
- **Record id.** `id` everywhere except `timesheets` (`timesheetId`);
  `timesheet_submission_rules` records have no identifier.

### Incremental read

Range filters are a JSON object `{"start","end"}` in `yyyy-MM-ddTHH:mm:ssZ`
(seconds only; millisecond input and empty ranges are rejected, verified
live). Since maps to `start`, Until to `end`, both optional; filtering is
provider-side.

| Object | Filter | Notes |
|---|---|---|
| `tasks`, `folders`, `approvals`, `timelogs` | `updatedDate` | |
| `audit_log` | `eventDate` | events are immutable |
| `comments` | none | `createdDate` is capped at a 7-day window and `updatedDate` is deprecated; at most 1000 comments are returned |
| `attachments` | none | `createdDate` is capped at a 31-day window |
| all others | none | no date filter in the API; full read every sync |
