package apollo

var (
	restAPIPrefix string    = "v1"     //nolint:gochecknoglobals
	pageQuery     string    = "page"   //nolint:gochecknoglobals
	pageSize      string    = "100"    //nolint:gochecknoglobals
	searchingPath string    = "search" //nolint:gochecknoglobals
	readOp        operation = "read"   //nolint:gochecknoglobals
	writeOp       operation = "write"  //nolint:gochecknoglobals
)

type ObjectType string

// postSearchObjects represents the objects that uses the searching endpoint,
// POST method for requesting records.
var postSearchObjects = []ObjectType{ //nolint:gochecknoglobals
	"mixed_people", "mixed_companies", "contacts",
	"accounts", "emails_campaigns",
}

// getSearchObjects represents the objects that uses the searching endpoint, GET method
// for requesting records.Tasks has a query parameter `open_factor_names` requirement.
var getSearchObjects = []ObjectType{"opportunities", "users"} //nolint:gochecknoglobals

// responseKey represent the results key fields in the response.
// some endpoints have more than one, data fields returned.
var responseKey = map[string][]string{ //nolint:gochecknoglobals
	"mixed_people":      {"people", "accounts"},
	"mixed_companies":   {"organizations", "accounts"},
	"opportunities":     {"opportunities"},
	"accounts":          {"accounts"},
	"emailer_campaigns": {"emailer_campaigns"},
	"users":             {"users"},
	"contacts":          {"contacts"},
}
