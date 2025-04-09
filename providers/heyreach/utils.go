package heyreach

import (
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	DefaultPageSize = 100
)

func matchReadObjectNameToEndpointPath(objectName string) (urlPath string, err error) {
	switch objectName {
	case objectNameCampaign:
		// https://documenter.getpostman.com/view/23808049/2sA2xb5F75#4aaf461e-54c4-4447-beb6-eb5ccca53fb3
		return "campaign/GetAll", nil
	case objectNameLiAccount:
		// https://documenter.getpostman.com/view/23808049/2sA2xb5F75#8a4d2924-d46c-443f-a4df-65216988a091
		return "li_account/GetAll", nil
	case objectNameList:
		// https://documenter.getpostman.com/view/23808049/2sA2xb5F75#c8ef68b4-c329-41ae-8298-72f2336bbb78
		return "list/GetAll", nil
	default:
		return "", common.ErrOperationNotSupportedForObject
	}
}

// To determine the next page records for the objects.
func makeNextRecord(offset int) common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Extract the data key value from the response.
		value, err := jsonquery.New(node).ArrayRequired("items")
		if err != nil {
			return "", err
		}

		if len(value) == 0 {
			return "", nil
		}

		nextStart := offset + DefaultPageSize

		return strconv.Itoa(nextStart), nil
	}
}
