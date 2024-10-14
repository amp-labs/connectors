package dpobjects

import "github.com/amp-labs/connectors/common/handy"

// Object represents data associated with Ampersand Object.
// Every object is associated with certain URL path, field name where items are stored in the JSON node.
type Object struct {
	URLPath  string
	NodePath string
}

type Map = handy.Map[string, Object]
