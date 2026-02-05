// nolint:revive,godoclint
package common

import (
	"errors"
)

var (
	// ErrMissingObjects is returned when no objects are provided in the request.
	ErrMissingObjects = errors.New("no objects provided")

	// ErrEmptyObject is returned when empty string is used as an object name.
	// Some APIs, connectors could be strict about it.
	ErrEmptyObject = errors.New("object name is empty")

	// ErrMissingRecordID is returned when resource id is missing in the request.
	ErrMissingRecordID = errors.New("no object ID provided")

	// ErrMissingRecordData is returned when write data is missing in the request.
	ErrMissingRecordData = errors.New("no data provided")

	// ErrMissingFields is returned when no fields are provided for reading.
	ErrMissingFields = errors.New("no fields provided in ReadParams")

	// ErrSinceUntilChronOrder is returned when the 'since' timestamp is after the 'until' timestamp in ReadParams.
	ErrSinceUntilChronOrder = errors.New("since cannot come after until")

	// ErrMissingFieldsMetadata is returned when the list of fields to create via UpsertMetadata is empty.
	ErrMissingFieldsMetadata = errors.New("no fields metadata provided in UpsertMetadata")

	// ErrMissingSearchFilters is returned when no field filters are provided for the Search operation.
	ErrMissingSearchFilters = errors.New("no filters provided for Search operation")

	// ErrPaginationControl is returned when controlling page size is not supported by connector..
	ErrPaginationControl = errors.New("pagination cannot be controlled by page size")
)

func (p ReadParams) ValidateParams(withRequiredFields bool) error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if withRequiredFields && len(p.Fields) == 0 {
		return ErrMissingFields
	}

	// If both 'since' and 'until' are set, ensure correct chronological order.
	if !p.Since.IsZero() && !p.Until.IsZero() {
		// Until must be after since, otherwise error.
		if p.Since.After(p.Until) {
			return ErrSinceUntilChronOrder
		}
	}

	return nil
}

func (p WriteParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if p.RecordData == nil {
		return ErrMissingRecordData
	}

	return nil
}

func (p DeleteParams) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if len(p.RecordId) == 0 {
		return ErrMissingRecordID
	}

	return nil
}

var (
	// ErrUnknownWriteType is returned when enum option for the write type is invalid.
	ErrUnknownWriteType = errors.New("unknown write type")
	// ErrUnsupportedWriteType is returned when connector doesn't implement write type.
	ErrUnsupportedWriteType = errors.New("write type is not supported")
)

func (p BatchWriteParam) ValidateParams() error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	// Neither "create" nor "update".
	if p.Type != WriteTypeCreate && p.Type != WriteTypeUpdate {
		return ErrUnknownWriteType
	}

	if len(p.Batch) == 0 {
		return ErrMissingRecordData
	}

	return nil
}

func (p *UpsertMetadataParams) ValidateParams() error {
	if p == nil {
		return ErrMissingFieldsMetadata
	}

	if len(p.Fields) == 0 {
		return ErrMissingFieldsMetadata
	}

	return nil
}

func (p SearchParams) ValidateParams(withRequiredFields bool) error {
	if len(p.ObjectName) == 0 {
		return ErrMissingObjects
	}

	if withRequiredFields && len(p.Fields) == 0 {
		return ErrMissingFields
	}

	if len(p.Filter.FieldFilters) == 0 {
		return ErrMissingSearchFilters
	}

	return nil
}
