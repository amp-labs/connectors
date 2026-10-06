package webflow

import (
	"time"

	"github.com/amp-labs/connectors/common/readhelper"
)

// Object names for read routing (match schemas.json keys).
const (
	objectActivityLogs      = "activity_logs"
	objectAssetFolders      = "asset_folders"
	objectAssets            = "assets"
	objectCollections       = "collections"
	objectComments          = "comments"
	objectComponents        = "components"
	objectCustomCodeBlocks  = "custom_code_blocks"
	objectCustomDomains     = "custom_domains"
	objectCustomFonts       = "custom_fonts"
	objectFormSubmissions   = "form_submissions"
	objectForms             = "forms"
	objectGoogleTags        = "google_tags"
	objectOrders            = "orders"
	objectPages             = "pages"
	objectProducts          = "products"
	objectRedirects         = "redirects"
	objectRegisteredScripts = "registered_scripts"
	objectSites             = "sites"
	objectWebhooks          = "webhooks"
)

// Record timestamp field and format. Webflow returns ISO 8601 with
// millisecond precision (e.g. 2026-10-06T07:33:22.807Z), which time.RFC3339 parses.
const (
	fieldLastUpdated = "lastUpdated"
	fieldProduct     = "product"

	timestampFormat = time.RFC3339
)

type paginationStyle int

const (
	// paginationNone: the endpoint returns the whole collection in one response
	// and ignores limit/offset (verified live for sites, collections, webhooks).
	paginationNone paginationStyle = iota
	// paginationOffset: limit/offset query params with a
	// pagination{limit,offset,total} object in the response.
	paginationOffset
)

// readSpec is the verified per-object read behaviour.
type readSpec struct {
	pagination paginationStyle

	// incrementalKey is the record timestamp used for Since/Until. It is set
	// only when the endpoint can return records sorted by that field newest
	// first, so the connector-side filter can stop paging early. Empty means
	// Since/Until are ignored and every read is a full read.
	incrementalKey string

	// idField locates the record identifier populated into ReadResultRow.Id.
	idField readhelper.IdFieldQuery
}

// readSpecs documents, per object, pagination, incremental support and the id
// field. Sources: the Webflow OpenAPI spec and live verification on
// 2026-10-06 (see metadata/README.md).
//
// Incremental read: no Webflow list endpoint accepts an "updated since"
// parameter. Comments are the only list that can be sorted server-side, so
// they are requested by lastUpdated descending and filtered connector-side
// with early stop. Every other list comes back in no useful order, so a
// connector-side filter would still have to walk every page; those objects
// are read in full and Since/Until are ignored.
//
//nolint:gochecknoglobals
var readSpecs = map[string]readSpec{
	// Account-scoped, single response.
	objectSites: {paginationNone, "", idField("id")},

	// Site-scoped, offset paginated.
	objectActivityLogs:      {paginationOffset, "", idField("id")},
	objectAssetFolders:      {paginationOffset, "", idField("id")},
	objectAssets:            {paginationOffset, "", idField("id")},
	objectComments:          {paginationOffset, fieldLastUpdated, idField("id")},
	objectComponents:        {paginationOffset, "", idField("id")},
	objectCustomFonts:       {paginationOffset, "", idField("id")},
	objectFormSubmissions:   {paginationOffset, "", idField("id")},
	objectForms:             {paginationOffset, "", idField("id")},
	objectOrders:            {paginationOffset, "", idField("orderId")},
	objectPages:             {paginationOffset, "", idField("id")},
	objectRegisteredScripts: {paginationOffset, "", idField("id")},
	// Custom code blocks carry no id field; ReadResultRow.Id stays empty.
	objectCustomCodeBlocks: {paginationOffset, "", idField("id")},
	// Products are {product, skus} envelopes; the id lives under product.
	objectProducts: {paginationOffset, "", readhelper.NewNestedIdField([]string{fieldProduct}, "id")},

	// Site-scoped, single response (limit/offset ignored).
	objectCollections:   {paginationNone, "", idField("id")},
	objectCustomDomains: {paginationNone, "", idField("id")},
	objectGoogleTags:    {paginationNone, "", idField("tagId")},
	objectRedirects:     {paginationNone, "", idField("id")},
	objectWebhooks:      {paginationNone, "", idField("id")},
}

func idField(name string) readhelper.IdFieldQuery {
	return readhelper.NewIdField(name)
}
