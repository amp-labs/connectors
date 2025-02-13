package marketo

import (
	"errors"
	"fmt"
	"slices"
)

const batchSize = 300 // nolint:gochecknoglobals

type writeResponse struct {
	Result  []map[string]any `json:"result"`
	Success bool             `json:"success"`
	Errors  []map[string]any `json:"errors"`
}

// nolint:gochecknoglobals
var (
	// IdResponseObjects represents a list of objects that uses `id` as a unique field in the response.
	IdResponseObjects = []string{"leads", "companies", "salespersons"}

	// marketoGUIDResponseObjects represents a list of objects that uses `marketoGUID` as the unique field in the response.
	marketoGUIDResponseObjects = []string{"namedAccountLists", "namedaccounts", "opportunities"}

	// idFilteringObjects represents objects that uses id as filtering values in read connector.
	idFilteringObjects = []string{"leads", "salespersons", "companies"}

	// metadataPaths represents a map of a few objects in Marketo that has unique resource for returning metadata fields.
	metadataPaths = map[string]string{
		"leads":         "rest/v1/leads/describe2.json",
		"companies":     "rest/v1/companies/describe.json",
		"namedaccounts": "rest/v1/namedaccounts/describe.json",
		"salespersons":  "rest/v1/salespersons/describe.json",
		"opportunities": "rest/v1/opportunities/describe.json",
	}

	ErrFailedConvertFields = errors.New("failed to convert the response message to metadata fields")
)

func constructErrMessage(a any) (string, error) {
	s, ok := a.([]map[string]any)
	if !ok {
		return "", errors.New("failed to convert the response message to an error type") // nolint:goerr113
	}

	return fmt.Sprint(s[0]["reasons"]), nil
}

func paginatesByIDs(object string) bool {
	// Most Marketo APIs requires filtering when reading, Important objects are Leads, Custom Objects, Companies
	// With this we use the general filter parameter `id` and iterate over the records.
	return slices.Contains(idFilteringObjects, object)
}

func hasMetadataResource(object string) (string, bool) {
	path, ok := metadataPaths[object]
	if !ok {
		return "", false
	}

	return path, true
}
