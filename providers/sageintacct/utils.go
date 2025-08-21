package sageintacct

import (
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
)

func buildURL(
	module common.ModuleID,
	params common.ReadParams,
	baseURL string,
) (*urlbuilder.URL, map[string]interface{}, error) {
	path, err := metadata.Schemas.LookupURLPath(module, params.ObjectName)
	if err != nil {
		return nil, nil, err
	}

	fullObjectName := strings.Split(path, "/objects/")[1]

	objectMetadata, err := metadata.Schemas.Select(module, []string{params.ObjectName})
	if err != nil {
		return nil, nil, err
	}

	var fieldNames []string

	for _, objectFields := range objectMetadata.Result {
		for fieldName := range objectFields.Fields {
			fieldNames = append(fieldNames, fieldName)
		}
	}

	payload := map[string]any{
		"object":      fullObjectName,
		"fields":      fieldNames,
		pageSizeParam: defaultPageSize,
		pageParam:     1,
	}

	if params.NextPage != "" {
		pageNum, err := strconv.Atoi(string(params.NextPage))
		if err != nil {
			return nil, nil, err
		}

		payload[pageParam] = pageNum
	}

	url, err := urlbuilder.New(baseURL, apiVersion, "services/core/query")
	if err != nil {
		return nil, nil, err
	}

	return url, payload, nil
}
