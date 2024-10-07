package smartlead

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

// Write method allows to
// * create campaigns
// * create/update email-accounts
// * create client
// Documentation links can be found within completeURLPath function.
func (c *Connector) Write(
	ctx context.Context, config common.WriteParams,
) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) == 0 {
		constructURLPathCreate(config, url)
	} else {
		constructURLPathUpdate(config, url)
	}

	res, err := c.JSON.Post(ctx, url.String(), config.RecordData)
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

var recordIdPaths = map[string]string{ // nolint:gochecknoglobals
	objectNameCampaign:     "id",
	objectNameEmailAccount: "emailAccountId",
	objectNameClient:       "clientId",
}

func constructURLPathCreate(config common.WriteParams, url *urlbuilder.URL) {
	switch config.ObjectName {
	case objectNameCampaign:
		// Create campaign.
		// https://api.smartlead.ai/reference/create-campaign
		url.AddPath("create")
	case objectNameEmailAccount:
		// Create account.
		// https://api.smartlead.ai/reference/create-an-email-account
		url.AddPath("save")
	case objectNameClient:
		// Add new client to the system.
		// https://api.smartlead.ai/reference/add-client-to-system-whitelabel-or-not
		url.AddPath("save")
	}
}

func constructURLPathUpdate(config common.WriteParams, url *urlbuilder.URL) {
	if config.ObjectName == objectNameEmailAccount {
		// Update account.
		// https://api.smartlead.ai/reference/update-email-account
		url.AddPath(config.RecordId)
	}
}

func constructWriteResult(body *ajson.Node, recordIdLocation string) (*common.WriteResult, error) {
	// ID is integer that is always stored under different field name.
	rawID, err := jsonquery.New(body).Integer(recordIdLocation, true)
	if err != nil {
		return nil, err
	}

	recordID := ""
	if rawID != nil {
		// optional
		recordID = strconv.FormatInt(*rawID, 10)
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Errors:   nil,
		Data:     nil,
	}, nil
}
