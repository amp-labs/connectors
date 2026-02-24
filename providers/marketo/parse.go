package marketo

import (
	"slices"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

// getNextRecordsURL returns the URL for the next page of results.
func getNextRecordsURL(node *ajson.Node) (string, error) {
	return jsonquery.New(node).StrWithDefault("nextPageToken", "")
}

func constructNextRecordsURL(object, nextToken string) common.NextPageFunc {
	if paginatesByIDs(object) {
		// Incase of Reading Records from the Objects requiring Filtering.
		// we construct Next-Page URLs using the filtered ids.
		// constructNextPageFilteredURL creates the next-page url by appending the next page ids in the query parameters
		return constructNextPageFilteredURL
	}

	if object == leads { // setting the activity NextPageToken for Leads
		return func(n *ajson.Node) (string, error) {
			return nextToken, nil
		}
	}

	return getNextRecordsURL
}

func constructNextPageFilteredURL(node *ajson.Node) (string, error) {
	jsonParser := jsonquery.New(node)

	data, err := jsonParser.ArrayRequired("result")
	if err != nil {
		return "", err
	}

	// If the records returned matches the maximum batchsize, there is a high probability of having more records.
	// We'd have to check for the next page records, also due deletes the is also a probability of having more records
	// even if the size do not reach 300.
	if len(data) > 0 {
		lastRecordID, err := jsonquery.New(data[len(data)-1]).IntegerRequired("id")
		if err != nil {
			return "", err
		}

		return strconv.Itoa(int(lastRecordID) + 1), nil
	}

	return "", nil
}

// getRecords returns the records from the response.
func getRecords(node *ajson.Node) ([]map[string]any, error) {
	result, err := jsonquery.New(node).ArrayOptional("result")
	if err != nil {
		return nil, err
	}

	return jsonquery.Convertor.ArrayToMap(result)
}

func usesStandardId(object string) bool {
	return slices.Contains(IdResponseObjects, object)
}

func usesMarketoGUID(object string) bool {
	return slices.Contains(marketoGUIDResponseObjects, object)
}

// constructStaticNextPageURL constructs the next page url for static API.
// static API uses offset pagination, while the leads API uses cusrsor pagination.
// Ex: https://experienceleague.adobe.com/en/docs/marketo-developer/marketo/rest/channels
func constructStaticNextPageURL(url *urlbuilder.URL) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		jsonParser := jsonquery.New(node)

		var offset string

		off, ok := url.GetFirstQueryParam("offset")
		if !ok {
			off = "0"
		}

		prevOffset, err := strconv.Atoi(off)
		if err != nil {
			return "", err
		}

		data, err := jsonParser.ArrayOptional("result")
		if err != nil {
			return "", err
		}

		if len(data) == maxReturn {
			newOffset := maxReturn + prevOffset
			offset = strconv.Itoa(newOffset)
		}

		url.WithQueryParam("offset", offset)

		return url.String(), nil
	}
}

func nextRecordsURL(objectName string, url *urlbuilder.URL) common.NextPageFunc {
	// reading the assets API uses offset pagination, while Leads API uses cursor
	//  pagination.
	// ref: https://developer.adobe.com/marketo-apis/api/asset/#operation/getAllChannelsUsingGET
	// ref: https://developer.adobe.com/marketo-apis/api/mapi
	if assetsObjects.Has(objectName) {
		return constructStaticNextPageURL(url)
	}

	return getNextRecordsURL
}
