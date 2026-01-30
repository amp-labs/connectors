package phoneburner

import (
	"context"
	"net/http"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

const (
	restPrefix = "rest"
	restVer    = "1"
)

var (
	paginatedObjects = datautils.NewStringSet("contacts", "members", "tags", "voicemails")
)

func buildReadRequest(ctx context.Context, baseURL string, params common.ReadParams) (*http.Request, error) {
	if len(params.NextPage) != 0 {
		// Next page.
		url, err := urlbuilder.New(params.NextPage.String())
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Accept", "application/json")

		return req, nil
	}

	url, err := urlbuilder.New(baseURL, restPrefix, restVer, params.ObjectName)
	if err != nil {
		return nil, err
	}

	// Many PhoneBurner list endpoints support page-based pagination. Defaulting these parameters
	// makes NextPage generation deterministic.
	if paginatedObjects.Has(params.ObjectName) {
		url.WithQueryParam("page_size", strconv.Itoa(100))
		url.WithQueryParam("page", "1")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

func parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	_ = ctx

	url, err := urlbuilder.FromRawURL(request.URL)
	if err != nil {
		return nil, err
	}

	records, err := recordsFunc(params.ObjectName)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		response,
		records,
		nextRecordsURL(url.String(), params.ObjectName),
		common.GetMarshaledData,
		params.Fields,
	)
}

func recordsFunc(objectName string) (common.RecordsFunc, error) {
	switch objectName {
	case "contacts":
		return common.ExtractRecordsFromPath("contacts", "contacts"), nil
	case "members":
		return common.ExtractRecordsFromPath("members", "members"), nil
	case "tags":
		return common.ExtractRecordsFromPath("tags", "tags"), nil
	case "voicemails":
		return common.ExtractRecordsFromPath("voicemails", "voicemails"), nil
	case "folders":
		return extractFoldersRecords(), nil
	default:
		return nil, common.ErrOperationNotSupportedForObject
	}
}

func nextRecordsURL(requestURL string, objectName string) common.NextPageFunc {
	if !paginatedObjects.Has(objectName) {
		return func(*ajson.Node) (string, error) { return "", nil }
	}

	return func(node *ajson.Node) (string, error) {
		wrapper, err := jsonquery.New(node).ObjectRequired(objectName)
		if err != nil {
			return "", err
		}

		page, err := jsonquery.New(wrapper).IntegerWithDefault("page", 0)
		if err != nil {
			return "", err
		}

		totalPages, err := jsonquery.New(wrapper).IntegerWithDefault("total_pages", 0)
		if err != nil {
			return "", err
		}

		if totalPages == 0 || page >= totalPages {
			return "", nil
		}

		nextURL, err := urlbuilder.New(requestURL)
		if err != nil {
			return "", err
		}

		nextURL.WithQueryParam("page", strconv.Itoa(int(page)+1))

		return nextURL.String(), nil
	}
}

func extractFoldersRecords() common.RecordsFunc {
	return func(node *ajson.Node) ([]map[string]any, error) {
		foldersNode, err := jsonquery.New(node).ObjectRequired("folders")
		if err != nil {
			return nil, err
		}

		m, err := jsonquery.Convertor.ObjectToMap(foldersNode)
		if err != nil {
			return nil, err
		}

		out := make([]map[string]any, 0, len(m))

		for _, v := range m {
			obj, ok := v.(map[string]any)
			if !ok || obj == nil {
				continue
			}

			out = append(out, obj)
		}

		return out, nil
	}
}

