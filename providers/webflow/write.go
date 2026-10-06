package webflow

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	recordPlaceholder = "{record_id}"

	// Registered scripts have two create endpoints. A payload carrying
	// sourceCode registers an inline script; otherwise a hosted one
	// (hostedLocation + integrityHash).
	registeredScriptsHostedPath = "/v2/sites/{site_id}/registered_scripts/hosted"
	registeredScriptsInlinePath = "/v2/sites/{site_id}/registered_scripts/inline"
	fieldSourceCode             = "sourceCode"

	fieldID            = "id"
	fieldSKU           = "sku"
	fieldPublishStatus = "publishStatus"
)

// writeSpec is the verified per-object write behaviour.
type writeSpec struct {
	// createPath is the POST path; empty means the object cannot be created.
	createPath string
	// updatePath is the item path with {record_id}; empty means no update.
	updatePath string
	// updateMethod defaults to PATCH; pages use PUT.
	updateMethod string
	// idField is the response field returned as WriteResult.RecordId.
	// Defaults to "id".
	idField string
	// createIDZoom / updateIDZoom locate the written record inside a response
	// envelope (products create returns {product, skus}; custom fonts return
	// {customFont}).
	createIDZoom []string
	updateIDZoom []string
}

// writeSpecs documents, per object, the create and update endpoints from the
// Webflow OpenAPI spec. Paths are explicit because several update endpoints
// are not nested under the site (collections, pages, assets, sites).
//
// Payloads are sent as given (application/json). Webflow ignores unknown and
// read-only fields on PATCH (verified live on assets) but validates field
// types strictly, so null values copied from a read record are rejected.
//
// Not writable: activity_logs, components, custom_code_blocks,
// custom_domains, forms and google_tags have no per-record write endpoint.
//
//nolint:gochecknoglobals
var writeSpecs = map[string]writeSpec{
	// Update only. Site creation needs a workspace id the connection does not
	// carry (POST /v2/workspaces/{workspace_id}/sites).
	objectSites: {updatePath: "/v2/sites/{record_id}"},
	objectPages: {updatePath: "/v2/pages/{record_id}", updateMethod: http.MethodPut},
	objectFormSubmissions: {
		updatePath: "/v2/sites/{site_id}/form_submissions/{record_id}",
	},
	// Resolving a thread is its only mutation (body: {resolved: bool}).
	objectComments: {updatePath: "/v2/sites/{site_id}/comments/{record_id}"},
	objectOrders: {
		updatePath: "/v2/sites/{site_id}/orders/{record_id}",
		idField:    "orderId",
	},

	// Create and update.
	objectCollections: {
		createPath: "/v2/sites/{site_id}/collections",
		updatePath: "/v2/collections/{record_id}",
	},
	// Create registers an upload (fileName + fileHash) and returns the asset
	// id plus S3 upload details; the file itself is uploaded by the caller.
	objectAssets: {
		createPath: "/v2/sites/{site_id}/assets",
		updatePath: "/v2/assets/{record_id}",
	},
	objectCustomFonts: {
		createPath:   "/v2/sites/{site_id}/custom_fonts",
		updatePath:   "/v2/sites/{site_id}/custom_fonts/{record_id}",
		createIDZoom: []string{"customFont"},
		updateIDZoom: []string{"customFont"},
	},
	// The API takes {product, sku, publishStatus}; records in the flattened
	// shape Read returns are wrapped into it (see adaptProductRecord). Create
	// answers {product, skus}; update answers the product.
	objectProducts: {
		createPath:   "/v2/sites/{site_id}/products",
		updatePath:   "/v2/sites/{site_id}/products/{record_id}",
		createIDZoom: []string{fieldProduct},
	},
	objectRedirects: {
		createPath: "/v2/sites/{site_id}/redirects",
		updatePath: "/v2/sites/{site_id}/redirects/{record_id}",
	},

	// Create only.
	objectAssetFolders:      {createPath: "/v2/sites/{site_id}/asset_folders"},
	objectWebhooks:          {createPath: "/v2/sites/{site_id}/webhooks"},
	objectRegisteredScripts: {createPath: registeredScriptsHostedPath},
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	spec, ok := writeSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	if params.ObjectName == objectProducts {
		record = adaptProductRecord(record)
	}

	url, method, err := c.buildWriteURL(params, spec, record)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Connector) buildWriteURL(
	params common.WriteParams, spec writeSpec, record common.Record,
) (*urlbuilder.URL, string, error) {
	var (
		path   string
		method string
	)

	switch {
	case params.IsCreate():
		path = spec.createPath
		method = http.MethodPost

		if params.ObjectName == objectRegisteredScripts {
			if _, inline := record[fieldSourceCode]; inline {
				path = registeredScriptsInlinePath
			}
		}
	default:
		path = strings.ReplaceAll(spec.updatePath, recordPlaceholder, params.RecordId)

		method = spec.updateMethod
		if method == "" {
			method = http.MethodPatch
		}
	}

	if path == "" {
		// Create-only or update-only object.
		return nil, "", common.ErrOperationNotSupportedForObject
	}

	path, err := c.substituteSiteID(path)
	if err != nil {
		return nil, "", err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, path)
	if err != nil {
		return nil, "", err
	}

	return url, method, nil
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || body == nil {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	spec := writeSpecs[params.ObjectName]

	zoom := spec.updateIDZoom
	if params.IsCreate() {
		zoom = spec.createIDZoom
	}

	idField := spec.idField
	if idField == "" {
		idField = fieldID
	}

	recordID, err := jsonquery.New(body, zoom...).TextWithDefault(idField, "")
	if err != nil {
		return nil, err
	}

	if recordID == "" {
		recordID = params.RecordId
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}

// adaptProductRecord accepts a product in the flattened shape Read returns
// (product fields at the top level, SKUs under "skus") and builds the
// {product, sku, publishStatus} body the API expects. A record that already
// carries a "product" key is taken as the API shape and passed through.
//
// The API writes one SKU per call: "sku" is used as given, otherwise the
// first element of "skus" (create requires one; update accepts none).
func adaptProductRecord(record common.Record) common.Record {
	if _, isAPIShape := record[fieldProduct]; isAPIShape {
		return record
	}

	// Work on a copy: GetRecord hands back the caller's own map.
	product := maps.Clone(record)
	wrapped := common.Record{}

	if status, ok := product[fieldPublishStatus]; ok {
		wrapped[fieldPublishStatus] = status
		delete(product, fieldPublishStatus)
	}

	if sku, ok := product[fieldSKU]; ok {
		wrapped[fieldSKU] = sku
		delete(product, fieldSKU)
	}

	if skus, ok := product[fieldSkus].([]any); ok {
		if _, hasSKU := wrapped[fieldSKU]; !hasSKU && len(skus) > 0 {
			wrapped[fieldSKU] = skus[0]
		}

		delete(product, fieldSkus)
	}

	wrapped[fieldProduct] = product

	return wrapped
}
