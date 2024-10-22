package requirements

// ComponentID represents unique identifier of a component which serves one concrete role in connector implementation.
type ComponentID string

// Single implementation.
const (
	Connector          ComponentID = "connector"
	Provider           ComponentID = "provider"
	Options            ComponentID = "options"
	Parameters         ComponentID = "parameters"
	CatalogVariables   ComponentID = "catalogVariables"
	ConnectorData      ComponentID = "connectorData"
	ErrorHandler       ComponentID = "errorHandler"
	HeaderSupplements  ComponentID = "headerSupplements"
	Clients            ComponentID = "clients"
	Closer             ComponentID = "closer"
	Reader             ComponentID = "reader"
	Writer             ComponentID = "writer"
	ReadObjectLocator  ComponentID = "readObjectLocator"
	NextPageBuilder    ComponentID = "nextPageBuilder"
	WriteResultBuilder ComponentID = "writeResultBuilder"
)

// Multiple Implementations.
const (
	ObjectRegistry         ComponentID = "objectRegistry"
	ObjectURLResolver      ComponentID = "objectUrlResolver"
	PaginationStartBuilder ComponentID = "paginationStartBuilder"
	ReadRequestBuilder     ComponentID = "readRequestBuilder"
	WriteRequestBuilder    ComponentID = "writeRequestBuilder"
	MetadataVariables      ComponentID = "metadataVariables"
)
