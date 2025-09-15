package sageintacct

import (
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/providers/sageintacct/metadata"
)

func getFullObjectName(module common.ModuleID, object string) (string, error) {
	path, err := metadata.Schemas.LookupURLPath(module, object)
	if err != nil {
		return "", err
	}

	fullObjectName := strings.Split(path, "/objects/")[1]

	return fullObjectName, nil
}

func mapValuesFromEnum(fieldDef SageIntacctFieldDef) []common.FieldValue {
	values := []common.FieldValue{}

	if len(fieldDef.Enum) > 0 {
		for _, v := range fieldDef.Enum {
			values = append(values, common.FieldValue{
				DisplayValue: naming.CapitalizeFirstLetter(v),
				Value:        v,
			})
		}
	}

	return values
}
