package calendly

import (
	"context"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common/naming"
)

// NormalizeEntityName normalizes entity names according to Calendly naming conventions.
// Calendly uses snake_case (lowercase with underscores) for all entity names in its API.
//
// Objects:
//   - Converts to lowercase plural with underscores (snake_case)
//   - Examples: "EventType" -> "event_types", "ScheduledEvent" -> "scheduled_events"
//
// Fields:
//   - Converts to lowercase with underscores (snake_case)
//   - Examples: "CreatedAt" -> "created_at", "StartTime" -> "start_time"
//
// Calendly API Reference: https://developer.calendly.com/api-docs/d7755e2f9e5fe-calendly-api
// All standard objects use plural snake_case: activity_log_entries, event_types, scheduled_events, etc.
// All fields use snake_case: uri, name, created_at, start_time, etc.
func (c *Connector) NormalizeEntityName(
	ctx context.Context, entity connectors.Entity, input string,
) (normalized string, err error) {
	switch entity {
	case connectors.EntityObject:
		return normalizeObjectName(input), nil
	case connectors.EntityField:
		return normalizeFieldName(input), nil
	default:
		// Unknown entity type, return unchanged
		return input, nil
	}
}

// normalizeObjectName converts object names to lowercase plural snake_case.
// Calendly's standard objects are always plural and use snake_case:
// - event_types (not eventTypes or EventTypes)
// - scheduled_events (not scheduledEvents or ScheduledEvents)
// - activity_log_entries, user_busy_times, webhook_subscriptions, etc.
func normalizeObjectName(input string) string {
	// Convert to plural form
	plural := naming.NewPluralString(input).String()

	// Convert to snake_case (lowercase with underscores)
	return naming.ToSnakeCase(plural)
}

// normalizeFieldName converts field names to lowercase snake_case.
// Calendly field names consistently use snake_case:
// - created_at, updated_at, start_time, end_time.
// - booking_method, scheduling_url, fully_qualified_name.
func normalizeFieldName(input string) string {
	// Convert to snake_case (lowercase with underscores)
	return naming.ToSnakeCase(input)
}
