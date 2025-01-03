package apollo

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
var readingSearchObjectsGET = []string{"opportunities", "users"}

// readingSearchObjects represents objects that read by search and uses POST method.
//
//nolint:gochecknoglobals
var readingSearchObjectsPOST = []string{"accounts", "contacts", "tasks", "emailer_campaigns"}

// readingListObjects represents objects that read by listing.
//
//nolint:gochecknoglobals,lll
var readingListObjects = []string{"contact_stages", "opportunity_stages", "account_stages", "email_accounts", "labels", "typed_custom_fields"}

// displayNameToObjectName represents a mapping between the docs displaynames to object names.
//
//nolint:gochecknoglobals,lll
var displayNameToObjectName = map[string]string{
	"sequences":      "emailer_campaigns",
	"deals":          "opportunities",
	"deal_stages":    "opportunity_stages",
	"lists_and_tags": "labels",
}
