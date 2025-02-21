package chilipiper

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

/*
Sample Write Response
{
    "id": "4edf8761-e5ee-48b2-81c8-c5e4849481fc",
    "workspaceId": "cad33722-df27-4691-bc11-1f2c89c1dd31",
    "name": "Dev Team",
    "members": [],
    "metadata": {
        "createdAt": "2025-01-24T09:37:47.321631Z",
        "createdBy": "user/67929af0725ce43853fd2b8c",
        "updatedAt": "2025-01-30T08:10:54.357774Z",
        "updatedBy": "apikey/api_b830c0b6__",
        "teamMembersMetadata": {
            "addedAt": {}
        },
        "revision": 3
    }
}
*/

type response struct {
	Id string `json:"id"`
	// The rest changes depending on the action.
}

func (conn *Connector) Write(ctx context.Context, config common.WriteParams) (*common.WriteResult, error) {
	if err := config.ValidateParams(); err != nil {
		return nil, err
	}

	writeURL, err := conn.buildWriteURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	if len(config.RecordId) > 0 {
		writeURL.AddPath(config.RecordId)
	}

	resp, err := conn.Client.Post(ctx, writeURL.String(), config.RecordData)
	if err != nil {
		return nil, err
	}

	return constructWriteResponse(resp), nil
}

func constructWriteResponse(resp *common.JSONHTTPResponse) *common.WriteResult {
	res, err := common.UnmarshalJSON[response](resp)
	if err != nil {
		return &common.WriteResult{
			Success: true,
		}
	}

	return &common.WriteResult{
		RecordId: res.Id,
		Success:  true,
	}
}
