# Calendly Connector

Integrates with Calendly API v2 for reading scheduled events and managing webhook subscriptions.

## Supported
- **Read:** `scheduled_events`
- **Subscribe:** Webhook events: `invitee.created`, `invitee.canceled`, `invitee_no_show.created`, `invitee_no_show.deleted`, `routing_form_submission.created`
- **Auth:** OAuth 2.0 required

## Usage
```go
connector, _ := calendly.NewConnector(common.ConnectorParams{AuthenticatedClient: httpClient})
result, _ := connector.Read(ctx, common.ReadParams{ObjectName: "scheduled_events", Fields: connectors.Fields("name")})
```

## Webhooks
- Create, update, delete subscriptions for supported events

## Testing
```bash
go test ./providers/calendly/...
``` 