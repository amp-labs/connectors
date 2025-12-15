package ringcentral

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

type ObjectsOperationURLs struct {
	ReadPath             string
	WritePath            string
	RecordsField         string
	usesCursorPagination bool
	usesOffsetPagination bool
	usesSyncToken        bool
}

var pathURLs = map[string]ObjectsOperationURLs{ // nolint: gochecknoglobals
	"caller-blocking/phone-numbers": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/caller-blocking/phone-numbers",
		RecordsField: records,
	},
	"forwarding-number": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/forwarding-number",
		RecordsField: records,
	},
	"answering-rule": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/answering-rule",
		RecordsField: records,
	},
	"company-answering-rule": { // needs updating name
		ReadPath:     "restapi/v1.0/account/~/answering-rule",
		RecordsField: records,
	},
	"comm-handling/states": {
		ReadPath:             "restapi/v2/accounts/~/extensions/~/comm-handling/states",
		RecordsField:         records,
		usesOffsetPagination: true,
	},
	"comm-handling/voice/state-rules": {
		ReadPath:             "restapi/v2/accounts/~/extensions/~/comm-handling/voice/state-rules",
		RecordsField:         records,
		usesOffsetPagination: true,
	},
	"comm-handling/voice/interaction-rules": {
		ReadPath:             "restapi/v2/accounts/~/extensions/~/comm-handling/voice/interaction-rules",
		RecordsField:         records,
		usesOffsetPagination: true,
	},
	"comm-handling/voice/forwarding-targets": {
		ReadPath:             "restapi/v2/accounts/~/extensions/~/comm-handling/voice/forwarding-targets",
		RecordsField:         "referencedExtensions",
		usesOffsetPagination: true,
	},
	"call-flip-numbers": {
		ReadPath:     "restapi/v2/accounts/~/extensions/~/call-flip-numbers",
		RecordsField: records,
	},
	"call-log": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/call-log",
		RecordsField: records,
	},
	"active-calls": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/active-calls",
		RecordsField: records,
	},
	"company-call-log": {
		ReadPath:     "restapi/v1.0/account/~/call-log",
		RecordsField: records,
	},
	"company-active-calls": {
		ReadPath:     "restapi/v1.0/account/~/active-calls",
		RecordsField: records,
	},
	"call-log-sync": {
		ReadPath:      "restapi/v1.0/account/~/extension/~/call-log-sync",
		RecordsField:  records,
		usesSyncToken: true,
	},
	"company-call-log-sync": {
		ReadPath:      "estapi/v1.0/account/~/call-log-sync",
		RecordsField:  records,
		usesSyncToken: true,
	},
	"call-log-extract-sync": {
		ReadPath:      "restapi/v1.0/account/~/call-log-extract-sync",
		RecordsField:  records,
		usesSyncToken: true,
	},
	"call-monitoring-groups": {
		ReadPath:     "restapi/v1.0/account/~/call-monitoring-groups",
		RecordsField: records,
	},
	"call-queues": {
		ReadPath:     "restapi/v1.0/account/~/call-queues",
		RecordsField: records,
	},
	"custom-greetings": {
		ReadPath:     "restapi/v1.0/account/~/call-recording/custom-greetings",
		RecordsField: records,
	},
	"dictionary/greeting": {
		ReadPath:     "restapi/v1.0/dictionary/greeting",
		RecordsField: records,
	},
	"vr-prompts": {
		ReadPath:     "restapi/v1.0/account/~/ivr-prompts",
		RecordsField: records,
	},
	"ivr-menus": {
		ReadPath:     "restapi/v1.0/account/~/ivr-menus",
		RecordsField: records,
	},
	"fax-cover-page": {
		ReadPath:     "restapi/v1.0/dictionary/fax-cover-page",
		RecordsField: records,
	},
	"message-store": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/message-store",
		RecordsField: records,
	},
	"message-sync": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/message-sync",
		RecordsField: records,
	},

	"a2p-sms/batches": {
		ReadPath:             "restapi/v1.0/account/~/a2p-sms/batches",
		RecordsField:         records,
		usesOffsetPagination: true,
	},
	"a2p-sms/messages": {
		ReadPath:             "restapi/v1.0/account/~/a2p-sms/messages",
		RecordsField:         records,
		usesOffsetPagination: true,
	},
	"message-store-templates": {
		ReadPath:     "restapi/v1.0/account/~/message-store-templates",
		RecordsField: records,
	},
	"user-message-store-templates": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/message-store-templates",
		RecordsField: records,
	},
	"sms/consents": {
		ReadPath:     "restapi/v2/accounts/~/sms/consents",
		RecordsField: records,
	},
	"sms-registration-brands": {
		ReadPath:     "restapi/v1.0/account/~/sms-registration-brands",
		RecordsField: records,
	},
	"events": {
		ReadPath:             "team-messaging/v1/events",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"chats": {
		ReadPath:             "team-messaging/v1/chats",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"recent/chats": {
		ReadPath:     "team-messaging/v1/recent/chats",
		RecordsField: records,
	},
	"favorites": {
		ReadPath:     "team-messaging/v1/favorites",
		RecordsField: records,
	},
	"conversations": {
		ReadPath:     "team-messaging/v1/conversations",
		RecordsField: records,
	},
	"data-export-tasks": {
		ReadPath:     "team-messaging/v1/data-export",
		RecordsField: "tasks",
	},
	"webhooks": {
		ReadPath:     "team-messaging/v1/webhooks",
		RecordsField: records,
	},
	"teams": {
		ReadPath:             "team-messaging/v1/teams",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"delegators": {
		ReadPath:     "rcvideo/v1/accounts/~/extensions/~/delegators",
		RecordsField: "items",
	},

	"meetings": {
		ReadPath:             "rcvideo/v1/history/account/~/meetings",
		RecordsField:         "meetings",
		usesCursorPagination: true, // needs double checking
	},

	"history/meetings": {
		ReadPath:             "rcvideo/v1/history/meetings",
		RecordsField:         "meetings",
		usesCursorPagination: true, // needs double checking
	},
	"recordings": {
		ReadPath:             "rcvideo/v1/account/~/recordings",
		RecordsField:         "recordings",
		usesCursorPagination: true,
	},
	"extension-recordings": {
		ReadPath:             "rcvideo/v1/account/~/extension/~/recordings",
		RecordsField:         "recordings",
		usesCursorPagination: true,
	},
	"webinars": {
		ReadPath:             "webinar/configuration/v1/webinars",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"configuration/sessions": {
		ReadPath:             "webinar/configuration/v1/sessions",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"company/sessions": {
		ReadPath:             "webinar/configuration/v1/company/sessions",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"history/sessions": {
		ReadPath:             "webinar/history/v1/sessions",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"history/company/sessions": {
		ReadPath:             "webinar/history/v1/company/sessions",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"webinar/recordings": {
		ReadPath:             "webinar/history/v1/recordings",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"webinar/company/recordings": {
		ReadPath:             "webinar/history/v1/company/recordings",
		RecordsField:         records,
		usesCursorPagination: true,
	},
	"webinar/subscriptions": {
		ReadPath:     "webinar/notifications/v1/subscriptions",
		RecordsField: records,
	},
	"custom-fields": {
		ReadPath:     "restapi/v1.0/account/~/custom-fields",
		RecordsField: records,
	},
	"sites": {
		ReadPath:     "restapi/v1.0/account/~/sites",
		RecordsField: records,
	},
	"accounts/phone-numbers": {
		ReadPath:     "restapi/v2/accounts/~/phone-numbers",
		RecordsField: records,
	},
	"extension/phone-number": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/phone-number",
		RecordsField: records,
	},
	"company/phone-number": {
		ReadPath:     "restapi/v1.0/account/~/phone-number",
		RecordsField: records,
	},
	"account/presence": {
		ReadPath:     "restapi/v1.0/account/~/presence",
		RecordsField: records,
	},
	"call-queue-presence": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/call-queue-presence",
		RecordsField: records,
	},
	"languages": {
		ReadPath:     "/restapi/v1.0/dictionary/language",
		RecordsField: records,
	},
	"countries": {
		ReadPath:     "/restapi/v1.0/dictionary/country",
		RecordsField: records,
	},
	"locations": {
		ReadPath:     "/restapi/v1.0/dictionary/location",
		RecordsField: records,
	},
	"states": {
		ReadPath:     "/restapi/v1.0/dictionary/state",
		RecordsField: records,
	},
	"timezones": {
		ReadPath:     "/restapi/v1.0/dictionary/timezone",
		RecordsField: records,
	},
	"permissions": {
		ReadPath:     "/restapi/v1.0/dictionary/permission",
		RecordsField: records,
	},
	"permission-category": {
		ReadPath:     "/restapi/v1.0/dictionary/permission-category",
		RecordsField: records,
	},
	"extension/grant": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/grant",
		RecordsField: records,
	},
	"user/features": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/features",
		RecordsField: records,
	},
	"emergency-locations": {
		ReadPath:     "restapi/v1.0/account/~/emergency-locations",
		RecordsField: records,
	},
	"user/emergency-locations": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/emergency-locations",
		RecordsField: records,
	},
	"users": {
		ReadPath:     "restapi/v1.0/account/~/emergency-address-auto-update/users",
		RecordsField: records,
	},
	"wireless-points": {
		ReadPath:     "restapi/v1.0/account/~/emergency-address-auto-update/wireless-points",
		RecordsField: records,
	},
	"networks": {
		ReadPath:     "restapi/v1.0/account/~/emergency-address-auto-update/networks",
		RecordsField: records,
	},
	"devices": {
		ReadPath:     "restapi/v1.0/account/~/emergency-address-auto-update/devices",
		RecordsField: records,
	},
	"switches": {
		ReadPath:     "restapi/v1.0/account/~/emergency-address-auto-update/switches",
		RecordsField: records,
	},
	"extension/devices": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/device",
		RecordsField: records,
	},
	"extensions": {
		ReadPath:     "restapi/v1.0/account/~/extension",
		RecordsField: records,
	},
	"user/templates": {
		ReadPath:     "restapi/v1.0/account/~/templates",
		RecordsField: records,
	},
	"resourecTpes": {
		ReadPath:     "scim/v2/ResourceTypes",
		RecordsField: "Resources",
	},
	"schemas": {
		ReadPath:     "scim/v2/Schemas",
		RecordsField: "Resources",
	},
	"direct-routing/users": {
		ReadPath:     "restapi/v1.0/account/~/ms-teams/direct-routing/users",
		RecordsField: "mappings",
	},
	"contacts": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/address-book/contact",
		RecordsField: records,
	},
	"favorite/contacts": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/favorite",
		RecordsField: records,
	},
	"directory/entries": {
		ReadPath:     "restapi/v1.0/account/~/directory/entries",
		RecordsField: records,
	},
	"directory/federation": {
		ReadPath:     "restapi/v1.0/account/~/directory/federation",
		RecordsField: records,
	},
	"company/assigned-role": {
		ReadPath:     "restapi/v1.0/account/~/assigned-role",
		RecordsField: records,
	},
	"user/assigned-role": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/assigned-role",
		RecordsField: records,
	},
	"user-role": {
		ReadPath:     "/restapi/v1.0/dictionary/user-role",
		RecordsField: records,
	},
	"company/user-role": {
		ReadPath:     "restapi/v1.0/account/~/user-role",
		RecordsField: records,
	},
	"assignable-roles": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/assignable-roles",
		RecordsField: records,
	},
	"administered-sites": {
		ReadPath:     "restapi/v1.0/account/~/extension/~/administered-sites",
		RecordsField: records,
	},
	"subscriptions": {
		ReadPath:     "restapi/v1.0/subscription",
		RecordsField: records,
	},
}

func GetFieldByJSONTag(resp *Response, jsonTag string) ([]map[string]any, error) {
	v := reflect.ValueOf(resp).Elem()
	t := v.Type()

	for i := range t.NumField() {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// Match the JSON tag
		if tag == jsonTag {
			fieldValue := v.Field(i)

			if fieldValue.Kind() != reflect.Slice {
				return nil, fmt.Errorf("field with tag '%s' is not a slice", jsonTag) //nolint: err113
			}

			result, ok := fieldValue.Interface().([]map[string]any)
			if !ok {
				return nil, fmt.Errorf("field with tag '%s' is not of type []map[string]any", jsonTag) //nolint: err113
			}

			return result, nil
		}
	}

	return nil, fmt.Errorf("field with json tag '%s' not found", jsonTag) // nolint: err113
}

func inferValue(value any) common.ValueType {
	v := reflect.ValueOf(value)

	switch v.Kind() { //nolint: exhaustive
	case reflect.String:
		return common.ValueTypeString
	case reflect.Float64:
		return common.ValueTypeFloat
	case reflect.Bool:
		return common.ValueTypeBoolean
	case reflect.Slice:
		return common.ValueTypeOther
	case reflect.Map:
		return common.ValueTypeOther
	default:
		return common.ValueTypeOther
	}
}

func nextRecordsURL(objectName string, url *urlbuilder.URL) common.NextPageFunc { //nolint: gocognit,cyclop,funlen
	return func(node *ajson.Node) (string, error) {
		objectInfo, exists := pathURLs[objectName]
		if !exists {
			return "", fmt.Errorf("error couldn't construct read url for object: %s", objectName) //nolint: err113
		}

		switch {
		case objectInfo.usesOffsetPagination:
			page, err := jsonquery.New(node, "paging").IntegerOptional("page")
			if err != nil {
				return "", err
			}

			totalPages, err := jsonquery.New(node, "paging").IntegerOptional("totalPages")
			if err != nil {
				return "", err
			}

			if *page < *totalPages {
				nextPage := *page + 1

				url.WithQueryParam("page", strconv.Itoa(int(nextPage)))

				return url.String(), nil
			}

		case objectInfo.usesCursorPagination:
			nextPageToken, err := jsonquery.New(node, "navigation").StringOptional("nextPageToken")
			if err != nil {
				return "", err
			}

			if nextPageToken == nil {
				nextPageToken, err = jsonquery.New(node, "pagination").StringOptional("nextPageToken")
				if err != nil {
					return "", err
				}
			}

			if nextPageToken != nil {
				url.WithQueryParam("pageToken", *nextPageToken)

				return url.String(), nil
			}

		default:
			nextPage, err := jsonquery.New(node, "navigation").StringOptional("nextPage")
			if err != nil {
				return "", err
			}

			if nextPage == nil {
				return "", nil
			}

			url, err = urlbuilder.New(*nextPage)
			if err != nil {
				return "", err
			}

			return url.String(), nil
		}

		return "", nil
	}
}

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
		},
	}
}

var creationTimeFrom = datautils.NewSet("webinars", "webinar/recordings", // nolint: gochecknoglobals
	"webinar/company/recordings")
