package dynamicscrm

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	ErrFetchEntityDefinition    = errors.New("failed to fetch object's entity definition")
	ErrFetchAttributes          = errors.New("failed to fetch object's attributes")
	ErrFetchAttributesPicklists = errors.New("failed to fetch object's PicklistType attributes")
	ErrFetchAttributesStatuses  = errors.New("failed to fetch object's StatusType attributes")
	ErrFetchAttributesStates    = errors.New("failed to fetch object's StateType attributes")
)

// This repository is a grouping of Microsoft Dataverse API concerned with fetching Object metadata.
//
// It provides with the following data:
// * entityDefinitionResponse - metadata of an object
// * attributesResponse - complete list of attributes that exist on the object
// Attributes with enumeration options grouped by attribute type:
// * attributesPicklistsResponse.
// * attributesStatusesResponse.
// * attributesStatesResponse.
type metadataDiscoveryRepository struct {
	client   *common.JSONHTTPClient
	buildURL func(path string) (*urlbuilder.URL, error)
}

func (r metadataDiscoveryRepository) fetchEntityDefinition(
	ctx context.Context, objectName naming.SingularString,
) (dao *entityDefinitionResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrFetchEntityDefinition, err)
		}
	}()
	// This endpoint returns schema of an object.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')", objectName.String())

	url, err := r.buildURL(path)
	if err != nil {
		return nil, err
	}

	// the only field we care about in response
	url.WithQueryParam("$select", "DisplayCollectionName")

	resp, err := r.performGetRequest(ctx, url)
	if err != nil {
		return nil, errors.Join(ErrObjectNotFound, err)
	}

	return common.UnmarshalJSON[entityDefinitionResponse](resp)
}

func (r metadataDiscoveryRepository) fetchAttributes(
	ctx context.Context, objectName naming.SingularString,
) (dao *attributesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrFetchAttributes, err)
		}
	}()
	// This endpoint will describe attributes present on schema and its properties.
	// Schema name must be singular.
	path := fmt.Sprintf("EntityDefinitions(LogicalName='%v')/Attributes", objectName.String())

	url, err := r.buildURL(path)
	if err != nil {
		return nil, err
	}

	// Filter attributes to ensure they are:
	// 1. Present in Read responses (IsValidODataAttribute == true)
	// 2. Can be queried in GET requests (IsValidForRead == true)
	// This ensures we only work with fields that are both
	// returned in the payload and can be used in query parameters.
	url.WithQueryParam("$filter", "(IsValidODataAttribute eq true and IsValidForRead eq true)")
	// We cannot use $select clause to scope response, unfortunately, `Targets` field breaks $select.
	// Falling back to requesting the whole payload.

	resp, err := r.performGetRequest(ctx, url)
	if err != nil {
		return nil, errors.Join(ErrObjectNotFound, err)
	}

	return common.UnmarshalJSON[attributesResponse](resp)
}

func (r metadataDiscoveryRepository) fetchAttributesPicklists(
	ctx context.Context, objectName naming.SingularString,
) (dao *attributesPicklistsResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrFetchAttributesPicklists, err)
		}
	}()

	resp, err := r.getOptionsForAttribute(ctx, objectName, "Microsoft.Dynamics.CRM.PicklistAttributeMetadata")
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[attributesPicklistsResponse](resp)
}

func (r metadataDiscoveryRepository) fetchAttributesStatuses(
	ctx context.Context, objectName naming.SingularString,
) (dao *attributesStatusesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrFetchAttributesStatuses, err)
		}
	}()

	resp, err := r.getOptionsForAttribute(ctx, objectName, "Microsoft.Dynamics.CRM.StatusAttributeMetadata")
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[attributesStatusesResponse](resp)
}

func (r metadataDiscoveryRepository) fetchAttributesStates(
	ctx context.Context, objectName naming.SingularString,
) (dao *attributesStatesResponse, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("%w: %w", ErrFetchAttributesStates, err)
		}
	}()

	resp, err := r.getOptionsForAttribute(ctx, objectName, "Microsoft.Dynamics.CRM.StateAttributeMetadata")
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[attributesStatesResponse](resp)
}

func (r metadataDiscoveryRepository) getOptionsForAttribute(
	ctx context.Context, objectName naming.SingularString, attributeName string,
) (*common.JSONHTTPResponse, error) {
	// This endpoint will fetch options metadata for an attribute.
	path := "EntityDefinitions(LogicalName='" + objectName.String() + "')" +
		"/Attributes/" + attributeName

	url, err := r.buildURL(path)
	if err != nil {
		return nil, err
	}

	// Request nested optionSet for each attribute of Picklist type for current object.
	// Reference: https://community.dynamics.com/forums/thread/details/?threadid=37c94bd2-7703-471b-a726-bcfea2bfa776
	url.WithQueryParam("$expand", "OptionSet($select=Options)")
	// Logical name is used to identify the attribute we are dealing with.
	url.WithQueryParam("$select", "LogicalName")

	resp, err := r.performGetRequest(ctx, url)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (r metadataDiscoveryRepository) performGetRequest(
	ctx context.Context, url *urlbuilder.URL,
) (*common.JSONHTTPResponse, error) {
	rsp, err := r.client.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	if _, ok := rsp.Body(); !ok {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return rsp, nil
}

// nolint:tagliatelle
type entityDefinitionResponse struct {
	DisplayCollectionName struct {
		localizedLabels
	} `json:"DisplayCollectionName"`
}

type attributesResponse struct {
	Values []attributeItem `json:"value"`
}

// nolint:tagliatelle
type attributeItem struct {
	LogicalName       string   `json:"LogicalName"`
	SchemaName        string   `json:"SchemaName"`
	Targets           []string `json:"Targets"`
	IsValidForCreate  bool     `json:"IsValidForCreate"`
	IsValidForUpdate  bool     `json:"IsValidForUpdate"`
	AttributeTypeName struct {
		Value string `json:"Value"`
	} `json:"AttributeTypeName"`
	DisplayName struct {
		localizedLabels
	} `json:"DisplayName"`
	Format string `json:"Format"`
}

// nolint:tagliatelle
type localizedLabels struct {
	LocalizedLabels []struct {
		Label string `json:"Label"`
	} `json:"LocalizedLabels"`
}

func (l localizedLabels) getName() (string, bool) {
	labels := l.LocalizedLabels
	if len(labels) == 0 {
		return "", false
	}

	return labels[0].Label, true
}

type attributesPicklistsResponse struct {
	attributesWithOptions
}

type attributesStatusesResponse struct {
	attributesWithOptions
}

type attributesStatesResponse struct {
	attributesWithOptions
}

// nolint:tagliatelle
type attributesWithOptions struct {
	Value []struct {
		LogicalName string `json:"LogicalName"`
		MetadataId  string `json:"MetadataId"`
		OptionSet   struct {
			MetadataId string `json:"MetadataId"`
			Options    []struct {
				Value    int  `json:"Value"`
				IsHidden bool `json:"IsHidden"`
				Label    struct {
					localizedLabels
				} `json:"Label"`
			} `json:"Options"`
		} `json:"OptionSet"`
	} `json:"value"`
}

func (a attributesWithOptions) getOptionsPerAttribute() datautils.NamedLists[common.FieldValue] {
	result := make(datautils.NamedLists[common.FieldValue])

	for _, attribute := range a.Value {
		for _, option := range attribute.OptionSet.Options {
			if !option.IsHidden {
				name, _ := option.Label.getName()
				result.Add(attribute.LogicalName, common.FieldValue{
					Value:        strconv.FormatInt(int64(option.Value), 10),
					DisplayValue: name,
				})
			}
		}
	}

	return result
}
