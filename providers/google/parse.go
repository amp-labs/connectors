package google

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

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

func getMarshaledData(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	withID := datautils.NewSet(fields...).Has("id")

	for i, record := range records {
		fieldsResult := common.ExtractLowercaseFieldsFromRaw(fields, record)
		if withID {
			if resourceName, ok := record["resourceName"].(string); ok {
				if _, recordID, ok := resourceIdentifierFormat(resourceName); ok {
					fieldsResult["id"] = recordID
				}
			}
		}

		data[i] = common.ReadResultRow{
			Fields: fieldsResult,
			Raw:    record,
		}
	}

	return data, nil
}
