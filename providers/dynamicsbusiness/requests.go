package dynamicsbusiness

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/httpkit"
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

const defaultPageSize = 100

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	url, err := c.buildReadURL(ctx, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Prefer", fmt.Sprintf("odata.maxpagesize=%v", defaultPageSize))

	return req, nil
}

// buildReadURL constructs a URL for a read operation.
// If params.NextPage is set, it returns the URL for the next page.
// Otherwise, it builds a fresh URL for the initial request, adding field selection
// and an optional incremental filter if applicable.
func (c *Connector) buildReadURL(ctx context.Context, params common.ReadParams) (*urlbuilder.URL, error) {
	if len(params.NextPage) != 0 {
		// Next page
		return constructURL(params.NextPage.String())
	}

	// First page
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	fields := params.Fields.List()
	if len(fields) != 0 {
		url.WithQueryParam("$select", strings.Join(fields, ","))
	}

	// Incremental read can be applied for the majority of objects using `lastModifiedDateTime`.
	// To avoid false assumptions, a probe request confirms its availability per object.
	if !params.Since.IsZero() {
		incremental, err := c.isIncrementObject(ctx, params)
		if err != nil {
			return nil, err
		}

		if incremental {
			applyIncrementalQuery(url, params.Since)
		}
	}

	return url, nil
}

func applyIncrementalQuery(url *urlbuilder.URL, since time.Time) {
	sinceValue := datautils.Time.FormatRFC3339inUTCWithMilliseconds(since)
	url.WithQueryParam("$filter", fmt.Sprintf("lastModifiedDateTime ge %v", sinceValue))
}

// isIncrementObject determines whether the given object supports incremental (time-based) reads.
//
// First checks the local registry cache. If no entry is found, it performs a one-time probe via API,
// then caches the result. Assumes provider behavior is stable and does not change support dynamically.
func (c *Connector) isIncrementObject(ctx context.Context, params common.ReadParams) (bool, error) {
	isIncremental, found := c.incrementalRegistry.Get(params.ObjectName)
	if !found {
		var err error

		isIncremental, err = c.fetchIsIncrementObject(ctx, params)
		if err != nil {
			return false, err
		}

		c.incrementalRegistry.Set(params.ObjectName, isIncremental)
	}

	return isIncremental, nil
}

// fetchIsIncrementObject makes a probe request to check if the object supports time-based filtering.
// This is determined by requesting one item and checking if the response is successful.
// Errors in the 5xx range indicate temporary failures; 4xx indicates lack of support.
func (c *Connector) fetchIsIncrementObject(ctx context.Context, params common.ReadParams) (bool, error) {
	url, err := c.getURL(params.ObjectName)
	if err != nil {
		return false, err
	}

	applyIncrementalQuery(url, params.Since)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return false, err
	}

	// Request a minimal response payload.
	req.Header.Set("Prefer", fmt.Sprintf("odata.maxpagesize=%v", 1))

	response, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		return false, err
	}

	defer response.Body.Close()

	// Probing failed. Cannot conclude support.
	if httpkit.Status5xx(response.StatusCode) {
		return false, fmt.Errorf(
			"%w: probing incremental capability for an object %v", common.ErrRequestFailed, params.ObjectName)
	}

	// Incremental querying is not supported by this object.
	if httpkit.Status4xx(response.StatusCode) {
		return false, nil
	}

	// 2xx indicates that query param can be accepted by the API.
	return httpkit.Status2xx(response.StatusCode), nil
}
