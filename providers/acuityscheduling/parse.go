package acuityscheduling

import (
	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func nextRecordsURL() common.NextPageFunc {
	return func(node *ajson.Node) (string, error) {
		// Acuity Scheduling API does not support pagination.
		return "", nil
	}
}
