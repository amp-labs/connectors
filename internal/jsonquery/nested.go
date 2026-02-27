package jsonquery

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spyzhov/ajson"
)

// SelfReference indicates that the query refers to the node resolved
// by the zoom path itself rather than a child key.
//
// When used as targetKey, no additional lookup is performed and the
// resolved node is returned directly ("unwrap" semantics).
const SelfReference = ""

func (q *Query) zoomIn(isUnwrap bool) (*ajson.Node, error) {
	if len(q.zoom) == 0 {
		// Nothing to zoom into.
		return q.node, nil
	}

	var (
		node = q.node
		err  error
	)

	// traverse nested JSON, use every key to zoom in
	for _, key := range q.zoom {
		if !node.IsObject() {
			return nil, fmt.Errorf("%w: at key %v", ErrNotObject, key)
		}

		if !node.HasKey(key) {
			message := fmt.Sprintf("%v; zoom=%v", key, strings.Join(q.zoom, " "))

			return nil, createKeyNotFoundErr(message)
		}

		node, err = node.GetKey(key)
		if err != nil {
			return nil, err
		}
	}

	if isUnwrap {
		// The query is asking to get current node.
		// The outer method will assert the node type.
		// This acts as "unwrapping" current node into some JSON type defined by outer method.
		return node, nil
	}

	// Target key is not empty so we want to query some field inside an object.
	// Therefore, node must be an object.
	if !node.IsObject() {
		// Element at the last zoom position was not an object.
		key := q.zoom[len(q.zoom)-1]

		return nil, fmt.Errorf("%w: at key %v", ErrNotObject, key)
	}

	return node, nil
}

func (q *Query) getInnerKey(targetKey string, optional bool) (*ajson.Node, error) {
	isUnwrap := targetKey == SelfReference

	zoomed, err := q.zoomIn(isUnwrap)
	if err != nil {
		if errors.Is(err, ErrKeyNotFound) && optional {
			// it is ok that part of zoom path didn't exist
			return nil, nil // nolint:nilnil
		}

		return nil, err
	}

	// Empty key means we are referencing current node.
	if isUnwrap {
		targetNode := zoomed
		if targetNode.IsNull() {
			return nil, handleNullNode(targetKey, optional)
		}

		return targetNode, nil
	}

	if !zoomed.HasKey(targetKey) {
		if optional {
			// null value in payload is allowed
			return nil, nil // nolint:nilnil
		}

		return nil, createKeyNotFoundErr(targetKey)
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
