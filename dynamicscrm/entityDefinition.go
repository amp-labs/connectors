package dynamicscrm

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/naming"
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
