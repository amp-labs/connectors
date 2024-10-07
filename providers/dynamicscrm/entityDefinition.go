package dynamicscrm

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Make a call to EntityDefinition endpoint.
// We are looking for one field: DisplayCollectionName.
// This is the formal name of a collection that this ObjectName represents.
func (c *Connector) getObjectDisplayName(
	ctx context.Context, objectName naming.SingularString,
) (string, error) {
	url, err := c.getEntityDefinitionURL(objectName)
	if err != nil {
		return "", err
	}

	// the only field we care about in response
	url.WithQueryParam("$select", "DisplayCollectionName")

	body, err := c.performGetRequest(ctx, url)
	if err != nil {
		return "", err
	}

	objectDisplayName, err := getObjectDisplayName(body, objectName)
	if err != nil {
		return "", err
	}

	return objectDisplayName, nil
}

func getObjectDisplayName(item *ajson.Node, objectName naming.SingularString) (string, error) {
	// Location of string from root: DisplayCollectionName->LocalizedLabels->Label
	// If it is not found default to object name pluralised, thats the best we can do.
	displayName, err := jsonquery.New(item, "DisplayCollectionName", "LocalizedLabels").
		StrWithDefault("Label", objectName.Plural().String())
	if err != nil {
		return "", errors.Join(ErrObjectNotFound, err)
	}

	return displayName, nil
}

// Internal GET request, where we expect JSON payload.
func (c *Connector) performGetRequest(ctx context.Context, url *urlbuilder.URL) (*ajson.Node, error) {
	rsp, err := c.JSON.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := rsp.Body()
	if !ok {
		return nil, errors.Join(ErrObjectNotFound, common.ErrEmptyJSONHTTPResponse)
	}

	return body, nil
}

func (c *Connector) getEntityDefinitionURL(arg naming.SingularString) (*urlbuilder.URL, error) {
	// This endpoint returns schema of an object.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')", arg.String())

	return c.getURL(path)
}

func (c *Connector) getEntityAttributesURL(arg naming.SingularString) (*urlbuilder.URL, error) {
	// This endpoint will describe attributes present on schema and its properties.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')/Attributes", arg.String())

	return c.getURL(path)
}

func (c *Connector) getURL(arg string) (*urlbuilder.URL, error) {
	return constructURL(c.BaseURL(), apiVersion, arg)
}
