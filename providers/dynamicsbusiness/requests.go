package dynamicsbusiness

import (
	"context"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common/naming"
)

// nolint:lll
// Microsoft Business Central has entityDefinitions which lists definition for each object.
// This can be scoped to retrieve single object using $filter query.
//
// Learn more about entity definitions:
// https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/powerplatform/powerplat-entity-modeling#labels-and-localization
// Finding API endpoint structure:
// https://learn.microsoft.com/en-us/dynamics365/business-central/dev-itpro/developer/devenv-develop-custom-api#to-create-api-pages-to-display-car-brand-and-car-model
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
