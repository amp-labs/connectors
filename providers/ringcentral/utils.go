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
	ReadPath             string `json:"read_path"`
	WritePath            string `json:"write_path"`
	UpdateMethod         string `json:"update_method"`
	RecordsField         string `json:"records_field"`
	UsesCursorPagination bool   `json:"uses_cursor_pagination"`
	UsesOffsetPagination bool   `json:"uses_offset_pagination"`
	UsesSyncToken        bool   `json:"uses_sync_token"`
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
		case objectInfo.UsesOffsetPagination:
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

		case objectInfo.UsesCursorPagination:
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
	writeSupport := []string{"*"}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
		},
	}
}

var creationTimeFrom = datautils.NewSet("webinars", "webinar/recordings", // nolint: gochecknoglobals
	"webinar/company/recordings")

var dateFromObjects = datautils.NewSet("call-log", "call-log-sync", // nolint: gochecknoglobals
	"message-store", "a2p-sms/batches", "a2p-sms/messages")
