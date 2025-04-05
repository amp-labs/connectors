package heyreach

import "github.com/amp-labs/connectors/common"

func matchObjectNameToEndpointPath(objectName string) (urlPath string, err error) {
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
