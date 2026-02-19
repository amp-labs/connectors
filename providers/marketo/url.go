package marketo

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

const ( //nolint:gochecknoglobals
	// API path components.
	activities      = "activities"
	leads           = "leads"
	pagingURLSuffix = "activities/pagingtoken"
	assetsPrefix    = "rest/asset/v1"
	leadsPrefix     = "rest/v1"

	// URL parameter keys.
	idFilter             = "id"
	sinceQuery           = "sinceDatetime"
	nextPageQuery        = "nextPageToken"
	activityTypeIDs      = "activityTypeIds"
	filterValuesQuery    = "filterValues"
	filterTypeQuery      = "filterType"
	earliestUpdatedQuery = "earliestUpdatedAt"
	latestUpdatedAtQuery = "latestUpdatedAt"
	fields               = "fields"

	newLeadActivityType = "12"
	maxReturn           = 200
)

type pagingTokenResponse struct {
	NextPageToken string `json:"nextPageToken"`
	Success       bool   `json:"success"`
}

var assetsObjects = datautils.NewSet( //nolint:gochecknoglobals
	"channels",
	"emailTemplates",
	"emails",
	"files",
	"folders",
	"form/fields",
	"forms",
	"landingPages",
	"redirectRules",
	"landingPageDomains",
	"landingPageTemplates",
	"programs",
	"segmentation",
	"smartLists",
	"smartCampaigns",
	"snippets",
	"staticLists",
	"tagTypes",
)

var leadsObjects = datautils.NewSet( //nolint: gochecknoglobals
	"activities",
	"ctivities/deletedleads",
	"activities/external/types",
	"activities/leadchanges",
	"activities/pagingtoken",
	"activities/types",
	"campaigns",
	"companies",
	"customobjects",
	"leads",
	"leads/schema/fields",
	"leads/partitions",
	"leads/push",
	"leads/submitForm",
	"leads/delete",
	"namedAccountLists",
	"namedAccountLists/delete",
	"namedaccounts",
	"namedaccounts/delete",
	"opportunities",
	"opportunities/roles",
	"salespersons",
	"lists",
	"stats/errors",
	"stats/usage",
	"stats/usage/last7days",
)

func constructAPIPrefix(objectName string) string {
	switch {
	case assetsObjects.Has(objectName):
		return assetsPrefix
	case leadsObjects.Has(objectName):
		return leadsPrefix
	default:
		return leadsPrefix
	}
}

func (c *Connector) constructReadURL(ctx context.Context, params common.ReadParams) (*urlbuilder.URL, string, error) {
	url, err := c.getAPIURL(params.ObjectName)
	if err != nil {
		return nil, "", err
	}

	nextPageToken, err := c.constructURLQueries(ctx, url, params)
	if err != nil {
		return nil, "", err
	}

	if err := c.handleActivitiesAPI(ctx, url, params); err != nil {
		return nil, "", err
	}

	// The only objects in Assets API supporting this are: Emails, Programs, SmartCampaigns,SmartLists
	// https://developer.adobe.com/marketo-apis/api/mapi/#operation/getProgramMembershipUsingGET
	if assetsObjects.Has(params.ObjectName) {
		if !params.Since.IsZero() {
			fmtTime := params.Since.Format(time.RFC3339)
			url.WithQueryParam(earliestUpdatedQuery, fmtTime)
		}

		if !params.Until.IsZero() {
			fmtTime := params.Until.Format(time.RFC3339)
			url.WithQueryParam(latestUpdatedAtQuery, fmtTime)
		}
	}

	return url, nextPageToken, nil
}

func (c *Connector) constructURLQueries(
	ctx context.Context, url *urlbuilder.URL, params common.ReadParams,
) (string, error) {
	// If were'reading leads, we don't handle this scenario here, this is handled in the handleLeadsAPI
	if params.ObjectName == leads {
		return c.handleLeadsAPI(ctx, url, params)
	}

	// TODO:  Handle other id filtering objects like opportunities, companies

	return "", nil
}

func (c *Connector) handleActivitiesAPI(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	// Check if this is an initial request to the Marketo Activities API.
	// For the first call (no NextPage token) with a Since timestamp,
	// fetch a paging token to ensure pagination starts from the correct time.
	// Then, append the token to the URL for subsequent pagination.
	if params.ObjectName == activities {
		if params.Filter == "" {
			return ErrFilterInvalid
		}

		url.WithQueryParam(activityTypeIDs, params.Filter)

		if err := c.addActivityNextParam(ctx, url, params); err != nil {
			return err
		}
	}

	return nil
}

func (c *Connector) generateLeadIDs(ctx context.Context, params common.ReadParams) ([]string, string, error) {
	var (
		token         string
		err           error
		nextPageToken string
	)

	if params.NextPage == "" {
		if params.Since.IsZero() {
			// use 1970 when we are doing backfill for Leads
			// similar to what we did with activities.
			params.Since = time.Unix(0, 0).UTC()
		}

		token, err = c.getPagingToken(ctx, params.Since)
		if err != nil {
			return nil, nextPageToken, fmt.Errorf("failed to retrieve paging token when reading leads: %w", err)
		}
	} else {
		token = params.NextPage.String()
	}

	resp, err := c.getLeadActivities(ctx, token)
	if err != nil {
		return nil, nextPageToken, fmt.Errorf("failed to retrieve activities when reading leads: %w", err)
	}

	if resp.MoreResult {
		nextPageToken = resp.NextPageToken
	}

	if len(resp.Result) == 0 {
		return nil, nextPageToken, ErrZeroRecords
	}

	ids := make([]string, 0, batchSize)

	for _, act := range resp.Result {
		ids = append(ids, strconv.Itoa(act.LeadID))
	}

	return ids, nextPageToken, nil
}

func (c *Connector) getLeadActivities(ctx context.Context, token string) (*readResponse, error) {
	url, err := c.getAPIURL(activities)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(activityTypeIDs, newLeadActivityType)
	url.WithQueryParam(nextPageQuery, token)

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	rsp, err := common.UnmarshalJSON[readResponse](res)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func (c *Connector) getPagingToken(ctx context.Context, since time.Time) (string, error) {
	pagingTokenURL, err := c.getAPIURL(pagingURLSuffix)
	if err != nil {
		return "", err
	}

	pagingTokenURL.WithQueryParam(sinceQuery, since.Format(time.RFC3339))

	resp, err := c.Client.Get(ctx, pagingTokenURL.String())
	if err != nil {
		return "", err
	}

	pagingResponse, err := common.UnmarshalJSON[pagingTokenResponse](resp)
	if err != nil {
		return "", err
	}

	return pagingResponse.NextPageToken, nil
}

func (c *Connector) handleLeadsAPI(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) (string, error) {
	ids, nextPage, err := c.generateLeadIDs(ctx, params)
	if err != nil {
		return nextPage, err
	}

	url.WithQueryParam(filterValuesQuery, strings.Join(ids, ","))
	url.WithQueryParam(filterTypeQuery, idFilter)
	url.WithQueryParam(fields, strings.Join(params.Fields.List(), ","))

	return nextPage, nil
}

func (c *Connector) addActivityNextParam(ctx context.Context, url *urlbuilder.URL, params common.ReadParams) error {
	if params.NextPage != "" {
		url.WithQueryParam(nextPageQuery, params.NextPage.String())

		return nil
	}

	// Manually setting the since timestamp to `1970-01-01` for retrieving
	// all lead activities in the instance.
	if params.Since.IsZero() {
		params.Since = time.Unix(0, 0).UTC()
	}

	// Get initial paging token for first request
	token, err := c.getPagingToken(ctx, params.Since)
	if err != nil {
		return err
	}

	url.WithQueryParam(nextPageQuery, token)

	return nil
}

func (c *Connector) constructMetadataURL(objectName string) (*urlbuilder.URL, error) {
	path, ok := hasMetadataResource(objectName)
	if !ok {
		return c.getAPIURL(objectName)
	}

	return urlbuilder.New(c.BaseURL, path)
}
