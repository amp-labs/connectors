package instantly

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Write method allows to
// * create campaigns
// * create/update email-accounts
// * create client
// Documentation links can be found within constructURLPath function.
func (c *Connector) Write(
	ctx context.Context, config common.WriteParams,
) (*common.WriteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Object name will be used to complete further URL construction below.
	url, err := c.getURL()
	if err != nil {
		return nil, err
	}

	var write common.WriteMethod
	if len(config.RecordId) == 0 {
		write = c.Client.Post

		constructURLPathCreate(config, url) // nolint:wsl
	} else {
		write = c.Client.Patch

		constructURLPathUpdate(config, url)
	}

	res, err := write(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	body, ok := res.Body()
	if !ok {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	recordIdNodePath := recordIdPaths[config.ObjectName]

	// write response was with payload
	return constructWriteResult(body, recordIdNodePath)
}

var recordIdPaths = map[string]*string{ // nolint:gochecknoglobals
	objectNameLeads:            nil, // ID is not returned for Leads.
	objectNameBlocklistEntries: handy.Pointers.Str("blocklist_id"),
	objectNameUniboxReplies:    handy.Pointers.Str("message_id"),
	objectNameTags:             handy.Pointers.Str("id"),
}

func constructURLPathCreate(config common.WriteParams, url *urlbuilder.URL) {
	switch config.ObjectName {
	case objectNameLeads:
		// Add lead to campaign.
		// https://developer.instantly.ai/campaign/add-leads-to-a-campaign
		url.AddPath("lead/add")
	case objectNameBlocklistEntries:
		// Add blocklist entry.
		// https://developer.instantly.ai/blocklist/add-entries-to-blocklist
		url.AddPath("blocklist/add/entries")
	case objectNameUniboxReplies:
		// Create message - unibox reply.
		// https://developer.instantly.ai/unibox/send-reply
		url.AddPath("unibox/emails/reply")
	case objectNameTags:
		// Create tag.
		// https://developer.instantly.ai/tags/create-a-new-tag
		url.AddPath("custom-tag")
	}
}

func constructURLPathUpdate(config common.WriteParams, url *urlbuilder.URL) {
	if config.ObjectName == objectNameTags {
		// Update account.
		// https://developer.instantly.ai/tags/update-tag
		url.AddPath("custom-tag", config.RecordId)
	}
}

func constructWriteResult(body *ajson.Node, recordIdLocation *string) (*common.WriteResult, error) {
	if recordIdLocation == nil {
		// ID is not present inside response. Therefore, empty.
		return &common.WriteResult{
			Success:  true,
			RecordId: "",
			Errors:   nil,
			Data:     nil,
		}, nil
	}

	// ID is integer that is always stored under different field name.
	recordID, err := jsonquery.New(body).Str(*recordIdLocation, false)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: *recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
