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

// zoomIn executes the zoom traversal stored in the Query and returns
// the node located at the final zoom position.
//
// The Query acts as a traversal recipe:
//
//	q.node — starting JSON node
//	q.zoom — ordered path of keys describing an expected object hierarchy
//
// Each zoom element must resolve through a JSON object while traversal
// continues. All intermediate nodes are therefore required to be objects.
// Encountering a non-object value during traversal results in ErrNotObject.
//
// JSON null is treated specially: a null value is considered a valid
// terminal node. zoomIn may stop at null without error. Whether null is
// acceptable is decided by the caller (Required vs Optional semantics).
// Traversal requires objects; null terminates traversal but is not an error.
//
// Final node rules:
//
//	isUnwrap == true
//	    The final zoom node is returned as-is. The node may be any JSON
//	    value, including object, array, string, number, bool, or null.
//
//	isUnwrap == false
//	    The caller intends to perform an additional key lookup. The
//	    final zoom node must therefore be either an object or null.
//	    A concrete non-object value (string, number, bool, or array)
//	    results in ErrNotObject.
//
// Errors:
//   - ErrNotObject if traversal encounters a concrete non-object value
//     where an object is structurally required.
//   - ErrKeyNotFound if a zoom key does not exist.
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
