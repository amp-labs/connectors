package acculynx

import (
	"context"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// fetchAppointmentLookahead reads one calendar's appointments for a lookahead
// read (see appointmentLookaheadFlag), the way #3723 did: every window of
// appointmentWindows, with the rows that repeat an earlier row's id and start
// dropped (see dropRepeatedAppointmentRows). It runs only when the read mode is
// appointmentReadLookahead.
func (c *Connector) fetchAppointmentLookahead(
	ctx context.Context,
	parentID string,
	params common.ReadParams,
	nested nestedSpec,
	reqURL *urlbuilder.URL,
) ([]common.ReadResultRow, error) {
	rows, err := c.fetchAppointmentWindows(ctx, parentID, params, nested, reqURL)
	if err != nil {
		return nil, err
	}

	return dropRepeatedAppointmentRows(rows), nil
}

// dropRepeatedAppointmentRows returns one calendar's rows of a lookahead read
// without the rows whose id and start repeat an earlier row's. AccuLynx matches
// a window's date range loosely at the day boundary, so a row can come back
// from the windows on both sides of it. The key is id plus start, as AccuLynx
// wrote it, not id alone: a multi-day or recurring appointment comes back as
// one row per day under one id, and a lookahead read keeps every day. Rows
// without an id are kept.
func dropRepeatedAppointmentRows(rows []common.ReadResultRow) []common.ReadResultRow {
	kept := make([]common.ReadResultRow, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		appointmentID, _ := row.Raw["id"].(string)
		start, _ := row.Raw["start"].(string)

		if appointmentID != "" {
			key := appointmentID + "|" + start
			if _, repeated := seen[key]; repeated {
				continue
			}

			seen[key] = struct{}{}
		}

		kept = append(kept, row)
	}

	return kept
}
