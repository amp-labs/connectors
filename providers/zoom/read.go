package zoom

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
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
	// The page is saved with the read's responses, so a retry fetches the same meetings' associations.
	rsp, err := c.Base.GetWithCheckpoint(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return c.parseRecords(ctx, rsp, objectName, selectedFields, url, associatedObjects)
}

func (c *Connector) parseRecords(ctx context.Context,
	rsp *common.JSONHTTPResponse,
	objectName string,
	selectedFields datautils.StringSet,
	url *urlbuilder.URL,
	associatedObjects []string,
) (*common.ReadResult, error) {
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

	urls := make([]*urlbuilder.URL, len(userIds))

	for index, userId := range userIds {
		url, err := c.GetReadUrlForUser(params.ObjectName, userId)
		if err != nil {
			return nil, err
		}

		url.WithQueryParam("page_size", core.DefaultPageSize)
		addTimeFilterParams(params, url)

		urls[index] = url
	}

	return c.readUsersPages(ctx, params, userIds, urls)
}

func (c *Connector) readNextPageForUsers(ctx context.Context,
	params common.ReadParams,
	aggregateToken readhelper.AggregateNextPage[string],
) (*common.ReadResult, error) {
	userIds := make([]string, len(aggregateToken))
	urls := make([]*urlbuilder.URL, len(aggregateToken))

	for index, token := range aggregateToken {
		url, err := urlbuilder.New(token.Value.String())
		if err != nil {
			return nil, err
		}

		userIds[index] = token.Context
		urls[index] = url
	}

	return c.readUsersPages(ctx, params, userIds, urls)
}

// readUsersPages reads one page for every user as a request batch, and aggregates the pages into
// one result whose next page token holds every user's next page.
func (c *Connector) readUsersPages(ctx context.Context,
	params common.ReadParams,
	userIds []string,
	urls []*urlbuilder.URL,
) (*common.ReadResult, error) {
	requests := make([]common.BatchRequest, len(urls))
	for index, url := range urls {
		requests[index] = common.BatchRequest{Method: http.MethodGet, URL: url.String()}
	}

	responses, err := c.JSONHTTPClient().DoBatch(ctx, requests)
	if err != nil {
		return nil, err
	}

	results := make(map[string]common.ReadResult, len(userIds))

	for index, rsp := range responses {
		if rsp.Err != nil {
			return nil, fmt.Errorf("reading %s for user %s: %w", params.ObjectName, userIds[index], rsp.Err)
		}

		result, err := c.parseRecords(ctx, rsp.Response, params.ObjectName, params.Fields, urls[index],
			params.AssociatedObjects)
		if err != nil {
			return nil, err
		}

		results[userIds[index]] = *result
	}

	return readhelper.AggregateReadResults(
		results,
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

// listAllUsers pages through the account's users. Each page is saved with the read's responses, so
// a read that is tried again sends the same per-user requests and reuses their saved responses.
func (c *Connector) listAllUsers(ctx context.Context) ([]string, error) {
	url, err := c.GetReadURL(core.ObjectNameUser)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("page_size", core.DefaultPageSize)

	userIds := make([]string, 0)

	for {
		res, err := c.GetWithCheckpoint(ctx, url.String())
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
