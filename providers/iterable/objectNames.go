package iterable

import (
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/iterable/metadata"
)

const (
	objectNameCatalogs  = "catalogs"
	objectNameJourneys  = "journeys"
	objectNameTemplates = "templates"
)

var paginatedObjects = datautils.NewSet( //nolint:gochecknoglobals
	objectNameCatalogs,
	objectNameJourneys,
)

// While reading is done against one object, the writing occurs by template type.
var templateWriteObjects = datautils.NewSet( //nolint:gochecknoglobals
	"templatesEmail",
	"templatesInApp",
	"templatesPush",
	"templatesSMS",
)

// Supported object names can be found under schemas.json.
var supportedObjectsByRead = metadata.Schemas.ObjectNames() //nolint:gochecknoglobals

var supportedObjectsByWrite = datautils.Map[string, string]{ //nolint:gochecknoglobals
	// https://api.iterable.com/api/docs#campaigns_create_campaign
	"campaigns": "/campaigns/create",
	// https://api.iterable.com/api/docs#catalogs_createCatalog
	// This endpoint doesn't use payload! Catalog name comes from the path {catalogName}.
	objectNameCatalogs: "/catalogs",
	// https://api.iterable.com/api/docs#lists_create
	"lists": "/lists",
	// https://api.iterable.com/api/docs#users_updateUser
	"users": "/users/update", // Update or create user.
	// https://api.iterable.com/api/docs#webhooks_updateWebhook
	"webhooks": "/webhooks", // Update webhook

	//
	// Template objects.
	//
	// https://api.iterable.com/api/docs#templates_upsertEmailTemplate
	"templatesEmail": "/templates/email/upsert",
	// https://api.iterable.com/api/docs#templates_upsertInAppTemplate
	"templatesInApp": "/templates/inapp/upsert",
	// https://api.iterable.com/api/docs#templates_upsertPushTemplate
	"templatesPush": "/templates/push/upsert",
	// https://api.iterable.com/api/docs#templates_upsertSMSTemplate
	"templatesSMS": "/templates/sms/upsert",
}

var supportedObjectsByDelete = datautils.Map[string, string]{ //nolint:gochecknoglobals
	// https://api.iterable.com/api/docs#catalogs_deleteCatalog
	objectNameCatalogs: "/catalogs", // by catalogName
	// https://api.iterable.com/api/docs#export_cancelExport
	"exports": "/export", // by jobId
	// https://api.iterable.com/api/docs#lists_delete
	"lists": "/lists", // by listId
	// https://api.iterable.com/api/docs#metadata_delete
	"metadata": "/metadata", // by table
	// https://api.iterable.com/api/docs#users_delete
	"users": "/users/byUserId", // by userId
}
