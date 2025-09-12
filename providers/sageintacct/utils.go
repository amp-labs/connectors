package sageintacct

import (
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
)

func buildReadBody(module common.ModuleID, params common.ReadParams) (map[string]interface{}, error) {
	path, err := metadata.Schemas.LookupURLPath(module, params.ObjectName)
	if err != nil {
		return nil, err
	}

	fullObjectName := strings.Split(path, "/objects/")[1]

	objectMetadata, err := metadata.Schemas.Select(module, []string{params.ObjectName})
	if err != nil {
		return nil, err
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
			return nil, err
		}

		payload[pageParam] = pageNum
	}

	return payload, nil
}
