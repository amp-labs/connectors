package outreach

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Record represents a single object record returned by the Outreach connector.
type Record struct {
	Data DataItem `json:"data"`
}

type Associations struct {
	ObjectId          string                          // ObjectId represents the id of the object we are reading
	AssociatedObjects map[string][]common.Association // Associated objects
}

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and
// configuration parameters. It returns the nested Attributes values read results or an error
// if the operation fails.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	if err := config.ValidateParams(true); err != nil {
		return nil, err
	}

	url, err := c.buildReadURL(config)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	// Unmarshal the *common.JSONHTTPResponse into data.
	data, err := common.UnmarshalJSON[Data](res)
	if err != nil {
		return nil, err
	}

	assoc := make([]Associations, 0)

	asc, err := c.fetchAssociations(ctx, data, config.AssociatedObjects, assoc)
	if err != nil {
		return nil, err
	}

	return common.ParseResult(res,
		getRecords,
		getNextRecordsURL,
		getDataMarshaller(common.FlattenNestedFields(attributesKey), asc),
		config.Fields,
	)
}

func (c *Connector) buildReadURL(config common.ReadParams) (*urlbuilder.URL, error) {
	// If NextPage is set, then we're reading the next page of results.
	// The NextPage URL has all the necessary parameters.
	if len(config.NextPage) > 0 {
		return urlbuilder.New(config.NextPage.String())
	}

	url, err := c.getApiURL(config.ObjectName)
	if err != nil {
		return nil, err
	}

	// If Since is not set, then we're doing a backfill. We read all rows (in pages)
	// If Since is present, we turn it into the format the Outreach API expects
	if !config.Since.IsZero() {
		t := config.Since.Format(time.DateOnly)
		fmtTime := t + "..inf"
		url.WithQueryParam("filter[updatedAt]", fmtTime)
	}

	return url, nil
}

func (c *Connector) fetchAssociations(ctx context.Context, d *Data, assc []string,
	assocII []Associations,
) ([]Associations, error) { //nolint:lll
	var err error

	for _, rcd := range d.Data {
		masc := make(map[string][]common.Association)

		for typ, dt := range rcd.Relationships {
			rel, ok := dt.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("invalid relationship structure for type %s", typ) // nolint: err113
			}

			switch data := rel["data"].(type) {
			case nil:
				continue
			case map[string]any:
				masc, err = c.processSingleAssociation(ctx, data, typ, assc, masc)
				if err != nil {
					return nil, fmt.Errorf("failed to process association %s: %w", typ, err)
				}
			case []any:
				masc, err = c.processMultipleAssociations(ctx, data, typ, assc, masc)
				if err != nil {
					return nil, fmt.Errorf("failed to process multiple associations for %s: %w", typ, err)
				}
			default:
				return nil, fmt.Errorf("unexpected data type for association %s", typ) // nolint: err113
			}
		}

		assocII = append(assocII, Associations{
			ObjectId:          strconv.Itoa(rcd.ID),
			AssociatedObjects: masc,
		})
	}

	return assocII, nil
}

func (c *Connector) processSingleAssociation(ctx context.Context, data map[string]any, typ string,
	assc []string, assoc map[string][]common.Association,
) (map[string][]common.Association, error) {
	objName, ok := data["type"].(string)
	if !ok {
		return nil, errors.New("missing or invalid 'type'") //nolint: err113
	}

	ascId, ok := data["id"].(float64)
	if !ok {
		return nil, errors.New("missing or invalid 'id'") //nolint: err113
	}

	// object type in the response is in the singular form of the objectname
	// but the Outreach APIs uses the plural form by default.
	objName = naming.NewPluralString(objName).String()
	path := objName + "/" + strconv.Itoa(int(ascId))

	// If the objectName is not in the associated request parameter, we return.
	// we only care for the requested assoctade objects.
	if !slices.Contains(assc, objName) {
		return assoc, nil
	}

	// don't make the call, if we already have the data.
	targetId := strconv.Itoa(int(ascId))
	for _, ass := range assoc[objName] {
		// Check if we already have this combination of ObjectId and AssociationType
		if ass.ObjectId == targetId && ass.AssociationType == typ {
			return assoc, nil
		}
	}

	// Check if we have the object with same id but different association type
	// if true, no need to make an API call, we can re-use the available data, just update the associationn type.
	// A good example for such scenario is when a sequence has a user for associationType creator and updator.
	// if the same user is the creator and updator, no need to make an extra call.
	for _, ass := range assoc[objName] {
		if ass.ObjectId == targetId {
			assoc[objName] = append(assoc[objName], common.Association{
				ObjectId:        targetId,
				AssociationType: typ,
				Raw:             ass.Raw,
			})

			return assoc, nil
		}
	}

	// TODO Maybe before making the call, check if in the previous objectids assocaitions.
	// we already have similar objectId + associated type.

	// when we have no such object, we make the API call.
	assRec, err := c.getAssociation(ctx, path)
	if err != nil {
		return nil, err
	}

	return addAssociation(assoc, typ, objName, targetId, assRec)
}

func (c *Connector) processMultipleAssociations(ctx context.Context, data []any, typ string, assc []string,
	assoc map[string][]common.Association,
) (map[string][]common.Association, error) {
	var err error

	for _, d := range data {
		rcd, ok := d.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid relationship structure for type %s", typ) //nolint: err113
		}

		assoc, err = c.processSingleAssociation(ctx, rcd, typ, assc, assoc)
		if err != nil {
			return nil, err
		}
	}

	return assoc, nil
}

func (c *Connector) getAssociation(ctx context.Context, path string) (*Record, error) {
	u, err := c.getApiURL(path)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(ctx, u.String())
	if err != nil {
		return nil, err
	}

	d, err := common.UnmarshalJSON[Record](resp)
	if err != nil {
		return nil, err
	}

	return d, nil
}

func addAssociation(assoc map[string][]common.Association, typ, objName string,
	id string, d *Record,
) (map[string][]common.Association, error) {
	assoc[objName] = append(assoc[objName], common.Association{
		ObjectId:        id,
		AssociationType: typ,
		Raw: map[string]any{
			"type":          d.Data.Type,
			"id":            d.Data.ID,
			"links":         d.Data.Links,
			"attributes":    d.Data.Attributes,
			"relationships": d.Data.Relationships,
		},
	})

	return assoc, nil
}
