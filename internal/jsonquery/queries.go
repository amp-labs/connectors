package jsonquery

import (
	"math"

	"github.com/spyzhov/ajson"
)

// Query is a helpful wrapper of jsonType library that adds errors when querying JSON payload.
//
// Usage examples, where node is JSON parsed via ajson library:
//
//	->	Must get *int64:					jsonquery.New(node).Integer("num", false)
//	->	Optional *string:					jsonquery.New(node).String("text", true)
//	->	Nested array:						jsonquery.New(node, "your", "path", "to", "array").Array("list", false)
//	->	Convert current obj to list:		jsonquery.New(node).Array("", false)
type Query[T any] interface {
	ObjectOptional(key string) (T, error)
	ObjectRequired(key string) (T, error)
	ArrayOptional(key string) ([]T, error)
	ArrayRequired(key string) ([]T, error)

	IntegerOptional(key string) (*int64, error)
	IntegerRequired(key string) (int64, error)
	StringOptional(key string) (*string, error)
	StringRequired(key string) (string, error)
	BoolOptional(key string) (*bool, error)
	BoolRequired(key string) (bool, error)

	IntegerWithDefault(key string, defaultValue int64) (int64, error)
	StrWithDefault(key string, defaultValue string) (string, error)
	TextWithDefault(key string, defaultValue string) (string, error)
	BoolWithDefault(key string, defaultValue bool) (bool, error)

	This() T
}

type jsonType interface {
	*ajson.Node | map[string]any
}
type NodeQuery struct {
	node *ajson.Node
	zoom []string
}
type MapQuery struct {
	node map[string]any
	zoom []string
}

// New constructs query searching for key. Extra keys are preceding forming a zoom path.
func New[T jsonType](json T, zoom ...string) Query[T] {
	switch data := any(json).(type) {
	case *ajson.Node:
		return any(&NodeQuery{
			node: data,
			zoom: zoom,
		}).(Query[T])
	case map[string]any:
		return any(&MapQuery{
			node: data,
			zoom: zoom,
		}).(Query[T])
	default:
		return any(&MapQuery{
			node: map[string]any{}, // empty map
			zoom: zoom,
		}).(Query[T]) // empty map
	}
}

// ObjectOptional returns node object if present.
// If the entity at the key path is not a node object, an error is returned.
// Empty key is interpreter as "this", in other words current node.
func (q *NodeQuery) ObjectOptional(key string) (*ajson.Node, error) {
	return q.internalQueryObject(key, true)
}

// ObjectRequired returns node object.
// If the entity at the key path is not a node object or is missing, an error is returned.
// Empty key is interpreter as "this", in other words current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *NodeQuery) ObjectRequired(key string) (*ajson.Node, error) {
	return q.internalQueryObject(key, false)
}

func (q *NodeQuery) internalQueryObject(key string, optional bool) (*ajson.Node, error) {
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
func (q *NodeQuery) IntegerOptional(key string) (*int64, error) {
	return q.internalQueryInteger(key, true)
}

// IntegerRequired returns integer.
// If the entity at the key path is not an integer or is missing, an error is returned.
// Empty key is interpreter as "this", in other words current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *NodeQuery) IntegerRequired(key string) (int64, error) {
	integer, err := q.internalQueryInteger(key, false)
	if err != nil {
		return 0, err
	}

	return *integer, nil
}

func (q *NodeQuery) internalQueryInteger(key string, optional bool) (*int64, error) {
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
func (q *NodeQuery) StringOptional(key string) (*string, error) {
	return q.internalQueryString(key, true)
}

// StringRequired returns string.
// If the entity at the key path is not a string or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *NodeQuery) StringRequired(key string) (string, error) {
	text, err := q.internalQueryString(key, false)
	if err != nil {
		return "", err
	}

	return *text, nil
}

func (q *NodeQuery) internalQueryString(key string, optional bool) (*string, error) {
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
func (q *NodeQuery) BoolOptional(key string) (*bool, error) {
	return q.internalQueryBool(key, true)
}

// BoolRequired returns boolean.
// If the entity at the key path is not a boolean or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *NodeQuery) BoolRequired(key string) (bool, error) {
	flag, err := q.internalQueryBool(key, false)
	if err != nil {
		return false, err
	}

	return *flag, nil
}

func (q *NodeQuery) internalQueryBool(key string, optional bool) (*bool, error) {
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
func (q *NodeQuery) ArrayOptional(key string) ([]*ajson.Node, error) {
	return q.internalQueryArray(key, true)
}

// ArrayRequired returns array of nodes.
// If the entity at the key path is not an array or is missing, an error is returned.
// Empty key is interpreted as "this", in other words a current node.
// Missing key returns ErrKeyNotFound. Null value returns ErrNullJSON.
func (q *NodeQuery) ArrayRequired(key string) ([]*ajson.Node, error) {
	return q.internalQueryArray(key, false)
}

func (q *NodeQuery) internalQueryArray(key string, optional bool) ([]*ajson.Node, error) {
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

func (q *NodeQuery) This() *ajson.Node {
	return q.node
}

func (m MapQuery) ObjectOptional(key string) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) ObjectRequired(key string) (map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) ArrayOptional(key string) ([]map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) ArrayRequired(key string) ([]map[string]any, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) IntegerOptional(key string) (*int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) IntegerRequired(key string) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) StringOptional(key string) (*string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) StringRequired(key string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) BoolOptional(key string) (*bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) BoolRequired(key string) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) IntegerWithDefault(key string, defaultValue int64) (int64, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) StrWithDefault(key string, defaultValue string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) TextWithDefault(key string, defaultValue string) (string, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) BoolWithDefault(key string, defaultValue bool) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (m MapQuery) This() map[string]any {
	return m.node
}
