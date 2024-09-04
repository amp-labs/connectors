package apollo

type ObjectType string

var searchingPath string = "search" //nolint:gochecknoglobals

// postSearchObjects represents the objects that uses the searching endpoint, POST method
// for requesting records.
var postSearchObjects = []ObjectType{ //nolint:gochecknoglobals
	"mixed_people", "mixed_companies", "contacts",
	"accounts", "emails_campaigns",
}

// getSearchObjects represents the objects that uses the searching endpoint, GET method
// for requesting records.
// Tasks has a query parameter `open_factor_names` requirement.
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

// var sortKValues = map[string]string{
// 	"mixed_people":      "",
// 	"contacts":          "contact_last_activity_date",
// 	"opportunities":     "opportunity_created_at",
// 	"accounts":          "account_last_activity_date",
// 	"mixed_companies":   "",
// 	"emailer_campaigns": "",
// 	"users":             "",
// }
