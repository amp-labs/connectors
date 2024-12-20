package iterable

import (
	"context"
	"errors"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

var ErrCatalogCreate = errors.New("payload must have string 'name' to create a catalog")

func (c *Connector) Write(
	ctx context.Context, config common.WriteParams,
) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Object name will be used to complete further URL construction below.
	url, err := c.getWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if config.ObjectName == objectNameCatalogs {
		if err = createCatalog(config, url); err != nil {
			return nil, err
		}
	}

	res, err := c.Client.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordIdNodePath := recordIdPaths[config.ObjectName]

	// write response was with payload
	return constructWriteResult(body, recordIdNodePath)
}

// This is the only endpoint that doesn't use payload.
// New catalog requires a name which is passed as path parameter in URL.
// However, the write connector will accept payload of shape:
//
//	{
//		"name": "catalogName"
//	}
func createCatalog(config common.WriteParams, url *urlbuilder.URL) error {
	payload, isJSON := config.RecordData.(map[string]any)
	if !isJSON {
		return ErrCatalogCreate
	}

	name, found := payload["name"]
	if !found {
		return ErrCatalogCreate
	}

	catalogName, isString := name.(string)
	if !isString {
		return ErrCatalogCreate
	}

	url.AddPath(catalogName)

	return nil
}

type path struct {
	id   string
	zoom []string
}

func newPath(id string, zoom ...string) path {
	return path{
		id:   id,
		zoom: zoom,
	}
}

var recordIdPaths = datautils.Map[string, path]{ // nolint:gochecknoglobals
	"campaigns":      newPath("campaignId"),
	"catalogs":       newPath("id", "params"),
	"lists":          newPath("listId"),
	"templatesEmail": newPath("TODO", "params"), // TODO template ID
	"templatesInApp": newPath("TODO", "params"), // TODO template ID
	"templatesPush":  newPath("TODO", "params"), // TODO template ID
	"templatesSMS":   newPath("TODO", "params"), // TODO template ID
	"users":          newPath("TODO", "params"), // TODO template ID
	"webhooks":       newPath("id"),
}

func constructWriteResult(body *ajson.Node, recordIdLocation path) (*common.WriteResult, error) {
	// ID is integer that is always stored under different field name.
	intIdentifier, err := jsonquery.New(body, recordIdLocation.zoom...).Integer(recordIdLocation.id, false)
	if err != nil {
		return nil, err
	}

	recordID := strconv.FormatInt(*intIdentifier, 10)

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
