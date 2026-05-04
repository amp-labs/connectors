package subscriber

import (
	"github.com/amp-labs/connectors/common"
)

type (
	// State describes the set of object events that the connector should
	// try to achieve (desired) and that are also reported back as the observed outcome.
	//
	// It is part of the connector's input/output contract: the caller expresses intent
	// with State, and the connector returns the same type reflecting what actually
	// happened (which may differ on failure or rollback).
	State map[ObjectName]common.ObjectEvents
)

func newState(objectNames []ObjectName) State {
	events := make(State)
	for _, objectName := range objectNames {
		events[objectName] = common.ObjectEvents{}
	}

	return events
}
