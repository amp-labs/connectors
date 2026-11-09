package auphan

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
)

// Reports are read one window at a time, so a backfill never asks a live POS for years of data in one call.
// The next window travels in NextPage as "<start>/<end>" (RFC 3339), where end is the end of the whole read.
const (
	reportWindowDays    = 30
	backfillStartYear   = 2000 // where a read without Since starts, as in the bigquery connector
	nextPageTokenFormat = time.RFC3339
)

var (
	errInvalidNextPage = errors.New("invalid next page token")
	errInvalidTimezone = errors.New("invalid timezone")
	errMissingNextPage = errors.New("report request has no next page token")
)

// reportPageKey carries a report request's NextPage from buildReportRequest to parseReadResponse.
type reportPageKey struct{}

type reportWindow struct {
	from time.Time // first second of this window
	to   time.Time // last second of this window (Auphan's date_to is inclusive)
	next string    // NextPage for the window after this one, empty on the last window
}

// window picks the report window for this read.
// Without Since, it starts at backfillStartYear; without Until, it ends now.
func (c *Connector) window(params common.ReadParams) (*reportWindow, error) {
	var from, end time.Time

	if params.NextPage != "" {
		var err error
		if from, end, err = parseNextPage(params.NextPage.String()); err != nil {
			return nil, err
		}
	} else {
		from, end = params.Since, params.Until

		if end.IsZero() {
			end = time.Now()
		}
	}

	// With a store timezone, times are converted to it. Without one, they keep their own location,
	// because Auphan does not report the store's timezone.
	if c.location != nil {
		from, end = from.In(c.location), end.In(c.location)
	}

	if from.IsZero() {
		from = time.Date(backfillStartYear, time.January, 1, 0, 0, 0, 0, end.Location())
	}

	next := from.AddDate(0, 0, reportWindowDays)
	if next.After(end) {
		return &reportWindow{from: from, to: end}, nil
	}

	return &reportWindow{
		from: from,
		to:   next.Add(-time.Second),
		next: next.Format(nextPageTokenFormat) + "/" + end.Format(nextPageTokenFormat),
	}, nil
}

// parseNextPage reads a "<start>/<end>" NextPage token.
func parseNextPage(token string) (time.Time, time.Time, error) {
	start, stop, ok := strings.Cut(token, "/")
	if !ok {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: %q", errInvalidNextPage, token)
	}

	from, err := time.Parse(nextPageTokenFormat, start)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: %w", errInvalidNextPage, err)
	}

	end, err := time.Parse(nextPageTokenFormat, stop)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: %w", errInvalidNextPage, err)
	}

	return from, end, nil
}

func withNextPage(ctx context.Context, next string) context.Context {
	return context.WithValue(ctx, reportPageKey{}, next)
}

// nextPageFrom returns the NextPage that buildReportRequest stored on the request.
func nextPageFrom(ctx context.Context) (string, error) {
	next, ok := ctx.Value(reportPageKey{}).(string)
	if !ok {
		return "", errMissingNextPage
	}

	return next, nil
}

func loadLocation(name string) (*time.Location, error) {
	if name == "" {
		return nil, nil //nolint:nilnil // no store timezone configured
	}

	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, fmt.Errorf("%w %q: %w", errInvalidTimezone, name, err)
	}

	return location, nil
}
