package exensify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	err := config.ValidateParams(true)
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByRead.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	reqURL, err := urlbuilder.New(c.ProviderInfo().BaseURL)
	if err != nil {
		return nil, err
	}

	body, err := buildReadBody(config.ObjectName)
	if err != nil {
		return nil, err
	}

	newForm := url.Values{}

	newForm.Set("requestJobDescription", body)

	ecoded := newForm.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL.String(), bytes.NewBufferString(ecoded))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.HTTPClient().Client.Do(req)
	if err != nil {
		logging.Logger(ctx).Error("failed to get metadata", "object", objectName, "err", err.Error())

		return nil, fmt.Errorf("failed to get metadata for object %s: %w", objectName, err)
	}

	return common.ParseResult(
		rsp,
		common.ExtractOptionalRecordsFromPath("data"),
		makeNextRecordsURL(),
		common.GetMarshaledData,
		config.Fields,
	)
}

func (c *Connector) buildURL(config common.ReadParams) (*urlbuilder.URL, error) {
	if len(config.NextPage) != 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getAPIURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	fieldsStr := strings.Join(config.Fields.List(), ",")

	url.WithQueryParam("opt_fields", fieldsStr)

	if supportLimitAndOffset.Has(config.ObjectName) {
		url.WithQueryParam("limit", strconv.Itoa(DefaultPageSize))
	}

	return url, err
}
