package iterable

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
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

	// write response was with payload
	return constructWriteResult(body, config.ObjectName)
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

func newPath(id string, zoom ...string) *path {
	return &path{
		id:   id,
		zoom: zoom,
	}
}

var recordIDPaths = datautils.Map[string, *path]{ // nolint:gochecknoglobals
	"campaigns":      newPath("campaignId"),
	"catalogs":       newPath("id", "params"),
	"lists":          newPath("listId"),
	"templatesEmail": nil,
	"templatesInApp": nil,
	"templatesPush":  nil,
	"templatesSMS":   nil,
	"users":          nil,
	"webhooks":       newPath("id"),
}

func constructWriteResult(body *ajson.Node, objectName string) (*common.WriteResult, error) {
	recordID, err := locateWriteRecordID(body, objectName)
	if err != nil {
		return nil, err
	}

	if templateWriteObjects.Has(objectName) {
		// Template objects return ID as part of the message.
		// Other objects, return identifier using a dedicated property.
		// Sadly, string processing is needed to extract ID.
		recordID, err = extractTemplateWriteRecordID(body)
		if err != nil {
			return nil, err
		}
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}

func locateWriteRecordID(body *ajson.Node, objectName string) (string, error) {
	recordIDLocation := recordIDPaths[objectName]
	if recordIDLocation == nil {
		return "", nil
	}

	// ID is integer that is always stored under different field name.
	intIdentifier, err := jsonquery.New(body, recordIDLocation.zoom...).IntegerRequired(recordIDLocation.id)
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(intIdentifier, 10), nil
}

func extractTemplateWriteRecordID(body *ajson.Node) (string, error) {
	message, err := jsonquery.New(body).StringOptional("msg")
	if err != nil {
		return "", err
	}

	// Ex:
	//	Input: "Upserted 1 templates with IDs: 15939824"
	//	Output: "15939824"
	parts := strings.Split(*message, "IDs: ")
	if len(parts) != 2 { // nolint:mnd
		return "", nil
	}

	return parts[1], nil
}
