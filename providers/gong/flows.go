package gong

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/gong/metadata"
)

// readFlows handles the special case for reading flows which requires flowOwnerEmail query param.
// It fetches all users first, then iterates through their emails to fetch flows for each user.
// Flows are aggregated across users and deduplicated by id, because shared/company flows may
// appear in multiple users' result sets while personal flows are unique to their owner.
// Some users may not have engage license or may not be added to flows, so we handle errors gracefully.
// ref: https://gong.app.gong.io/settings/api/documentation#get-/v2/flows
func (c *Connector) readFlows(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { // nolint:lll,cyclop,funlen
	users, err := c.fetchAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch users: %w", err)
	}

	if len(users) == 0 {
		return emptyFlowsResult(), nil
	}

	includeUserAssoc := wantsUserAssociation(config.AssociatedObjects)

	readOpts, isReadOptsOk := config.Opts.(ReadParamsOpts)
	readAllUsers := isReadOptsOk && readOpts.ReadFlowsForAllUsers

	// Collect every user's flows into one map keyed by flow id.
	// Company and shared flows show up for many users, so the map drops the repeats.
	aggregated := make(map[string]common.ReadResultRow)

	for _, user := range users {
		userEmail, ok := user.Raw["emailAddress"].(string)
		if !ok || userEmail == "" {
			continue
		}

		// Only fetch flows for active users.
		if active, _ := user.Raw["active"].(bool); !active {
			continue
		}

		flows, err := c.fetchFlowsForUser(ctx, userEmail, config)
		if err != nil {
			// A user without an engage license (or not added to any flow) errors
			// here. We skip them so one user can't fail the whole sync, and log it
			// so a genuine failure (5xx / rate-limit / network) is still visible.
			logging.Logger(ctx).Warn("could not fetch flows for user, skipping",
				"user", userEmail, "error", err.Error())

			continue
		}

		if len(flows) == 0 {
			continue
		}

		mergeUserFlows(aggregated, flows, user, includeUserAssoc, readAllUsers)

		if !readAllUsers {
			// When ReadFlowsForAllUsers is false,
			// We only read flows with visibility "Company", so we only need to read flows from one user.
			// We treat failed readOpts assertion as false ReadFlowsForAllUsers.
			break
		}
	}

	if len(aggregated) == 0 {
		return emptyFlowsResult(), nil
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

// mergeUserFlows adds a user's flows to the aggregate, deduplicating by id and
// attaching the owning user to personal flows when requested.
//
// Gong reports visibility relative to the queried email: a flow reads "Personal"
// only in its owner's result set and "Shared" for everyone else who can see it.
// The same flow id therefore shows up in several users' results with different
// visibility, and users are fetched in an arbitrary order. So we can't just keep
// the first copy we see: if a non-owner is fetched first we'd store the "Shared"
// copy and never attach the owner. When this user owns the flow (their copy says
// "Personal") we prefer that copy, overwriting any "Shared" one stored earlier,
// so the owner association is reliable no matter what order users come back in.
//
// When readAllUsers is false we are in company-only mode, so flows that aren't
// "Company" visibility are dropped.
func mergeUserFlows(
	aggregated map[string]common.ReadResultRow,
	flows []common.ReadResultRow,
	owner common.ReadResultRow,
	includeUserAssoc bool,
	readAllUsers bool,
) {
	for _, flow := range flows {
		visibility, _ := flow.Raw["visibility"].(string)

		// Company-only mode: skip anything that isn't a company flow.
		if !readAllUsers && !strings.EqualFold(visibility, "Company") {
			continue
		}

		ownsFlow := strings.EqualFold(visibility, "Personal")

		// Already collected and this isn't the owner's copy: nothing to add.
		if _, seen := aggregated[flow.Id]; seen && !ownsFlow {
			continue
		}

		if includeUserAssoc {
			attachUserToPersonalFlow(&flow, owner)
		}

		aggregated[flow.Id] = flow
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

// attachUserToPersonalFlow links a flow to the user who owns it, but only for
// personal flows. Company and shared flows aren't tied to one person, so we skip them.
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

// fetchAllUsers retrieves every user from the /users endpoint, following
// pagination so users beyond the first page are not missed.
func (c *Connector) fetchAllUsers(ctx context.Context) ([]common.ReadResultRow, error) {
	url, err := c.getReadURL(objectNameUsers)
	if err != nil {
		return nil, err
	}

	return c.fetchAllPages(ctx, url, objectNameUsers, nil)
}

// fetchFlowsForUser fetches every flow for a specific user email, following
// pagination so a user with many flows does not lose the overflow.
func (c *Connector) fetchFlowsForUser(ctx context.Context, userEmail string, config common.ReadParams,
) ([]common.ReadResultRow, error) {
	url, err := c.getReadURL(objectNameFlows)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("flowOwnerEmail", userEmail)

	return c.fetchAllPages(ctx, url, objectNameFlows, config.Fields)
}

// fetchAllPages issues GET requests against url, following Gong's records.cursor
// until none is returned, and accumulates every page of rows.
func (c *Connector) fetchAllPages(
	ctx context.Context,
	url *urlbuilder.URL,
	objectName string,
	fields datautils.Set[string],
) ([]common.ReadResultRow, error) {
	responseFieldName := metadata.Schemas.LookupArrayFieldName(c.Module.ID, objectName)

	var rows []common.ReadResultRow

	for {
		res, err := c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		result, err := common.ParseResult(res,
			common.ExtractRecordsFromPath(responseFieldName),
			getNextRecordsURL,
			common.GetMarshaledData,
			fields,
		)
		if err != nil {
			return nil, err
		}

		rows = append(rows, result.Data...)

		if result.NextPage == "" {
			break
		}

		// Re-request with the next cursor; flowOwnerEmail (if set) is preserved on url.
		url.WithQueryParam("cursor", result.NextPage.String())
	}

	return rows, nil
}
