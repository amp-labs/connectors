package batch

import (
	"errors"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common/urlbuilder"
)

// ErrURLNotEnoughSpace is returned when the base URL is already too long to accommodate
// any identifiers in the 'conditions' query parameter.
var ErrURLNotEnoughSpace = errors.New("URL is too long to add any identifiers in the 'conditions' query")

// identifierGroup holds a batch of identifiers that fit within the available URL space.
type identifierGroup struct {
	// IDs is the list of identifiers in this group.
	IDs []string
	// TotalSize is the total character count of the encoded identifiers plus commas.
	TotalSize int
}

type urlWrapper struct {
	// URL is the complete URL with the 'conditions' query parameter containing the identifiers.
	URL string
	// estimatedSize is the expected length as calculated by the algorithm.
	estimatedSize int
}

// withIdentifiers generates multiple URLs from a base URL by adding identifiers to the
// 'conditions' query parameter (format: "id in (1,2,3)"). When the number of identifiers
// would cause the URL to exceed maxURLLength, it splits them into multiple URLs.
//
// URL encoding is accounted for:
//   - Comma ',' becomes '%2C' (3 characters instead of 1)
//   - Parentheses '()' become '%28' and '%29' (3 characters each)
//   - Space ' ' becomes '+' (1 character, same length)
//
// Parameters:
//   - baseURL: The base URL with path any preexisting query params but without the 'conditions' query parameter
//   - identifiers: List of identifier strings to include in the conditions
//   - maxURLLength: Maximum allowed URL length (e.g., 2000 for ConnectWise API)
//
// Returns:
//   - A slice of URLs
//   - ErrURLNotEnoughSpace is returned if maxURLLength is too small to accommodate even a single identifier
//   - One URL if so identifiers are set.
func withIdentifiers(baseURL *urlbuilder.URL, identifiers []string, maxURLLength int) ([]urlWrapper, error) {
	// Calculate available space for identifiers after accounting for base URL and existing query structure.
	base := baseURL.String()
	baseSize := len(base)

	if len(identifiers) == 0 {
		return []urlWrapper{{
			URL:           baseURL.String(),
			estimatedSize: baseSize,
		}}, nil
	}

	// Fixed query parameter structure (encoded):
	// "?conditions=id+in+%28%29" represents "?conditions=id in ()"
	conditionsQuerySize := len("?conditions=id+in+%28%29")
	reservedSize := baseSize + conditionsQuerySize
	leftOverSpace := maxURLLength - reservedSize

	// Group identifiers so each group fits within the available space.
	identifierChunks, err := groupIdentifiersBySpace(identifiers, leftOverSpace)
	if err != nil {
		return nil, err
	}

	// Build URLs for each group of identifiers.
	urls := make([]urlWrapper, 0)

	for _, group := range identifierChunks {
		url, err := urlbuilder.New(base)
		if err != nil {
			return nil, err
		}
		url.WithQueryParam("conditions", fmt.Sprintf("id in (%v)", strings.Join(group.IDs, ",")))
		urls = append(urls, urlWrapper{
			URL:           url.String(),
			estimatedSize: group.TotalSize + reservedSize,
		})
	}

	return urls, nil
}

// groupIdentifiersBySpace splits identifiers into groups where each group's total size
// (identifiers plus encoded commas) does not exceed maxSpace.
//
// Encoding considerations:
//   - Each comma between identifiers becomes '%2C' (3 characters)
//   - First identifier has no preceding comma
//
// Returns ErrURLNotEnoughSpace if a single identifier exceeds maxSpace.
func groupIdentifiersBySpace(ids []string, maxSpace int) ([]identifierGroup, error) {
	var (
		groups = make([]identifierGroup, 0)
		group  = identifierGroup{
			IDs:       make([]string, 0),
			TotalSize: 0,
		}
	)

	for _, identifier := range ids {
		partSize := len(identifier)
		if group.TotalSize != 0 {
			// Add comma size for non-first items: ',' becomes '%2C' (3 characters).
			partSize += 3
		}

		// Check if adding this identifier would exceed available space.
		potentialSize := group.TotalSize + partSize
		if potentialSize <= maxSpace {
			// Safe to add identifier without exceeding the limit.
			group.IDs = append(group.IDs, identifier)
			group.TotalSize += partSize
		} else {
			if group.TotalSize == 0 || len(identifier) > maxSpace {
				// This identifier cannot fit: either no items in current group yet,
				// or single identifier itself exceeds maxSpace.
				return nil, ErrURLNotEnoughSpace
			}

			// The current group is full; save it and start a new group.
			groups = append(groups, group)

			// Start new group with current identifier.
			group = identifierGroup{
				IDs:       []string{identifier},
				TotalSize: len(identifier),
			}
		}
	}

	// Add the final group (likely not completely full).
	if len(group.IDs) > 0 {
		groups = append(groups, group)
	}

	return groups, nil
}
