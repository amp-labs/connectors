package dpobjects

import (
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/deep/requirements"
)

var (
	// Implementations.
	_ Support     = Registry{}
	_ URLResolver = URLFormat{}
)

// Support is a connector component, which describes what objects are supported by which operation.
//
// The goal of this interface is to achieve:
// * ObjectName ==> Is supported?
type Support interface {
	requirements.ConnectorComponent

	IsReadSupported(objectName string) bool
	IsWriteSupported(objectName string) bool
	IsDeleteSupported(objectName string) bool
}

type Method string

const (
	ReadMethod   Method = "READ"
	CreateMethod Method = "CREATE"
	UpdateMethod Method = "UPDATE"
	DeleteMethod Method = "DELETE"
)

// URLResolver is a connector component, which knows how to build URL for an object.
// Every object is associated with URL. The URL may differ based on HTTP Method.
// Ex: object: orders, GET /orders, POST /cart/finalize
//
// The goal of this interface is to achieve:
// * ObjectName ==> URL
type URLResolver interface {
	requirements.ConnectorComponent

	FindURL(method Method, baseURL, objectName string) (*urlbuilder.URL, error)
}
