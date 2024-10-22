package requirements

// ComponentID represents unique identifier of a component which serves one concrete role in connector implementation.
type ComponentID string

// Single implementation.
const (
	Connector        ComponentID = "connector"
	Provider         ComponentID = "provider"
	Options          ComponentID = "options"
	Parameters       ComponentID = "parameters"
	CatalogVariables ComponentID = "catalogVariables"
	ConnectorData    ComponentID = "connectorData"
	ErrorHandler     ComponentID = "errorHandler"
	Clients          ComponentID = "clients"
	Closer           ComponentID = "closer"
)

// Multiple Implementations.
const MetadataVariables ComponentID = "metadataVariables"
