package gong

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

// readFlows handles the special case for reading flows which requires flowOwnerEmail query param.
// It fetches all users first, then iterates through their emails to fetch flows for each user.
// Flows are aggregated across users and deduplicated by id, because shared/company flows may
// appear in multiple users' result sets while personal flows are unique to their owner.
// Some users may not have engage license or may not be added to flows, so we handle errors gracefully.
// ref: https://gong.app.gong.io/settings/api/documentation#get-/v2/flows
func (c *Connector) readFlows(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { // nolint:cyclop,lll
	users, err := c.fetchAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		return emptyFlowsResult(), nil
	}

	includeUserAssoc := wantsUserAssociation(config.AssociatedObjects)

	// Aggregate flows across all users, keyed by id to dedupe shared/company flows
	// that surface in multiple users' result sets.
	aggregated := make(map[string]common.ReadResultRow)

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

		// Attach user association to flows
		for _, flow := range flows {
			if _, seen := aggregated[flow.Id]; seen {
				continue
			}

			if includeUserAssoc {
				attachUserToPersonalFlow(&flow, user)
			}

			aggregated[flow.Id] = flow
		}
	}

	data := make([]common.ReadResultRow, 0, len(aggregated))
	for _, flow := range aggregated {
		data = append(data, flow)
	}

	// Stable order: map iteration is random in Go.
	sort.Slice(data, func(i, j int) bool {
		return data[i].Id < data[j].Id
	})

	return &common.ReadResult{
		Rows:     int64(len(data)),
		Data:     data,
		NextPage: "",
		Done:     true,
	}, nil
}

func emptyFlowsResult() *common.ReadResult {
	return &common.ReadResult{
		Rows:     0,
		Data:     nil,
		NextPage: "",
		Done:     true,
	}
}

// wantsUserAssociation reports whether the caller asked for user records to be
// attached alongside each flow.
func wantsUserAssociation(associatedObjects []string) bool {
	for _, name := range associatedObjects {
		switch strings.ToLower(name) {
		case "user", "users":
			return true
		}
	}

	return false
}

// attachUserToPersonalFlow attaches the owning user as an association if the flow
// has Personal visibility. Shared and Company flows are not user-specific.
func attachUserToPersonalFlow(flow *common.ReadResultRow, user common.ReadResultRow) {
	visibility, _ := flow.Raw["visibility"].(string)
	if !strings.EqualFold(visibility, "Personal") {
		return
	}

	if flow.Associations == nil {
		flow.Associations = make(map[string][]common.Association)
	}

	flow.Associations[objectNameUsers] = append(
		flow.Associations[objectNameUsers],
		common.Association{
			ObjectId: user.Id,
			Raw:      user.Raw,
		},
	)
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
