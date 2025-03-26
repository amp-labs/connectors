package seismic

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(n *ajson.Node) (string, error) {
		// Reporting API do not support pagination.
		// https://developer.seismic.com/seismicsoftware/reference/h1-reporting-api-overview
		return "", nil
	}
}
