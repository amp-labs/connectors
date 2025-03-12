package endpoints

import (
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/components"
)

// Catalog holds registries that map ObjectNames to their respective URLs and REST methods.
// Registries are divided by Read, Write(Create/Update), Delete connector operations.
// To create OperationRegistry use NewOperationRegistry.
type Catalog struct {
	// Transport is used in URL construction.
	Transport *components.Transport

	// ReadOperation is used by connector's Read operation.
	// If reading from a static file, use OperationRegistryFromStaticSchema to populate this field.
	ReadOperation OperationRegistry

	// CreateOperation is used by connector's Write operation for new records (without a record ID).
	CreateOperation OperationRegistry

	// UpdateOperation is used by connector's Write operation for existing records (with a record ID).
	// Use {{.recordID}} in the URL template to be replaced with the actual record ID.
	// If omitted, the record ID is appended to the URL with a preceding slash.
	// Example:
	//   "/orders"					=> "/orders/123"
	//   "/orders/{{.recordID}}"	=> "/orders/123"
	//   "/orders({{.recordID}})"	=> "/orders(123)" -- for providers using special syntax.
	UpdateOperation OperationRegistry

	// DeleteOperation is used by connector's Delete operation.
	DeleteOperation OperationRegistry
}

// OperationContext encapsulates the necessary details for executing an API call.
// It is produced from OperationSpec.
// TODO: Endpoints may have exceptions in payload construction, headers. This struct needs to be extended to solve this.
// TODO: Ideally, the OperationContext should be as self-sufficient as possible.
type OperationContext struct {
	// URL of the API endpoint.
	// Note: Query parameters can be appended separately.
	URL *urlbuilder.URL

	// HTTP Method for the request (e.g., GET, POST, PUT, PATCH, DELETE).
	Method string
}

// CreateReadOperation constructs URL matching read operation registry in the catalog.
func (c *Catalog) CreateReadOperation(
	config common.ReadParams,
) (*OperationContext, error) {
	if c.Transport == nil || c.ReadOperation == nil {
		return &OperationContext{}, nil
	}

	// Find operation associated with the object.
	description := c.ReadOperation[c.Transport.Module()].Get(config.ObjectName)
	if description.isEmpty() {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Build URL according to the defined rules of
	baseURL := c.Transport.BaseURL()

	url, err := urlbuilder.New(baseURL, description.getURLPath(""))
	if err != nil {
		return nil, err
	}

	return &OperationContext{
		URL:    url,
		Method: description.Method,
	}, nil
}

// CreateWriteOperation constructs URL matching respective Create/Update operation registry in the catalog.
func (c *Catalog) CreateWriteOperation(
	config common.WriteParams,
) (*OperationContext, error) {
	registry := c.chooseWriteRegistry(config.RecordId)

	if c.Transport == nil || registry == nil {
		return &OperationContext{}, nil
	}

	// Find operation associated with the object.
	description := registry[c.Transport.Module()].Get(config.ObjectName)
	if description.isEmpty() {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Build URL according to the defined rules of
	baseURL := c.Transport.BaseURL()

	url, err := urlbuilder.New(baseURL, description.getURLPath(config.RecordId))
	if err != nil {
		return nil, err
	}

	return &OperationContext{
		URL:    url,
		Method: description.Method,
	}, nil
}

// CreateDeleteOperation constructs URL matching delete operation registry in the catalog.
func (c *Catalog) CreateDeleteOperation(
	config common.DeleteParams,
) (*OperationContext, error) {
	if c.Transport == nil || c.DeleteOperation == nil {
		return &OperationContext{}, nil
	}

	// Find operation associated with the object.
	description := c.DeleteOperation[c.Transport.Module()].Get(config.ObjectName)
	if description.isEmpty() {
		return nil, common.ErrOperationNotSupportedForObject
	}

	// Build URL according to the defined rules of
	baseURL := c.Transport.BaseURL()

	url, err := urlbuilder.New(baseURL, description.getURLPath(config.RecordId))
	if err != nil {
		return nil, err
	}

	return &OperationContext{
		URL:    url,
		Method: description.Method,
	}, nil
}

func (c *Catalog) chooseWriteRegistry(recordID string) OperationRegistry {
	if len(recordID) != 0 {
		return c.UpdateOperation
	}

	return c.CreateOperation
}
