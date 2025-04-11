package atlassian

import (
	"context"
	"errors"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

var (
	ErrDiscoveryFailure  = errors.New("failed to collect post authentication data")
	ErrContainerNotFound = errors.New("cloud container was not found for chosen workspace")
)

func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	cloudId, err := c.retrieveCloudId(ctx)
	if err != nil {
		return nil, errors.Join(ErrDiscoveryFailure, err)
	}

	c.cloudId = cloudId

	return &common.PostAuthInfo{
		CatalogVars: AuthMetadataVars{
			CloudId: cloudId,
		}.AsMap(),
		RawResponse: nil,
	}, nil
}

// retrieveCloudId gives CloudId for the workspace.
// Cloud ID will be used to build URL paths.
// After authentication completes we can call introspect API to find this data.
//
// Request: Get Cloud ID.
// Response example:
// [
//
//	{
//	    "id": "9e1477fd-54ef-41fe-b747-bc9e6a11a925",
//	    "url": "https://{{workspaceRef}}.atlassian.net",
//	    "name": "{{workspaceRef}}",
//	    "scopes": [
//	        "manage:jira-project",
//	        "manage:jira-configuration",
//	        "read:jira-work",
//	        "manage:jira-webhook",
//	        "write:jira-work",
//	        "read:jira-user"
//	    ],
//	    "avatarUrl": "https://site-admin-avatar-cdn.prod.public.atl-paas.net/avatars/240/pencilmarker.png"
//	}
//
// ].
func (c *Connector) retrieveCloudId(ctx context.Context) (string, error) {
	url, err := c.getAccessibleSitesURL()
	if err != nil {
		return "", err
	}

	res, err := c.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return "", err
	}

	body, ok := res.Body()
	if !ok {
		return "", errors.Join(ErrContainerNotFound, common.ErrEmptyJSONHTTPResponse)
	}

	arr, err := body.GetArray()
	if err != nil {
		return "", err
	}

	for _, item := range arr {
		workspaceName, err := jsonquery.New(item).StringRequired("name")
		if err != nil {
			return "", err
		}

		if workspaceName == c.workspace {
			// Names match, select this container.
			// Returns cloudID.
			return jsonquery.New(item).StringRequired("id")
		}
	}

	// The container that matches connectors workspace was not found.
	// Hence, we couldn't resolve cloud id.
	return "", ErrContainerNotFound
}
