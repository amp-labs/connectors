package sageintacct

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
)

func buildURL(module common.ModuleID, params common.ReadParams, baseURL string) (*urlbuilder.URL, map[string]interface{}, error) {

	path, err := metadata.Schemas.LookupURLPath(module, params.ObjectName)
	if err != nil {
		return nil, nil, err
	}

	fullObjectName := strings.Split(path, "/objects/")[1]

	payload := map[string]any{
		"object": fullObjectName,
		"fields": []string{
			"id",
			"href",
		},
	}

	url, err := urlbuilder.New(baseURL, apiVersion, "services/core/query")
	if err != nil {
		return nil, nil, err
	}

	return url, payload, nil
}
