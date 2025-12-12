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

// ObjectOptional returns node object if present.
// If the entity at the key path is not a node object, an error is returned.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) ObjectOptional(key string) (*ajson.Node, error) {
	return q.internalQueryObject(key, true)
}

// ObjectRequired returns node object.
// If the entity at the key path is not a node object or is missing, an error is returned.
// Empty key is interpreter as "this", in other words current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *Query) ObjectRequired(key string) (*ajson.Node, error) {
	return q.internalQueryObject(key, false)
}

func (q *Query) internalQueryObject(key string, optional bool) (*ajson.Node, error) { // nolint:funcorder
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

// IntegerOptional returns integer if present.
// If the entity at the key path is not an integer, an error is returned.
// Empty key is interpreter as "this", in other words current node.
func (q *Query) IntegerOptional(key string) (*int64, error) {
	return q.internalQueryInteger(key, true)
}

// IntegerRequired returns integer.
// If the entity at the key path is not an integer or is missing, an error is returned.
// Empty key is interpreter as "this", in other words current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *Query) IntegerRequired(key string) (int64, error) {
	integer, err := q.internalQueryInteger(key, false)
	if err != nil {
		return 0, err
	}

	return *integer, nil
}

func (q *Query) internalQueryInteger(key string, optional bool) (*int64, error) { // nolint:funcorder
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

// StringOptional returns string if present.
// If the entity at the key path is not a string, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
func (q *Query) StringOptional(key string) (*string, error) {
	return q.internalQueryString(key, true)
}

// StringRequired returns string.
// If the entity at the key path is not a string or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *Query) StringRequired(key string) (string, error) {
	text, err := q.internalQueryString(key, false)
	if err != nil {
		return "", err
	}

	return *text, nil
}

func (q *Query) internalQueryString(key string, optional bool) (*string, error) { // nolint:funcorder
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

// BoolOptional returns boolean if present.
// If the entity at the key path is not a boolean, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
func (q *Query) BoolOptional(key string) (*bool, error) {
	return q.internalQueryBool(key, true)
}

// BoolRequired returns boolean.
// If the entity at the key path is not a boolean or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *Query) BoolRequired(key string) (bool, error) {
	flag, err := q.internalQueryBool(key, false)
	if err != nil {
		return false, err
	}

	return *flag, nil
}

func (q *Query) internalQueryBool(key string, optional bool) (*bool, error) { // nolint:funcorder
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

// ArrayOptional returns array of nodes if present.
// If the entity at the key path is not an array, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
func (q *Query) ArrayOptional(key string) ([]*ajson.Node, error) {
	return q.internalQueryArray(key, true)
}

// ArrayRequired returns array of nodes.
// If the entity at the key path is not an array or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *Query) ArrayRequired(key string) ([]*ajson.Node, error) {
	return q.internalQueryArray(key, false)
}

func (q *Query) internalQueryArray(key string, optional bool) ([]*ajson.Node, error) {
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
