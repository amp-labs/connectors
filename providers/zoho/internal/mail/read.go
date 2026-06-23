package mail

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

func (a *Adapter) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	obj, err := lookupObject(config.ObjectName)
	if err != nil {
		return nil, err
	}

	url, err := a.buildReadURL(config, obj)
	if err != nil {
		return nil, err
	}

	resp, err := a.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		extractRecordsFromKeyPath(obj.recordsPath),
		a.makeNextRecordsURL(url, obj),
		readhelper.MakeGetMarshaledDataWithId(readhelper.NewIdField(obj.recordIdKey)),
		config.Fields,
	)
}

func (a *Adapter) buildReadURL(config common.ReadParams, obj objectDescriptor) (*urlbuilder.URL, error) {
	if config.NextPage != "" {
		return urlbuilder.New(string(config.NextPage))
	}

	url, err := a.objectURL(obj)
	if err != nil {
		return nil, err
	}

	if obj.pagination != nil {
		url.WithQueryParam(obj.pagination.offsetParam, strconv.Itoa(obj.pagination.startOffset))
		url.WithQueryParam("limit", strconv.Itoa(pageSize(config, obj.pagination)))
	}

	return url, nil
}

// pageSize return pageSize from config if it's set and valid;
// otherwise it returns the pagination's max limit.
func pageSize(config common.ReadParams, p *pagination) int {
	if config.PageSize > 0 && config.PageSize <= p.maxLimit {
		return config.PageSize
	}

	return p.maxLimit
}
