package slack

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

// GetPostAuthInfo retrieves the instance teamId using HTTP GET API Call.
func (c *Connector) GetPostAuthInfo(ctx context.Context) (*common.PostAuthInfo, error) {
	logging.With(ctx, "provider", "slack", "step", "get_post_auth_info")

	teamID, err := c.retrieveInstanceTeamID(ctx)
	if err != nil {
		return nil, err
	}

	c.teamId = teamID

	catalogVars := map[string]string{
		"teamId": teamID,
	}

	return &common.PostAuthInfo{
		CatalogVars: &catalogVars,
	}, nil
}

func (c *Connector) retrieveInstanceTeamID(ctx context.Context) (string, error) {
	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, "auth.test")
	if err != nil {
		return "", fmt.Errorf("failed to build Slack URL: %w", err)
	}

	resp, err := c.JSONHTTPClient().Post(ctx, url.String(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to execute Get API call: %w", err)
	}

	return parseTeamIDResponse(resp)
}

type authTestResponse struct {
	TeamID string `json:"team_id"`
}

func parseTeamIDResponse(resp *common.JSONHTTPResponse) (string, error) {
	body, ok := resp.Body()
	if !ok {
		return "", common.ErrEmptyJSONHTTPResponse
	}

	authTestResp, err := jsonquery.ParseNode[authTestResponse](body)
	if err != nil {
		return "", fmt.Errorf("failed to parse team_id response: %w", err)
	}

	if authTestResp.TeamID == "" {
		return "", errors.New("failed to obtain team_id from auth.test API") // nolint: err113
	}

	return authTestResp.TeamID, nil
}

type AuthMetadataVars struct {
	TeamId string
}

// NewAuthMetadataVars parses map into the model.
func NewAuthMetadataVars(dictionary map[string]string) *AuthMetadataVars {
	return &AuthMetadataVars{
		TeamId: dictionary["teamId"],
	}
}

func (v AuthMetadataVars) AsMap() *map[string]string {
	return &map[string]string{
		"teamId": v.TeamId,
	}
}
