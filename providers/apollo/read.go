package apollo

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

const (
	accounts  = "accounts"
	contacts  = "contacts"
	updatedAt = "updated_at"
	createdAt = "created_at"
)

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and provided read parameters.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) { //nolint: cyclop,funlen,lll
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.getAPIURL(config.ObjectName, readOp)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam(perPage, pageSize)

	// If NextPage is set, then we're reading the next page of results.
	if len(config.NextPage) > 0 {
		url.WithQueryParam(pageQuery, config.NextPage.String())
	} else {
		// If we're reading this for the first time, we make a call to retrieve
		// custom fields, add them and their labels in the connector instance field customFields.
		if err := c.retrieveCustomFields(ctx, config); err != nil {
			return nil, err
		}
	}

	var res *common.JSONHTTPResponse

	// If object uses POST searching, then we have to use the search endpoint, POST method.
	// The Search endpoint has a 50K record limit.
	switch {
	case in(config.ObjectName, readingSearchObjectsGET, readingListObjects):
		res, err = c.Client.Get(ctx, url.String())
		if err != nil {
			return nil, err
		}
	case in(config.ObjectName, readingSearchObjectsPOST):
		if config.ObjectName == accounts {
			url.WithQueryParam("sort_by_field", "account_created_at")
		}

		if config.ObjectName == contacts {
			url.WithQueryParam("sort_by_field", "contact_updated_at")
		}

		return c.Search(ctx, config, url)
	default:
		return nil, common.ErrObjectNotSupported
	}

	return common.ParseResult(res,
		recordsWrapperFunc(config.ObjectName),
		getNextRecords,
		c.apolloMarshaledData(config.ObjectName),
		config.Fields,
	)
}

func (c *Connector) retrieveCustomFields(ctx context.Context, config common.ReadParams) error {
	// we need check if there are custom fields
	fldLabels := make(map[string]string)

	if usesFieldsResource.Has(config.ObjectName) {
		metadata, err := c.retrieveFields(ctx, config.ObjectName)
		if err != nil {
			return err
		}

		for id, fld := range metadata.Fields {
			if *fld.IsCustom {
				fldName := id
				label := fld.DisplayName

				fldLabels[fldName] = label
			}
		}

		c.customFields = fldLabels
	}

	return nil
}
