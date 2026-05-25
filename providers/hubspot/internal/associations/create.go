package associations

import (
	"context"
	"fmt"

	"github.com/amp-labs/connectors/common"
)

// BatchCreate creates many associations from one object type to many other object types.
//
// Each association is represented as a pair of identifiers:
//
//	one from the source object (the "from" side) and
//	one from the target object (the "to" side).
//
// Both identifiers must exist for their respective object types;
//
//	otherwise the association may fail or be rejected by the backing API.
//
// The method returns a BatchCreateResult that indicates overall success and any errors
// encountered during the batch operation.
func (s Strategy) BatchCreate( // nolint:cyclop
	ctx context.Context, params *BatchCreateParams,
) (*BatchCreateResult, error) {
	if params == nil || len(params.payloads) == 0 {
		// Nothing to do.
		return &BatchCreateResult{Success: true}, nil
	}

	result := &BatchCreateResult{
		Success: true,
		Errors:  make([]error, 0),
	}

	objectSchema, warning, err := s.fetchObjectSchema(ctx, params.objectName)
	if err != nil {
		return nil, err
	}

	if warning != nil {
		// Error fetching object schema.
		// Nothing can be done without this data.
		result.AddError(warning)

		return result, nil
	}

	registry, warning := params.makePayloadsForRelationships(objectSchema)
	if warning != nil {
		result.AddError(warning)
	}

	// HubSpot's API expects homogeneous pairs of object IDs: one from the source object type
	// and one from the target object type.
	// Each relationship can involve many such pairs. The payload includes list of id pairs.
	for relationship, payload := range registry {
		url, err := s.getCreateAssociationsURL(relationship.from, relationship.to)
		if err != nil {
			return nil, err
		}

		resp, warning := s.clientCRM.Post(ctx, url.String(), payload)
		if warning != nil {
			// A request failure here is not a hard stop; we continue processing
			// the remaining relationships but save the error in the final result.
			result.AddError(warning)

			continue
		}

		response, err := common.UnmarshalJSON[BatchCreateResponse](resp)
		if err != nil {
			return nil, err
		}

		if len(response.Errors) != 0 {
			for _, responseError := range response.Errors {
				responseError.fromObject = relationship.from
				responseError.toObject = relationship.to
				result.AddError(responseError)
			}
		}
	}

	return result, nil
}

// BatchCreateResult holds the result of a Strategy.BatchCreate operation.
type BatchCreateResult struct {
	// Success indicates whether the operation completed without hard errors.
	// A false value usually means at least one request failed fully or partially.
	Success bool
	// Errors collects all individual failures encountered during the batch operation,
	Errors []error
}

func (r *BatchCreateResult) AddError(err error) {
	r.Success = false
	r.Errors = append(r.Errors, fmt.Errorf("batch create associations failure: %w", err))
}

// BatchCreateResponse is the response from the batch create associations API call.
// It mainly conveys error details when the operation partially or fully fails.
// Successful payloads are present in response but not exposed here.
type BatchCreateResponse struct {
	// Errors is a list of messages explaining why some or all associations failed.
	// It is returned when the status code is 207 Multi-Status.
	Errors []BatchCreateResponseError `json:"errors"`
}

type BatchCreateResponseError struct {
	Status     string `json:"status"`
	Category   string `json:"category"`
	Message    string `json:"message"`
	fromObject string
	toObject   string
}

func (e BatchCreateResponseError) Error() string {
	return fmt.Sprintf(
		"category(%v), message(%v), fromObject(%v), toObject(%v)",
		e.Category, e.Message, e.fromObject, e.toObject,
	)
}
