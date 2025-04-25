package awsic

import (
	"bytes"
	"context"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

func (c *Connector) buildReadRequest(ctx context.Context, params common.ReadParams) (*http.Request, error) {
	command, ok := readCommands[params.ObjectName]
	if !ok {
		return nil, ErrCommandNotFound
	}

	ctx = context.WithValue(ctx, common.AWSServiceContextKey, command.ServiceDomain)

	// TODO base URL should be altered for each request. Need to think about proxy. For now this is a hack.
	baseURL := strings.Replace(c.ProviderInfo().BaseURL, "AWS_SERVICE_PLACEHOLDER", command.ServiceDomain, -1)

	reader := bytes.NewReader(command.PayloadBuilder(map[string]string{
		"IdentityStoreID": c.identityStoreID,
		"InstanceArn":     c.instanceArn,
	}))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, reader)
	if err != nil {
		return nil, err
	}

	// Required headers
	req.Header.Set("Content-Type", "application/x-amz-json-1.1")
	req.Header.Set("X-Amz-Target", command.ServiceName+"."+command.Command)

	return req, nil
}

func (c *Connector) parseReadResponse(
	ctx context.Context,
	params common.ReadParams,
	request *http.Request,
	response *common.JSONHTTPResponse,
) (*common.ReadResult, error) {
	// TODO map object name to the location of records
	recordsLocation := params.ObjectName

	return common.ParseResult(
		response,
		common.ExtractOptionalRecordsFromPath(recordsLocation),
		func(node *ajson.Node) (string, error) {
			return jsonquery.New(node).StrWithDefault("NextToken", "") // TODO validate true for all
		},
		common.GetMarshaledData,
		params.Fields,
	)
}
