package intercom

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/scrapper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/intercom/metadata"
)

var ErrLoadFailure = errors.New("cannot load metadata")

func (c *Connector) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	return c.listObjectMetadata(ctx, naming.NewSingularStrings(objectNames))
}

func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames naming.SingularStrings,
) (*common.ListObjectMetadataResult, error) {
	// Ensure that objectNames are not empty
	if len(objectNames) == 0 {
		return nil, common.ErrMissingObjects
	}

	schemas, err := metadata.LoadSchemas()
	if err != nil {
		return nil, ErrLoadFailure
	}

	for _, objectName := range objectNames {
		if err = c.substituteSchema(ctx, schemas, objectName); err != nil {
			return nil, err
		}
	}

	result, err := schemas.Select(objectNames.Plural())
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Connector) substituteSchema(ctx context.Context, schemas *scrapper.ObjectMetadataResult,
	objectName naming.SingularString,
) error {
	url, err := c.buildMetaURL(objectName)
	if err != nil {
		return err
	}

	rsp, err := c.get(ctx, url.String(), common.Header{
		Key:   "Intercom-Version",
		Value: c.Module,
	})
	if err != nil {
		if errors.Is(err, common.ErrBadRequest) {
			// We have BadRequest. No data attributes available for this model.
			// Nothing to replace in schema registry.
			return nil
		} else {
			// critical error
			return err
		}
	}

	attributes, err := extractDataAttributes(rsp)
	if err != nil {
		return err
	}
	// Replace fields that we got from ListDataAttributes endpoint.
	obj := schemas.Result[objectName.Plural().String()]
	for k, v := range attributes {
		obj.FieldsMap[k] = v
	}

	return nil
}

func extractDataAttributes(rsp *common.JSONHTTPResponse) (map[string]string, error) {
	attributes := make(map[string]string)

	arr, err := jsonquery.New(rsp.Body).Array("data", false)
	if err != nil {
		return nil, err
	}

	for _, node := range arr {
		name, err := jsonquery.New(node).Str("name", false)
		if err != nil {
			return nil, err
		}

		displayName, err := jsonquery.New(node).Str("label", false)
		if err != nil {
			return nil, err
		}

		attributes[*name] = *displayName
	}

	return attributes, nil
}

func (c *Connector) buildMetaURL(objectName naming.SingularString) (*urlbuilder.URL, error) {
	// Metadata is described by the following endpoint:
	//
	// GET /data_attributes?model=objectName
	//
	url, err := c.getURL("data_attributes")
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("model", objectName.String())

	return url, nil
}
