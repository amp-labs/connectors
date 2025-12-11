package snapchatads

import (
	"path"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const defaultPageSize = 100

var incrementalReadObject = datautils.NewSet( //nolint:gochecknoglobals
	"transactions",
)

// Endpoints that require an organization ID (provided via metadata) as shared ID in the URL path.
var endpointsRequiringOrganizationMetadata = datautils.NewSet( //nolint:gochecknoglobals
	"fundingsources",
	"billingcenters",
	"transactions",
	"adaccounts",
	"members",
	"roles",
)

func (c *Connector) constructURL(objName string) (*urlbuilder.URL, error) {
	if endpointsRequiringOrganizationMetadata.Has(objName) {
		// If it needs shared ID, build URL with organizationId in the path.
		return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, "organizations", c.organizationId, objName)
	}

	// Otherwise, build the normal URL.
	return urlbuilder.New(c.ProviderInfo().BaseURL, apiVersion, objName)
}

func getObjectNodePath(objName string) string {
	// Determine correct node path.
	// For all direct objects, the node path have "targeting_dimensions".
	nodePath := "targeting_dimensions"

	if endpointsRequiringOrganizationMetadata.Has(objName) {
		nodePath = objName
	}

	return nodePath
}

func makeNextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		pagination, err := jsonquery.New(node).ObjectOptional("paging")
		if err != nil {
			return "", err
		}

		if pagination != nil {
			nextLink, err := jsonquery.New(pagination).StringOptional("next_link")
			if err != nil {
				return "", err
			}

			if nextLink != nil {
				return *nextLink, nil
			}
		}

		return "", nil
	}
}

// MarshalledData to extract selected fields from the read response, implement the DataMarshaller.
// This is necessary because the important fields are embedded as an object,
// where the key is the singular form of the object name, inside an array
// whose key is the plural form of the object name.
type MarshalledData func([]map[string]any, []string) ([]common.ReadResultRow, error)

func DataMarshall(resp *common.JSONHTTPResponse, objName string) MarshalledData {
	return func(records []map[string]any, fields []string) ([]common.ReadResultRow, error) {
		node, ok := resp.Body()
		if !ok {
			return nil, common.ErrEmptyJSONHTTPResponse
		}

		arr, err := jsonquery.New(node).ArrayOptional(getObjectNodePath(objName))
		if err != nil {
			return nil, err
		}

		return getRecords(objName, arr, fields)
	}
}

func getRecords(objName string, records []*ajson.Node, fields []string,
) ([]common.ReadResultRow, error) {
	data := make([]common.ReadResultRow, len(records))

	objKey := naming.NewSingularString(objName).String()

	// We derive the objKey for direct objects as the last path segment of the object name.
	// Ex: Below is a response for the direct object(no sharedId in the url path) 'targeting/device/os_type'.
	// Here, the objKey is the last path segment of the object name.
	// "targeting_dimensions": [
	//     {
	//         "sub_request_status": "SUCCESS",
	//         "os_type": { -------> objKey
	//             "id": "1",
	//             "name": "iOS"
	//         }
	//     },
	if !endpointsRequiringOrganizationMetadata.Has(objName) {
		objKey = path.Base(objName)
	}

	for i, element := range records { // nolint:varnamelen
		originalRecord, err := jsonquery.Convertor.ObjectToMap(element)
		if err != nil {
			return nil, err
		}

		values, err := jsonquery.New(element).ObjectRequired(objKey)
		if err != nil {
			return nil, err
		}

		fieldRecord, err := jsonquery.Convertor.ObjectToMap(values)
		if err != nil {
			return nil, err
		}

		data[i].Raw = originalRecord
		data[i].Fields = common.ExtractLowercaseFieldsFromRaw(fields, fieldRecord)
	}

	return data, nil
}
