package apollo

import (
	"strings"

	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var (
	restAPIPrefix string    = "v1"
	pageSize      string    = "100"
	readOp        operation = "read"
	pageQuery     string    = "page"
	writeOp       operation = "write"
	searchingPath string    = "search"
)

// readingSearchObjectGET represents objects that read by search and uses GET method.
//
//nolint:gochecknoglobals
var readingSearchObjectsGET = []string{"opportunities", "users", "deals"}

// readingSearchObjects represents objects that read by search and uses POST method.
//
//nolint:gochecknoglobals
var readingSearchObjectsPOST = []string{"accounts", "contacts", "tasks", "emailer_campaigns", "sequences"}

// readingListObjects represents objects that read by listing.
//
//nolint:gochecknoglobals,lll
var readingListObjects = []string{"contact_stages", "opportunity_stages", "account_stages", "email_accounts", "labels", "typed_custom_fields", "deal_stages", "lists_and_tags"}

// productNameToObjectName represents a mapping between the docs displaynames to object names.
//
//nolint:gochecknoglobals,lll
var productNameToObjectName = map[string]string{
	"sequences":      "emailer_campaigns",
	"deals":          "opportunities",
	"deal_stages":    "opportunity_stages",
	"lists_and_tags": "labels",
}

var usesFieldsResource = datautils.NewStringSet("contacts", "accounts", "opportunities") // nolint: gochecknoglobals

// Apollo uses mismatched API object names and display names in the documentation.
// We want to support both naming conventions. This function checks whether the provided objectName
// is a display name, and if so, maps it to the corresponding API object name.
func constructSupportedObjectName(obj string) string {
	// we want to update the objectName if the provided objectName
	// is the product name from the API docs to the supported objectName.
	// Example: sequence would be mapped to emailer_campaigns.
	// ref: https://docs.apollo.io/reference/search-for-sequences
	mappedObjectName, ok := productNameToObjectName[strings.ToLower(obj)]
	if ok {
		// Renaming the ObjectName to the mapped object.
		obj = mappedObjectName
	}

	return obj
}
