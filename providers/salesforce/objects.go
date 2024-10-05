package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) GetSupportedObjects(ctx context.Context) ([]common.SupportedObject, error) {
	url, err := c.getURISobjects()
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	objectsResponse, err := common.UnmarshalJSON[SObjectsResponse](res)
	if err != nil {
		return nil, err
	}

	sObjects := objectsResponse.Sobjects
	result := make([]common.SupportedObject, 0, len(sObjects))

	for _, object := range sObjects {
		if object.Queryable {
			result = append(result, common.SupportedObject{
				Name:        object.Name,
				DisplayName: object.Label,
			})
		}
	}

	return result, nil
}

type SObjectsResponse struct {
	Encoding     string    `json:"encoding"`
	MaxBatchSize int       `json:"maxBatchSize"`
	Sobjects     []SObject `json:"sobjects"`
}

type SObject struct {
	Name      string `json:"name"`
	Label     string `json:"label"`
	Queryable bool   `json:"queryable"`
}
