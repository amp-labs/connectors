package smartlead

import (
	"context"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/handy"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/spyzhov/ajson"
)

const (
	objectNameCampaign = "campaigns"
	objectEmailAccount = "email-accounts"
	objectNameClient   = "client"
)

var supportedObjectsByWrite = handy.NewSet([]string{ //nolint:gochecknoglobals
	objectNameCampaign,
	objectEmailAccount,
	objectNameClient,
})

// Write method allows to
// * create campaigns
// * create/update email-accounts
// * create client
// Documentation links can be found within completeURLPath function.
func (c *Connector) Write(
	ctx context.Context, config common.WriteParams,
) (*common.WriteResult, error) {
	if len(config.ObjectName) == 0 {
		return nil, common.ErrMissingObjects
	}

	url, err := c.getURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if !supportedObjectsByWrite.Has(config.ObjectName) {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Finish URL format according to ObjectName.
	recordIdLocation, err := completeURLPath(config, url)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Post(ctx, url.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	if res == nil || res.Body == nil {
		// it is unlikely to have no payload
		return &common.WriteResult{
			Success: true,
		}, nil
	}

	// write response was with payload
	return constructWriteResult(res.Body, recordIdLocation)
}

// Maps common.WriteParams to URL format for that object.
// In addition, returns field name where ID will be stored on successful response.
func completeURLPath(config common.WriteParams, url *urlbuilder.URL) (string, error) {
	if len(config.RecordId) == 0 {
		return completeURLPathCreate(config, url)
	}

	return completeURLPathUpdate(config, url)
}

func completeURLPathCreate(config common.WriteParams, url *urlbuilder.URL) (string, error) {
	switch config.ObjectName {
	case objectNameCampaign:
		// Create campaign.
		// https://api.smartlead.ai/reference/create-campaign
		url.AddPath("create")

		return "id", nil
	case objectEmailAccount:
		// Create account.
		// https://api.smartlead.ai/reference/create-an-email-account
		url.AddPath("save")

		// ID is located under the same field for create/update operations.
		return "emailAccountId", nil
	case objectNameClient:
		// Add new client to the system.
		// https://api.smartlead.ai/reference/add-client-to-system-whitelabel-or-not
		url.AddPath("save")

		return "clientId", nil
	default:
		return "", common.ErrOperationNotSupportedForObject
	}
}

func completeURLPathUpdate(config common.WriteParams, url *urlbuilder.URL) (string, error) {
	switch config.ObjectName {
	case objectEmailAccount:
		// Update account.
		// https://api.smartlead.ai/reference/update-email-account
		url.AddPath(config.RecordId)

		// ID is located under the same field for create/update operations.
		return "emailAccountId", nil
	default:
		// The code should be unreachable if checks before this function call were correct.
		return "", common.ErrOperationNotSupportedForObject
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
