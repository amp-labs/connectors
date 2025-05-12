package dynamicsbusiness

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/naming"
)

func (c *Connector) buildSingleObjectMetadataRequest(ctx context.Context, objectName string) (*http.Request, error) {
	url, err := c.getMetadataURL()
	if err != nil {
		return nil, err
	}

	// Entity name is always singular.
	// There is `entitySetName` field which is plural but filtering using this property is not allowed.
	entityName := naming.NewSingularString(objectName).String()
	url.WithQueryParam("$filter", fmt.Sprintf("entityName eq '%v'", entityName))

	return http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
}
