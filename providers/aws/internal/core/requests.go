package core

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
)

// nolint:lll
const (
	// Mime https://docs.aws.amazon.com/AWSSimpleQueueService/latest/SQSDeveloperGuide/sqs-making-api-requests-json.html#sqs-api-constructing-endpoints-json
	Mime = "application/x-amz-json-1.1"
)

// ReadPayload
// nolint:tagliatelle
type ReadPayload struct {
	MaxResults *int    `json:"MaxResults,omitempty"`
	NextToken  *string `json:"NextToken,omitempty"`
}

func NewReadPayload(params common.ReadParams) *ReadPayload {
	var nextToken *string
	if len(params.NextPage) != 0 {
		nextToken = goutils.Pointer(params.NextPage.String())
	}

	return &ReadPayload{
		NextToken: nextToken,
	}
}

func BuildRequest(
	ctx context.Context, baseURL, serviceDomain, serviceSigningName string, command Command, payload any,
) (*http.Request, error) {
	reader, err := getReader(payload)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, common.AWSServiceContextKey, serviceSigningName)
	baseURL = strings.Replace(baseURL, "<<SERVICE_DOMAIN>>", serviceDomain, 1)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL, reader)
	if err != nil {
		return nil, err
	}

	// Required headers
	req.Header.Set("Content-Type", Mime)
	req.Header.Set("X-Amz-Target", command.String())

	return req, nil
}

func getReader(payload any) (*bytes.Reader, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(data), nil
}
