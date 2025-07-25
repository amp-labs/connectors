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

type RecordAssociations struct {
	ObjectId          string                          // ObjectId represents the id of the object we are reading
	AssociatedObjects map[string][]common.Association // Associated objects
}

// Read retrieves data based on the provided configuration parameters.
//
// This function executes a read operation using the given context and
// configuration parameters. It returns the nested Attributes values read results or an error
// if the operation fails.
func (c *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var asc []RecordAssociations

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

	rasc := make([]RecordAssociations, 0)

	// If were reading with associations, we make the API calls to retrieve associated objects.
	if len(config.AssociatedObjects) > 0 {
		asc, err = c.fetchAssociations(ctx, data, config.AssociatedObjects, rasc)
		if err != nil {
			return nil, err
		}
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

func (c *Connector) fetchAssociations(ctx context.Context, d *Data, asc []string, //nolint: lll, cyclop
	ras []RecordAssociations,
) ([]RecordAssociations, error) {
	var err error

	// we need this to keep track of the list of the available associated objects
	// so we can compare it with the requested objects.
	var ascObj []string

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
				masc, ascObj, err = c.processSingleAssociation(ctx, data, typ, asc, ascObj, masc)
				if err != nil {
					return nil, fmt.Errorf("failed to process association %s: %w", typ, err)
				}
			case []any:
				masc, ascObj, err = c.processMultipleAssociations(ctx, data, typ, asc, ascObj, masc)
				if err != nil {
					return nil, fmt.Errorf("failed to process multiple associations for %s: %w", typ, err)
				}
			default:
				return nil, fmt.Errorf("unexpected data type for association %s", typ) // nolint: err113
			}
		}

		ras = append(ras, RecordAssociations{
			ObjectId:          strconv.Itoa(rcd.ID),
			AssociatedObjects: masc,
		})
	}

	// We use here to double-check if all requested associated objects are available in the response associated list.
	// else we error out.
	for _, obj := range asc {
		if !slices.Contains(ascObj, obj) {
			return nil, fmt.Errorf("couldn't find associated records of: %s", obj) //nolint: err113
		}
	}

	return ras, nil
}

func (c *Connector) processSingleAssociation(ctx context.Context, data map[string]any, typ string, // nolint: funlen
	asc []string, ascObj []string, masc map[string][]common.Association,
) (map[string][]common.Association, []string, error) {
	objName, ok := data["type"].(string)
	if !ok {
		return nil, ascObj, errors.New("missing or invalid 'type'") //nolint: err113
	}

	ascId, ok := data["id"].(float64)
	if !ok {
		return nil, ascObj, errors.New("missing or invalid 'id'") //nolint: err113
	}

	// object type in the response is in the singular form of the objectname
	// but the Outreach APIs uses the plural form by default.
	objName = naming.NewPluralString(objName).String()
	path := objName + "/" + strconv.Itoa(int(ascId))

	ascObj = append(ascObj, objName)

	// If the objectName is not in the associated request parameter, we return.
	// we only care for the requested assoctade objects.
	if !slices.Contains(asc, objName) {
		return masc, ascObj, nil
	}

	// don't make the call, if we already have the data.
	targetId := strconv.Itoa(int(ascId))
	for _, ass := range masc[objName] {
		// Check if we already have this combination of ObjectId and AssociationType
		if ass.ObjectId == targetId && ass.AssociationType == typ {
			return masc, ascObj, nil
		}
	}

	// Check if we have the object with same id but different association type
	// if true, no need to make an API call, we can re-use the available data, just update the associationn type.
	// A good example for such scenario is when a sequence has a user for associationType creator and updator.
	// if the same user is the creator and updator, no need to make an extra call.
	for _, ass := range masc[objName] {
		if ass.ObjectId == targetId {
			masc[objName] = append(masc[objName], common.Association{
				ObjectId:        targetId,
				AssociationType: typ,
				Raw:             ass.Raw,
			})

			return masc, ascObj, nil
		}
	}

	// TODO Maybe before making the call, check if in the previous objectids assocaitions.
	// we already have similar objectId + associated type.

	// when we have no such object, we make the API call.
	assRec, err := c.getAssociation(ctx, path)
	if err != nil {
		return nil, ascObj, err
	}

	masc = addAssociation(masc, typ, objName, targetId, assRec)

	return masc, ascObj, nil
}

func (c *Connector) processMultipleAssociations(ctx context.Context, data []any, typ string, asc []string,
	ascObj []string, masc map[string][]common.Association,
) (map[string][]common.Association, []string, error) {
	var err error

	for _, d := range data {
		rcd, ok := d.(map[string]any)
		if !ok {
			return nil, ascObj, fmt.Errorf("invalid relationship structure for type %s", typ) //nolint: err113
		}

		masc, ascObj, err = c.processSingleAssociation(ctx, rcd, typ, asc, ascObj, masc)
		if err != nil {
			return nil, ascObj, err
		}
	}

	return masc, ascObj, nil
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

func addAssociation(masc map[string][]common.Association, typ, objName string,
	id string, d *Record,
) map[string][]common.Association {
	masc[objName] = append(masc[objName], common.Association{
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

	return masc
}
