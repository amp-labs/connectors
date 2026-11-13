package zoom

import (
	"context"
	"errors"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/zoom/internal/associations"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

func (c *Connector) Read(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if err := params.ValidateParams(true); err != nil {
		return nil, err
	}

	if !core.SupportedObjectsByRead.Has(params.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	if params.IsFirstPage() {
		return c.readFirstPage(ctx, params)
	}

	return c.readNextPage(ctx, params)
}

func (c *Connector) readFirstPage(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if core.ObjectCategoryAllUsers.Has(params.ObjectName) {
		return c.readForUsers(ctx, params)
	}

	url, err := c.GetReadURL(params.ObjectName)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", core.DefaultPageSize)
	addTimeFilterParams(params, url)

	return c.readRecords(ctx, params.ObjectName, params.Fields, url, params.AssociatedObjects)
}

func (c *Connector) readNextPage(ctx context.Context, params common.ReadParams) (*common.ReadResult, error) {
	if token, ok := readhelper.GetAggregateToken[string](params.NextPage); ok {
		return c.readNextPageForUsers(ctx, params, token)
	}

	// Default read of the next page.
	// It is not an aggregate token but a raw URL.
	url, err := urlbuilder.New(params.NextPage.String())
	if err != nil {
		return nil, err
	}

	return c.readRecords(ctx, params.ObjectName, params.Fields, url, params.AssociatedObjects)
}

func (c *Connector) readRecords(ctx context.Context,
	objectName string,
	selectedFields datautils.StringSet,
	url *urlbuilder.URL,
	associatedObjects []string,
) (*common.ReadResult, error) {
	rsp, err := c.Base.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	responseFieldName := objectNameToResponseField[objectName]
	identifier := objectNameToReadResponseIdentifier.Get(objectName)

	return common.ParseResult(
		rsp,
		common.MakeRecordsFunc(responseFieldName),
		getNextRecordURL(url),
		associations.CreateDataMarshallerWithAssociations(ctx,
			c.Base, common.ObjectName(objectName), associatedObjects, identifier,
		),
		selectedFields,
	)
}

func (c *Connector) readForUsers(
	ctx context.Context,
	params common.ReadParams,
) (*common.ReadResult, error) {
	userIds, err := c.listAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Parallelized read across all users.
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(userIds))
	for index, userId := range userIds {
		// Each task reads records for a specific user.
		url, err := c.GetReadUrlForUser(params.ObjectName, userId)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("page_size", core.DefaultPageSize)
		addTimeFilterParams(params, url)

		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			if err := core.Wait(ctx); err != nil {
				return userId, nil, err
			}

			result, err := c.readRecords(ctx, params.ObjectName, params.Fields, url, params.AssociatedObjects)

			return userId, result, err
		}
	}

	return executeReadTasksUsers(ctx, tasks)
}

func (c *Connector) readNextPageForUsers(ctx context.Context,
	params common.ReadParams,
	aggregateToken readhelper.AggregateNextPage[string],
) (*common.ReadResult, error) {
	tasks := make([]parallelfetch.Task[string, common.ReadResult], len(aggregateToken))
	for index, token := range aggregateToken {
		userId := token.Context
		nextPageToken := token.Value.String()

		url, err := urlbuilder.New(nextPageToken)
		if err != nil {
			return nil, err
		}

		// Create parallel task for this account's next page
		tasks[index] = func(ctx context.Context) (taskID string, data *common.ReadResult, err error) {
			if err := core.Wait(ctx); err != nil {
				return userId, nil, err
			}

			result, err := c.readRecords(ctx, params.ObjectName, params.Fields, url, params.AssociatedObjects)

			return userId, result, err
		}
	}

	return executeReadTasksUsers(ctx, tasks)
}

func executeReadTasksUsers(ctx context.Context,
	tasks []parallelfetch.Task[string, common.ReadResult],
) (*common.ReadResult, error) {
	result := parallelfetch.Execute(ctx, tasks, core.MaxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return readhelper.AggregateReadResults(
		result.Records,
		func(userId string) string {
			// User ID uniquely identifies the source for next-page token resolution.
			// The context for this NextPageToken is represented solely by userId.
			return userId
		},
		func(userId string, row *common.ReadResultRow) {
			// Enhance fields to indicate what user this row is associated with.
			row.Fields[core.FieldUserId] = userId
		},
	)
}

func (c *Connector) listAllUsers(ctx context.Context) ([]string, error) {
	url, err := c.GetReadURL(core.ObjectNameUser)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", core.DefaultPageSize)

	userIds := make([]string, 0)

	for {
		res, err := c.JSONHTTPClient().Get(ctx, url.String())
		if err != nil {
			return nil, err
		}

		usersResp, err := common.UnmarshalJSON[usersListResponse](res)
		if err != nil {
			return nil, err
		}

		for _, user := range usersResp.Users {
			userIds = append(userIds, user.Id)
		}

		if usersResp.NextPageToken == "" || len(userIds) == 0 {
			return userIds, nil // => Desired return with collected ids.
		}

		// Prepare next page URL by advancing the page token.
		url.WithQueryParam("next_page_token", usersResp.NextPageToken)
	}
}

type usersListResponse struct {
	NextPageToken string `json:"next_page_token"`
	Users         []struct {
		Id string `json:"id"`
	} `json:"users"`
}

// nolint:nestif
func addTimeFilterParams(config common.ReadParams, url *urlbuilder.URL) {
	// Zoom API uses "from" and "to" query parameters with date format YYYY-MM-DD.
	// Reference: https://developers.zoom.us/docs/api/meetings/#tag/archiving/get/archive_files
	if incrementalObjects.Has(config.ObjectName) {
		// Some objects require mandatory from/to parameters
		if mandatoryDateObjects.Has(config.ObjectName) {
			// Set default date range to last 29 days if not provided
			startDate := time.Now().AddDate(0, 0, -29)
			endDate := time.Now()

			if !config.Since.IsZero() {
				startDate = config.Since
			}

			if !config.Until.IsZero() {
				endDate = config.Until
			}

			url.WithQueryParam("from", startDate.UTC().Format(DateFormat))
			url.WithQueryParam("to", endDate.UTC().Format(DateFormat))
		} else {
			if !config.Since.IsZero() {
				url.WithQueryParam("from", config.Since.UTC().Format(DateFormat))
			}

			if !config.Until.IsZero() {
				url.WithQueryParam("to", config.Until.UTC().Format(DateFormat))
			}
		}
	}
}
