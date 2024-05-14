package jsonquery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spyzhov/ajson"
)

func (q *Query) zoomIn() (*ajson.Node, error) {
	var err error

	node := q.node

	// traverse nested JSON, use every key to zoom in
	for _, key := range q.zoom {
		if !node.HasKey(key) {
			message := fmt.Sprintf("%v; zoom=%v", key, strings.Join(q.zoom, " "))

			return nil, createKeyNotFoundErr(message)
		}

		node, err = node.GetKey(key)
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (q *Query) getInnerKey(targetKey string, optional bool) (*ajson.Node, error) {
	zoomed, err := q.zoomIn()
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) && optional {
			// it is ok that part of zoom path didn't exist
			return nil, nil // nolint:nilnil
		}

		return nil, err
	}

	if !zoomed.HasKey(targetKey) {
		if optional {
			// null value in payload is allowed
			return nil, nil // nolint:nilnil
		} else {
			return nil, createKeyNotFoundErr(targetKey)
		}
	}

	targetNode, err := zoomed.GetKey(targetKey)
	if err != nil {
		return nil, err
	}

	if targetNode.IsNull() {
		return nil, handleNullNode(targetKey, optional)
	}

	return targetNode, nil
}
