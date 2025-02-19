package google

import (
	"context"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/google/metadata"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	if !supportedObjectsByRead[c.Module.ID].Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, config.ObjectName)

	return common.ParseResult(res,
		common.GetOptionalRecordsUnderJSONPath(responseFieldName),
		makeNextRecordsURL(url),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	// First page
	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if c.Module.ID == ModuleCalendar {
		url.WithQueryParam("maxResults", strconv.Itoa(CalendarDefaultPageSize))

		if config.ObjectName == objectNameCalendarList {
			// This is the only object to support search by deleted items.
			// https://developers.google.com/calendar/api/v3/reference/calendarList/list
			if config.Deleted {
				url.WithQueryParam("showDeleted", "true")
			}
		}
	}

	if c.Module.ID == ModulePeople {
		url.WithQueryParam("pageSize", strconv.Itoa(PeopleDefaultPageSize))

		switch config.ObjectName {
		case objectNameMyConnections:
			readURLForMyConnections(config, url)
		case objectNameContactGroups:
			readURLForContactGroups(config, url)
		case objectNameOtherContacts:
			readURLForOtherContacts(config, url)
		case objectNamePeopleDirectory:
			readURLForPeopleDirectory(config, url)
		}
	}

	return url, nil
}

// readURLForMyConnections constructs the query parameters for the request.
// - Filters out fields that are not permitted in personFields.
// https://developers.google.com/people/api/rest/v1/people.connections/list
func readURLForMyConnections(config common.ReadParams, url *urlbuilder.URL) {
	personFields := make([]string, 0)

	for fieldName := range queryFieldsForMyConnections {
		if config.Fields.Has(fieldName) {
			personFields = append(personFields, fieldName)
		}
	}

	if len(personFields) != 0 {
		url.WithQueryParam("personFields", strings.Join(personFields, ","))
		url.AddEncodingExceptions(map[string]string{
			"%2C": ",",
		})
	}
}

// readURLForContactGroups constructs the query parameters for the request.
// - Filters out fields that are not permitted in groupFields.
// https://developers.google.com/people/api/rest/v1/contactGroups/list
func readURLForContactGroups(config common.ReadParams, url *urlbuilder.URL) {
	groupFields := make([]string, 0)

	for fieldName := range queryFieldsForContactGroups {
		if config.Fields.Has(fieldName) {
			groupFields = append(groupFields, fieldName)
		}
	}

	if len(groupFields) != 0 {
		url.WithQueryParam("groupFields", strings.Join(groupFields, ","))
		url.AddEncodingExceptions(map[string]string{
			"%2C": ",",
		})
	}
}

// readURLForContactGroups constructs the query parameters for the request.
// - Filters out fields that are not permitted in readMask.
// - Ensures the sources query parameter mirrors the selected group of fields, acting as a mode indicator.
func readURLForOtherContacts(config common.ReadParams, url *urlbuilder.URL) {
	var (
		readMaskFields    = datautils.NewStringSet()
		sourceTypeProfile = false
	)

	for field := range config.Fields {
		if readMaskForOtherContacts[otherContactsSourceContact].Has(field) {
			readMaskFields.AddOne(field)
		}

		if readMaskForOtherContacts[otherContactsSourceProfile].Has(field) {
			sourceTypeProfile = true

			readMaskFields.AddOne(field)
		}
	}

	if sourceTypeProfile {
		// Contact must be specified. Not including it is not permitted
		// https://developers.google.com/people/api/rest/v1/otherContacts/list
		url.WithQueryParamList("sources", []string{
			otherContactsSourceContact,
			otherContactsSourceProfile,
		})
	} else {
		url.WithQueryParam("sources", otherContactsSourceContact)
	}

	if len(readMaskFields) != 0 {
		url.WithQueryParam("readMask", strings.Join(readMaskFields.List(), ","))
		url.AddEncodingExceptions(map[string]string{
			"%2C": ",",
		})
	}
}

// https://developers.google.com/people/api/rest/v1/people/listDirectoryPeople
func readURLForPeopleDirectory(config common.ReadParams, url *urlbuilder.URL) {
	// Sources is a required field. Request all possible sources.
	url.WithQueryParamList("sources", []string{
		"DIRECTORY_SOURCE_TYPE_DOMAIN_CONTACT",
		"DIRECTORY_SOURCE_TYPE_DOMAIN_PROFILE",
	})

	readMaskFields := make([]string, 0)

	for fieldName := range readMaskForPeopleDirectory {
		if config.Fields.Has(fieldName) {
			readMaskFields = append(readMaskFields, fieldName)
		}
	}

	if len(readMaskFields) != 0 {
		url.WithQueryParam("readMask", strings.Join(readMaskFields, ","))
		url.AddEncodingExceptions(map[string]string{
			"%2C": ",",
		})
	}
}

const (
	otherContactsSourceContact = "READ_SOURCE_TYPE_CONTACT"
	otherContactsSourceProfile = "READ_SOURCE_TYPE_PROFILE"
)

var queryFieldsForMyConnections = datautils.NewSet( // nolint: gochecknoglobals
	"addresses",
	"ageRanges",
	"biographies",
	"birthdays",
	"calendarUrls",
	"clientData",
	"coverPhotos",
	"emailAddresses",
	"events",
	"externalIds",
	"genders",
	"imClients",
	"interests",
	"locales",
	"locations",
	"memberships",
	"metadata",
	"miscKeywords",
	"names",
	"nicknames",
	"occupations",
	"organizations",
	"phoneNumbers",
	"photos",
	"relations",
	"sipAddresses",
	"skills",
	"urls",
	"userDefined",
)

var queryFieldsForContactGroups = datautils.NewSet( // nolint: gochecknoglobals
	"clientData",
	"groupType",
	"memberCount",
	"metadata",
	"name",
)

// https://developers.google.com/people/api/rest/v1/otherContacts/list
var readMaskForOtherContacts = map[string]datautils.StringSet{ // nolint:gochecknoglobals
	otherContactsSourceContact: datautils.NewSet(
		"emailAddresses",
		"metadata",
		"names",
		"phoneNumbers",
		"photos",
	),
	otherContactsSourceProfile: datautils.NewSet(
		"addresses",
		"ageRanges",
		"biographies",
		"birthdays",
		"calendarUrls",
		"clientData",
		"coverPhotos",
		"emailAddresses",
		"events",
		"externalIds",
		"genders",
		"imClients",
		"interests",
		"locales",
		"locations",
		"memberships",
		"metadata",
		"miscKeywords",
		"names",
		"nicknames",
		"occupations",
		"organizations",
		"phoneNumbers",
		"photos",
		"relations",
		"sipAddresses",
		"skills",
		"urls",
		"userDefined",
	),
}

var readMaskForPeopleDirectory = datautils.NewSet( // nolint:gochecknoglobals
	"addresses",
	"ageRanges",
	"biographies",
	"birthdays",
	"calendarUrls",
	"clientData",
	"coverPhotos",
	"emailAddresses",
	"events",
	"externalIds",
	"genders",
	"imClients",
	"interests",
	"locales",
	"locations",
	"memberships",
	"metadata",
	"miscKeywords",
	"names",
	"nicknames",
	"occupations",
	"organizations",
	"phoneNumbers",
	"photos",
	"relations",
	"sipAddresses",
	"skills",
	"urls",
	"userDefined",
)
