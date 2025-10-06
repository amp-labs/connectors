package gong

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

// readFlows handles the special case for reading flows which requires flowOwnerEmail query param.
// It fetches all users first, then iterates through their emails to fetch flows for each user.
// Some users may not have engage license or may not be added to flows, so we handle errors gracefully.
// ref: https://gong.app.gong.io/settings/api/documentation#get-/v2/flows
func (c *Connector) readFlows(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { // nolint:cyclop,lll
	users, err := c.fetchAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		return &common.ReadResult{
			Rows:     0,
			Data:     nil,
			NextPage: "",
			Done:     true,
		}, nil
	}

	for _, user := range users {
		userEmail, ok := user.Raw["emailAddress"].(string)
		if !ok || userEmail == "" {
			continue
		}

		flows, err := c.fetchFlowsForUser(ctx, userEmail, config)
		if err != nil || len(flows) == 0 {
			// Some users may not have engage license or may not be added to flows
			// we ignore these errors and continue
			continue
		}

		// Return as soon as we find flows for a user
		return &common.ReadResult{
			Rows:     int64(len(flows)),
			Data:     flows,
			NextPage: "",
			Done:     true,
		}, nil
	}

	return &common.ReadResult{
		Rows:     0,
		Data:     nil,
		NextPage: "",
		Done:     true,
	}, nil
}

// fetchAllUsers retrieves all users from the /users endpoint.
func (c *Connector) fetchAllUsers(ctx context.Context) ([]common.ReadResultRow, error) {
	url, err := c.getReadURL(objectNameUsers)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectNameUsers)

	result, err := common.ParseResult(res,
		common.ExtractRecordsFromPath(responseFieldName),
		getNextRecordsURL,
		common.GetMarshaledData,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// fetchFlowsForUser fetches flows for a specific user email.
func (c *Connector) fetchFlowsForUser(ctx context.Context, userEmail string, config common.ReadParams,
) ([]common.ReadResultRow, error) {
	url, err := c.getReadURL(objectNameFlows)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("flowOwnerEmail", userEmail)

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectNameFlows)

	result, err := common.ParseResult(res,
		common.ExtractRecordsFromPath(responseFieldName),
		getNextRecordsURL,
		common.GetMarshaledData,
		config.Fields,
	)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}
