package auphan

import (
	"context"
	"encoding/json"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/auphan"
	"github.com/amp-labs/connectors/test/utils"
)

// creds is ./auphan-creds.json. The store ID sits under substitutions.workspace, which the
// credscanning reader does not look at (it wants metadata.workspace), so the file is read directly.
type creds struct {
	APIKey        string `json:"apiKey"`
	Substitutions struct {
		Workspace string `json:"workspace"`
	} `json:"substitutions"`
}

func GetAuphanConnector(ctx context.Context) *auphan.Connector {
	data, err := os.ReadFile(credscanning.LoadPath(providers.Auphan))
	if err != nil {
		utils.Fail("error reading Auphan creds file", "error", err)
	}

	var file creds
	if err := json.Unmarshal(data, &file); err != nil {
		utils.Fail("error parsing Auphan creds file", "error", err)
	}

	if file.APIKey == "" || file.Substitutions.Workspace == "" {
		utils.Fail("Auphan creds file needs apiKey and substitutions.workspace")
	}

	info, err := providers.ReadInfo(providers.Auphan)
	if err != nil {
		utils.Fail("error reading provider info", "error", err)
	}

	headerName, headerValue, err := info.GetApiKeyHeader(file.APIKey)
	if err != nil {
		utils.Fail("error getting API key header", "error", err)
	}

	client, err := common.NewApiKeyHeaderAuthHTTPClient(ctx, headerName, headerValue)
	if err != nil {
		utils.Fail("error creating API key client", "error", err)
	}

	conn, err := auphan.NewConnector(common.ConnectorParams{
		AuthenticatedClient: client,
		Workspace:           file.Substitutions.Workspace,
	})
	if err != nil {
		utils.Fail("error creating Auphan connector", "error", err)
	}

	return conn
}
