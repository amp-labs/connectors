package jsonquery

import (
	"math"

	"github.com/spyzhov/ajson"
)

// Query enforces strict structural expectations when traversing JSON.
// Package jsonquery provides strict, typed extraction utilities for JSON
// data represented as ajson.Node trees.
//
// The package is designed for situations where JSON is intentionally NOT
// unmarshalled into Go structs, but instead accessed dynamically while still
// enforcing strong structural (zoom) and type guarantees.
//
// A Query represents a reusable traversal recipe consisting of:
//   - a starting JSON node
//   - a zoom path describing an expected object hierarchy
//
// Queries separate JSON access into three explicit phases:
//
//	Phase 1 -- Structural navigation (zoom)
//		Walks through an expected object hierarchy.
//		Intermediate nodes must be JSON objects unless the value is null.
//		Missing keys may be allowed depending on optional/required mode.
//
//	Phase 2 -- Presence semantics
//		Determines whether missing or null values are acceptable.
//		Optional queries:
//			- missing keys are allowed
//			- JSON null values are allowed
//		Required queries:
//			- missing keys return ErrKeyNotFound
//			- JSON null returns ErrNullJSON
//
//	Phase 3 -- Type assertion
//		If a value exists and is non-null, it must match the requested type,
//		otherwise mismatches always return errors.
//		string		-> ErrNotString
//		integer		-> ErrNotNumeric (ErrNotInteger for floats)
//		object		-> ErrNotObject
//		array		-> ErrNotArray
//		bool		-> ErrNotBool
//
// ----------------------------------------
// Zoom path
//
//	jsonquery.New(node, "a", "b", "c")
//
// describes navigation equivalent to:
//
//	node["a"]["b"]["c"]
//
// All zoom elements are expected to resolve through JSON objects.
// A null value terminates traversal early and returns ErrKeyNotFound.
// Optional/Required will determine if this error should be ignored or propagated respectively.
//
// ----------------------------------------
// Self reference (unwrap)
// Passing an empty key ("") means:	"operate on the current zoom result".
//
// Example:
//
//	jsonquery.New(node, "user", "name").StringRequired("")
//
// extracts the value located exactly at the zoom destination,
// in this case "name". Value type of "user" must be an object.
// Note: `StringOptional("")` would allow "name" to be nil, or "user" to be missing.
//
// ----------------------------------------
// Optional vs Required.
// Optional queries relax only presence requirements.
// They DO NOT relax type expectations.
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

// ObjectOptional returns the object at key if present.
//
// Examples:
//
//	ObjectOptional("")          			// current node
//	ObjectOptional("user")      			// direct lookup
//	New(node, "a","b").ObjectOptional("") 	// node[a][b] -> object/nil
//	New(node, "a","b").ObjectOptional("c") 	// node[a][b][c] -> object/nil
//
// Errors:
// ErrNotObject, if non-null value is not an object.
func (q *Query) ObjectOptional(key string) (*ajson.Node, error) {
	return q.internalQueryObject(key, true)
}

// ObjectRequired returns the object at key.
//
// Examples:
//
//	ObjectRequired("")						// current node
//	ObjectRequired("user")					// direct lookup
//	New(node, "a","b").ObjectRequired("")	// node[a][b] -> object
//	New(node, "a","b").ObjectRequired("c")	// node[a][b][c] -> object
//
// Errors:
//
// ErrKeyNotFound, if zoom has missing key
// ErrNullJSON, if null
// ErrNotObject, if not an object.
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

// IntegerOptional returns the integer at key if present.
//
// Examples:
//
//	IntegerOptional("")          			// current node
//	IntegerOptional("count")      			// direct lookup
//	New(node, "a","b").IntegerOptional("")	// node[a][b] -> integer/nil
//	New(node, "a","b").IntegerOptional("c")	// node[a][b][c] -> integer/nil
//
// Errors:
// ErrNotNumeric, if value is not numeric
// ErrNotInteger, if numeric but not integral.
func (q *Query) IntegerOptional(key string) (*int64, error) {
	return q.internalQueryInteger(key, true)
}

// IntegerRequired returns the integer at key.
//
// Examples:
//
//	IntegerRequired("")						// current node
//	IntegerRequired("count")				// direct lookup
//	New(node, "a","b").IntegerRequired("")	// node[a][b] -> integer
//	New(node, "a","b").IntegerRequired("c")	// node[a][b][c] -> integer
//
// Errors:
// ErrKeyNotFound, if zoom has missing key
// ErrNullJSON, if null
// ErrNotNumeric, if value is not numeric
// ErrNotInteger, if numeric but not integral.
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

// StringOptional returns the string at key if present.
//
// Examples:
//
//	StringOptional("")          			// current node
//	StringOptional("name")      			// direct lookup
//	New(node, "a","b").StringOptional("")	// node[a][b] -> string/nil
//	New(node, "a","b").StringOptional("c")	// node[a][b][c] -> string/nil
//
// Errors:
// ErrNotString, if a non-null value is not a string.
func (q *Query) StringOptional(key string) (*string, error) {
	return q.internalQueryString(key, true)
}

// StringRequired returns the string at key.
//
// Examples:
//
//	StringRequired("")						// current node
//	StringRequired("name")					// direct lookup
//	New(node, "a","b").StringRequired("")	// node[a][b] -> string
//	New(node, "a","b").StringRequired("c")	// node[a][b][c] -> string
//
// Errors:
// ErrKeyNotFound, if zoom has missing key
// ErrNullJSON, if null
// ErrNotString, if not a string.
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

// BoolOptional returns the bool at key if present.
//
// Examples:
//
//	IntegerOptional("")          			// current node
//	BoolOptional("enabled")      			// direct lookup
//	New(node, "a","b").BoolOptional("")		// node[a][b] -> bool/nil
//	New(node, "a","b").BoolOptional("c")	// node[a][b][c] -> bool/nil
//
// Errors:
// ErrNotBool, if a non-null value is not boolean.
func (q *Query) BoolOptional(key string) (*bool, error) {
	return q.internalQueryBool(key, true)
}

// BoolRequired returns the bool at key.
//
// Examples:
//
//	BoolRequired("")						// current node
//	BoolRequired("enabled")					// direct lookup
//	New(node, "a","b").BoolRequired("")		// node[a][b] -> bool
//	New(node, "a","b").BoolRequired("c")	// node[a][b][c] -> bool
//
// Errors:
// ErrKeyNotFound, if zoom has missing key
// ErrNullJSON, if null
// ErrNotBool, if not boolean.
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

// ArrayOptional returns the array at key if present.
//
// Examples:
//
//	ArrayOptional("")          				// current node
//	ArrayOptional("items")      			// direct lookup
//	New(node, "a","b").ArrayOptional("")	// node[a][b] -> array/nil
//	New(node, "a","b").ArrayOptional("c")	// node[a][b][c] -> array/nil
//
// Errors:
// ErrNotArray, if a non-null value is not an array.
func (q *Query) ArrayOptional(key string) ([]*ajson.Node, error) {
	return q.internalQueryArray(key, true)
}

// ArrayRequired returns the integer at key.
//
// Examples:
//
//	ArrayRequired("")						// current node
//	ArrayRequired("count")					// direct lookup
//	New(node, "a","b").ArrayRequired("")	// node[a][b] -> integer
//	New(node, "a","b").ArrayRequired("c")	// node[a][b][c] -> integer
//
// Errors:
// ErrKeyNotFound, if zoom has missing key
// ErrNullJSON, if null
// ErrNotArray, if not an array.
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
