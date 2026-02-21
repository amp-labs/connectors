package expensify

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/scanning/credscanning"
	"github.com/amp-labs/connectors/providers"
	"github.com/amp-labs/connectors/providers/expensify"
	"github.com/amp-labs/connectors/test/utils"
)

func GetConnector(ctx context.Context) *expensify.Connector {
	filePath := credscanning.LoadPath(providers.Expensify)
	reader := utils.MustCreateProvCredJSON(filePath, false, credscanning.Fields.ApiKey, credscanning.Fields.Secret)

	apiKey := reader.Get(credscanning.Fields.ApiKey)
	apiSecret := reader.Get(credscanning.Fields.Secret)

	client, err := common.NewCustomAuthHTTPClient(ctx, common.WithCustomBodyModifier(bodyModifier(apiKey, apiSecret)))
	if err != nil {
		utils.Fail("error creating client", "error", err)
	}

	conn, err := expensify.NewConnector(
		common.ConnectorParams{AuthenticatedClient: client},
	)

	if err != nil {
		utils.Fail("error creating connector", "error", err)
	}

	return conn
}

func bodyModifier(partnerUserID, partnerUserSecret string) func(req *http.Request) error {
	return func(req *http.Request) error {
		reqBody, err := io.ReadAll(req.Body)
		if err != nil {
			return err
		}

		req.Body.Close()

		formValues, err := url.ParseQuery(string(reqBody))
		if err != nil {
			return err
		}

		jobDescJSON := formValues.Get("requestJobDescription")

		var jobDesc map[string]any

		if jobDescJSON != "" {
			err = json.Unmarshal([]byte(jobDescJSON), &jobDesc)
			if err != nil {
				return err
			}
		} else {
			jobDesc = map[string]any{}
		}

		jobDesc["credentials"] = map[string]any{
			"partnerUserID":     partnerUserID,
			"partnerUserSecret": partnerUserSecret,
		}

		modifiedJobDescJSON, err := json.Marshal(jobDesc)
		if err != nil {
			return err
		}

		newForm := url.Values{}
		newForm.Set("requestJobDescription", string(modifiedJobDescJSON))

		encoded := newForm.Encode()

		req.Body = io.NopCloser(bytes.NewReader([]byte(encoded)))
		req.ContentLength = int64(len(encoded))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		return nil
	}
}
