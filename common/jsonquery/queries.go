package jsonquery

import (
	"math"

	"github.com/spyzhov/ajson"
)

// Query is a helpful wrapper of ajson library that adds errors when querying JSON payload.
//
// Usage examples, where node is JSON parsed via ajson library:
//
//	->	Must get *int64:	jsonquery.New(node).Integer("num", false)
//	->	Optional *string:	jsonquery.New(node).String("text", true)
//	->	Nested array:		jsonquery.New(node, "your", "path", "to", "array").Array("list", false)
type Query struct {
	node *ajson.Node
	zoom []string
}

// New constructs query searching for key. Extra keys are preceding forming a zoom path.
func New(node *ajson.Node, zoom ...string) *Query {
	return &Query{
		node: node,
		zoom: zoom,
	}
}

func (q *Query) Object(key string, optional bool) (*ajson.Node, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if !node.IsObject() {
		return nil, ErrNotObject
	}

	return node, nil
}

func (q *Query) Integer(key string, optional bool) (*int64, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	count, err := node.GetNumeric()
	if err != nil {
		return nil, ErrNotNumeric
	}

	if math.Mod(count, 1.0) != 0 {
		return nil, ErrNotInteger
	}

	result := int64(count)

	return &result, nil
}

func (q *Query) Str(key string, optional bool) (*string, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	txt, err := node.GetString()
	if err != nil {
		return nil, ErrNotString
	}

	return &txt, nil
}

func (q *Query) Array(key string, optional bool) ([]*ajson.Node, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	arr, err := node.GetArray()
	if err != nil {
		return nil, formatProblematicKeyError(key, ErrNotArray)
	}

	return arr, nil
}

func (q *Query) ArraySize(key string) (int64, error) {
	arr, err := q.Array(key, false)
	if err != nil {
		return 0, err
	}

	return int64(len(arr)), nil
}
