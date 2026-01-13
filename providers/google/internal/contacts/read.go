package contacts

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	// Page size reference, which applies to all objects:
	// https://developers.google.com/people/api/rest/v1/contactGroups/list
	defaultPageSize = 1000
)

func (a *Adapter) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(params)
	if err != nil {
		return nil, err
	}

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}

func (a *Adapter) buildReadURL(params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		return urlbuilder.New(params.NextPage.String())
	}

	// First page
	url, err := a.getReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("pageSize", strconv.Itoa(defaultPageSize))

	switch params.ObjectName {
	case objectNameContactGroups:
		// https://developers.google.com/people/api/rest/v1/contactGroups/list
		readURLForContactGroups(params, url)
	case objectNameMyConnections:
		// https://developers.google.com/people/api/rest/v1/people.connections/list
		readURLForMyConnections(params, url)
	case objectNameOtherContacts:
		// https://developers.google.com/people/api/rest/v1/otherContacts/list
		readURLForOtherContacts(params, url)
	case objectNamePeopleDirectory:
		// https://developers.google.com/people/api/rest/v1/people/listDirectoryPeople
		readURLForPeopleDirectory(params, url)
	}

	return url, nil
}

func (a *Adapter) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	resp *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	responseFieldName := Schemas.LookupArrayFieldName(a.Module(), params.ObjectName)

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(resp,
		common.ExtractOptionalRecordsFromPath(responseFieldName),
		makeNextRecordsURL(url),
		getMarshaledData,
		params.Fields,
	)
}

func getMarshaledData(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	withID := datautils.NewSet(fields...).Has("id")

	for index, record := range records {
		fieldsResult := common.ExtractLowercaseFieldsFromRaw(fields, record)

		if withID {
			if recordID, found := extractID(record); found {
				fieldsResult["id"] = recordID
			}
		}

		data[index] = common.ReadResultRow{
			Fields: fieldsResult,
			Raw:    record,
		}
	}

	return data, nil
}

func extractID(record map[string]any) (recordID string, found bool) {
	resourceNameField, found := record["resourceName"]
	if !found {
		return "", false
	}

	resourceName, convertible := resourceNameField.(string)
	if !convertible {
		return "", false
	}

	return resourceIdentifierFormat(resourceName)
}

func makeNextRecordsURL(url *urlbuilder.URL) common.NextPageFunc {
	// Alter current request URL to progress with the next page token.
	return func(node *ajson.Node) (string, error) {
		pageToken, err := jsonquery.New(node).StrWithDefault("nextPageToken", "")
		if err != nil {
			return "", err
		}

		if len(pageToken) == 0 {
			// Next page doesn't exist
			return "", nil
		}

		url.AddEncodingExceptions(map[string]string{
			"%3D": "=",
		})
		url.WithQueryParam("pageToken", pageToken)

		return url.String(), nil
	}
}

// readURLForContactGroups constructs the query parameters for the request.
// - Filters out fields that are not permitted in groupFields.
// https://developers.google.com/people/api/rest/v1/contactGroups/list
func readURLForContactGroups(params common.ReadParams, url *urlbuilder.URL) {
	groupFields := datautils.NewSetFromList([]string{
		"clientData", "groupType", "memberCount", "metadata", "name",
	}).Intersection(params.Fields)

	if len(groupFields) != 0 {
		url.WithQueryParam("groupFields", strings.Join(groupFields, ","))
	}
}

// readURLForMyConnections constructs the query parameters for the request.
// - Filters out fields that are not permitted in personFields.
// https://developers.google.com/people/api/rest/v1/people.connections/list
func readURLForMyConnections(params common.ReadParams, url *urlbuilder.URL) {
	personFields := datautils.NewStringSet(
		"addresses", "ageRanges", "biographies", "birthdays", "calendarUrls", "clientData", "coverPhotos",
		"emailAddresses", "events", "externalIds", "genders", "imClients", "interests", "locales", "locations",
		"memberships", "metadata", "miscKeywords", "names", "nicknames", "occupations", "organizations",
		"phoneNumbers", "photos", "relations", "sipAddresses", "skills", "urls", "userDefined",
	).Intersection(params.Fields)

	if len(personFields) != 0 {
		url.WithQueryParam("personFields", strings.Join(personFields, ","))
	}
}

// readURLForContactGroups constructs the query parameters for the request.
// - Filters out fields that are not permitted in readMask.
// - Ensures the sources query parameter mirrors the selected group of fields, acting as a mode indicator.
func readURLForOtherContacts(params common.ReadParams, url *urlbuilder.URL) {
	contactsFields := readMaskForOtherContacts[otherContactsSourceContact].Intersection(params.Fields)
	profileFields := readMaskForOtherContacts[otherContactsSourceProfile].Intersection(params.Fields)

	// Deduce which sources to specify based on the field provided.
	// https://developers.google.com/people/api/rest/v1/otherContacts/list
	sources := make([]string, 0)
	if len(contactsFields) != 0 {
		sources = []string{otherContactsSourceContact}
	}

	// Profile source must be provided alongside contact source.
	if len(profileFields) != 0 {
		sources = []string{otherContactsSourceContact, otherContactsSourceProfile}
	}

	if len(sources) != 0 {
		url.WithQueryParamList("sources", sources)
	}

	// Combine Profile and Contacts fields.
	readMaskFields := datautils.NewStringSet()
	readMaskFields.Add(contactsFields)
	readMaskFields.Add(profileFields)

	if len(readMaskFields) != 0 {
		url.WithQueryParam("readMask", strings.Join(readMaskFields.List(), ","))
	}
}

// https://developers.google.com/people/api/rest/v1/people/listDirectoryPeople
func readURLForPeopleDirectory(config common.ReadParams, url *urlbuilder.URL) {
	// Sources is a required field. Request all possible sources.
	// https://developers.google.com/people/api/rest/v1/DirectorySourceType
	url.WithQueryParamList("sources", []string{
		"DIRECTORY_SOURCE_TYPE_DOMAIN_CONTACT",
		"DIRECTORY_SOURCE_TYPE_DOMAIN_PROFILE",
	})

	readMaskFields := datautils.NewStringSet(
		"addresses", "ageRanges", "biographies", "birthdays", "calendarUrls", "clientData", "coverPhotos",
		"emailAddresses", "events", "externalIds", "genders", "imClients", "interests", "locales", "locations",
		"memberships", "metadata", "miscKeywords", "names", "nicknames", "occupations", "organizations",
		"phoneNumbers", "photos", "relations", "sipAddresses", "skills", "urls", "userDefined",
	).Intersection(config.Fields)

	if len(readMaskFields) != 0 {
		url.WithQueryParam("readMask", strings.Join(readMaskFields, ","))
	}
}

const (
	otherContactsSourceContact = "READ_SOURCE_TYPE_CONTACT"
	otherContactsSourceProfile = "READ_SOURCE_TYPE_PROFILE"
)

// https://developers.google.com/people/api/rest/v1/otherContacts/list
var readMaskForOtherContacts = map[string]datautils.StringSet{ // nolint:gochecknoglobals
	otherContactsSourceContact: datautils.NewSet(
		"emailAddresses", "metadata", "names", "phoneNumbers", "photos",
	),
	otherContactsSourceProfile: datautils.NewSet(
		"addresses", "ageRanges", "biographies", "birthdays", "calendarUrls", "clientData", "coverPhotos",
		"emailAddresses", "events", "externalIds", "genders", "imClients", "interests", "locales", "locations",
		"memberships", "metadata", "miscKeywords", "names", "nicknames", "occupations", "organizations",
		"phoneNumbers", "photos", "relations", "sipAddresses", "skills", "urls", "userDefined",
	),
}
