package jsonquery

import (
	"math"

	"github.com/spyzhov/ajson"
)

// Query is a helpful wrapper of ajson library that adds errors when querying JSON payload.
//
// Usage examples, where node is JSON parsed via ajson library:
//
//	->	Must get *int64:					jsonquery.New(node).Integer("num", false)
//	->	Optional *string:					jsonquery.New(node).String("text", true)
//	->	Nested array:						jsonquery.New(node, "your", "path", "to", "array").Array("list", false)
//	->	Convert current obj to list:		jsonquery.New(node).Array("", false)
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

// Object returns json object.
// Optional argument set to false will create error in case of missing value.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) Object(key string, optional bool) (*ajson.Node, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if node.IsNull() {
		return nil, handleNullNode(key, optional)
	}

	if !node.IsObject() {
		return nil, ErrNotObject
	}

	return node, nil
}

// Integer returns integer.
// Optional argument set to false will create error in case of missing value.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) Integer(key string, optional bool) (*int64, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if node.IsNull() {
		return nil, handleNullNode(key, optional)
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

// Str returns string.
// Optional argument set to false will create error in case of missing value.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) Str(key string, optional bool) (*string, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if node.IsNull() {
		return nil, handleNullNode(key, optional)
	}

	txt, err := node.GetString()
	if err != nil {
		return nil, ErrNotString
	}

	return &txt, nil
}

// Bool returns boolean.
// Optional argument set to false will create error in case of missing value.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) Bool(key string, optional bool) (*bool, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if node.IsNull() {
		return nil, handleNullNode(key, optional)
	}

	flag, err := node.GetBool()
	if err != nil {
		return nil, ErrNotBool
	}

	return &flag, nil
}

// Array returns list of nodes.
// Optional argument set to false will create error in case of missing value.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) Array(key string, optional bool) ([]*ajson.Node, error) {
	node, err := q.getInnerKey(key, optional)
	if err != nil {
		return nil, err
	}

	if node == nil {
		return nil, nil // nolint:nilnil
	}

	if node.IsNull() {
		return nil, handleNullNode(key, optional)
	}

	arr, err := node.GetArray()
	if err != nil {
		return nil, formatProblematicKeyError(key, ErrNotArray)
	}

	return arr, nil
}

// ArraySize returns the array size located under key.
// It is assumed that array value must be not null and present.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) ArraySize(key string) (int64, error) {
	arr, err := q.Array(key, false)
	if err != nil {
		return 0, err
	}

	return int64(len(arr)), nil
}
