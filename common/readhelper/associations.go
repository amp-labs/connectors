package readhelper

import (
	"context"
	"errors"
	"fmt"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/datautils"
)

var (
	ErrAssociationLookupNotImplemented = errors.New("associations lookup for an object is not implemented")
	ErrAssociationsUnsupported         = errors.New("associations between objects are not supported")
)

// HydrateAssociations returns a RowPostProcessor that hydrates read rows with
// related objects fetched through the provided fetcher.
//
// It collects row IDs, fetches associations for each requested related object via AssociationLookup,
// and attaches the resulting association data to the matching rows.
func HydrateAssociations(ctx context.Context,
	fromObjName common.ObjectName,
	toAssociatedObjects []string,
	lookup AssociationLookup,
) RowPostProcessor {
	return func(rows []common.ReadResultRow) error {
		return hydrateAssociations(ctx, fromObjName, toAssociatedObjects, rows, lookup)
	}
}

// AssociationLookup retrieves associations from one object type to another for
// the provided row IDs.
//
// It returns a map keyed by source row ID. Each value contains the associations
// discovered for that source row and target object.
//
// Implementations may return ErrAssociationLookupNotImplemented when lookup is
// not available for the given source object, or ErrAssociationsUnsupported when
// associations between the source and target objects are not supported.
type AssociationLookup func(
	ctx context.Context,
	fromObject common.ObjectName,
	fromIDs []RowID,
	toObject string,
) (map[RowID][]common.Association, error)

// RowID refers to common.ReadResultRow.Id.
type RowID = string

// hydrateAssociations populates rows with association data for the requested
// related object types.
//
// For each related object, it fetches associations keyed by source row ID and
// attaches them in place to the matching rows.
func hydrateAssociations(ctx context.Context,
	fromObjName common.ObjectName,
	toAssociatedObjects []string,
	rows []common.ReadResultRow,
	lookup AssociationLookup,
) error {
	if lookup == nil {
		return fmt.Errorf("%w: object %v", ErrAssociationLookupNotImplemented, fromObjName)
	}

	rowIDs := getUniqueIDs(rows)

	// Fetch associations for each source-target object pair.
	for _, targetObject := range toAssociatedObjects {
		// Fetch associations between two objects for the following source identifiers.
		associations, err := lookup(ctx, fromObjName, rowIDs, targetObject)
		if err != nil {
			if errors.Is(err, ErrAssociationLookupNotImplemented) {
				return fmt.Errorf("%w: object %v", ErrAssociationLookupNotImplemented, fromObjName)
			}

			if errors.Is(err, ErrAssociationsUnsupported) {
				return fmt.Errorf("%w: fromObject %v, toObject %v", err, fromObjName, targetObject)
			}

			return err
		}

		// Object has no associations to the targetObject.
		if len(associations) == 0 {
			continue
		}

		// Attach matching associations to the corresponding rows.
		for index, row := range rows {
			association, ok := associations[row.Id]
			if !ok {
				continue
			}

			// Init associations map.
			if rows[index].Associations == nil {
				rows[index].Associations = make(map[string][]common.Association)
			}

			// Store association
			rows[index].Associations[targetObject] = association
		}
	}

	return nil
}

// getUniqueIDs returns a slice of unsorted unique IDs from the given data.
func getUniqueIDs(rows []common.ReadResultRow) []RowID {
	identifiers := datautils.ForEach(rows, func(row common.ReadResultRow) string {
		return row.Id
	})

	return datautils.NewSetFromList(identifiers).List()
}
