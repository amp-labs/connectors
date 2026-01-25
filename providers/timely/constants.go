package timely

import (
	"github.com/amp-labs/connectors/internal/datautils"
)

// API version and base URL
const (
	apiVersion = "1.1"
	baseURL    = "https://api.timelyapp.com/1.1"
)

// OAuth2 endpoints
const (
	oauthAuthorizePath = "oauth/authorize"
	oauthTokenPath     = "oauth/token"
)

// Object to endpoint mapping for CRUD-style resources
var objectResourcePath = map[string]string{
	"accounts":                 "accounts",
	"account":                  "accounts/%d",
	"accountActivities":        "%d/activities",
	"clients":                  "%d/clients",
	"client":                   "%d/clients/%d",
	"day-locking":              "%d/day_properties",
	"day-locking-entry":        "%d/day_properties/%d",
	"events":                   "%d/events",
	"event":                    "%d/events/%d",
	"bulk-events":              "%d/bulk/events",
	"events-by-project":        "%d/projects/%d/events",
	"events-by-user":           "%d/users/%d/events",
	"event-for-project":        "%d/projects/%d/events",
	"event-for-other-user":     "%d/users/%d/events",
	"event-by-logged-in-user":  "%d/users/%d/events/%d",
	"event-by-project-by-user": "%d/projects/%d/events/%d",
	"start-timer-on-event":     "%d/events/%d/start",
	"stop-timer-on-event":      "%d/events/%d/stop",
	"projects":                 "%d/projects",
	"project":                  "%d/projects/%d",
	"users":                    "%d/users",
	"user":                     "%d/users/%d",
	"user-invite":              "%d/users/invite",
	"user-current":             "%d/users/current",
	"user-capacities":          "%d/users/capacities",
	"user-capacity":            "%d/users/%d/capacities",
	"user-permissions":         "%d/users/%d/permissions",
	"current-user-permissions": "%d/users/current/permissions",
	"labels":                   "%d/labels",
	"label":                    "%d/labels/%d",
	"teams":                    "%d/teams",
	"team":                     "%d/teams/%d",
	"webhooks":                 "%d/webhooks",
	"webhook":                  "%d/webhooks/%d",
	"forecasts":                "%d/forecasts",
	"forecast":                 "%d/forecasts/%d",
	"reports":                  "%d/reports",
	"filter-reports":           "%d/reports/filter",
	"roles":                    "%d/roles",
}

// Response fields
const (
	responseFieldData = "data"
	responseFieldMeta = "meta"
)

// Error messages
const (
	errInvalidAuth    = "Invalid authentication credentials"
	errRateLimit      = "Rate limit exceeded"
	errInvalidRequest = "Invalid request parameters"
	errNotFound       = "Resource not found"
	errServerError    = "Internal server error"
	errInvalidJSON    = "Invalid JSON payload"
)

// Pagination
const (
	defaultPageSize = 50
	maxPageSize     = 100
)

// Query parameters
const (
	queryParamLimit  = "limit"
	queryParamOffset = "offset"
	queryParamSort   = "sort"
	queryParamOrder  = "order"
)

// Supported objects for read operations
var readSupportedObjects = datautils.NewSet(
	"events",
	"projects",
	"clients",
	"users",
	"labels",
	"teams",
	"webhooks",
	"day-locking",
	"forecasts",
)

// Supported objects for write operations
var writeSupportedObjects = datautils.NewSet(
	"events",
	"projects",
	"clients",
	"labels",
	"teams",
	"webhooks",
	"day-locking",
	"forecasts",
)

// HTTP Status codes (see https://dev.timelyapp.com/#errors for meaning)
const (
	statusBadRequest          = 400
	statusUnauthorized        = 401
	statusForbidden           = 403
	statusNotFound            = 404
	statusUnprocessableEntity = 422
	statusInternalServerError = 500
	statusRateLimit           = 429
)

// Webhook event types
const (
	webhookEventCreated = "created"
	webhookEventUpdated = "updated"
	webhookEventDeleted = "deleted"
)

// Resource types for webhook subscriptions
const (
	resourceTypeEvent    = "events"
	resourceTypeProject  = "projects"
	resourceTypeClient   = "clients"
	resourceTypeUser     = "users"
	resourceTypeLabel    = "labels"
	resourceTypeTeam     = "teams"
	resourceTypeForecast = "forecasts"
)
