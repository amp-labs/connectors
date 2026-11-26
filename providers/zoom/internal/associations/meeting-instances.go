package associations

import (
	"context"
	"errors"
	"time"

	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/parallelfetch"
	"github.com/amp-labs/connectors/providers/zoom/internal/core"
)

const month = time.Hour * 24 * 30

// Returns a mapping from the meeting id to the latest meeting instance uuid.
//
// The registry may contain latest meeting uuid as an empty string.
// This must signify that the meeting has no latest instance during the last month.
func fetchLatestMeetingInstance(ctx context.Context,
	base *core.Base,
	meetingIds []readhelper.RowID,
) (map[string]string, error) {
	tasks := make([]parallelfetch.Task[string, string], len(meetingIds))
	for idx, meetingId := range meetingIds {
		tasks[idx] = func(ctx context.Context) (taskID string, data *string, err error) {
			instanceId, err := fetchLatestMeetingId(ctx, base, meetingId)

			return meetingId, instanceId, err
		}
	}

	result := parallelfetch.Execute(ctx, tasks, core.MaxReadConcurrency)
	if len(result.Errors) != 0 {
		return nil, errors.Join(result.Errors.Values()...)
	}

	return result.Records, nil
}

func fetchLatestMeetingId(ctx context.Context, base *core.Base, meetingId string) (*string, error) {
	url, err := base.GetMeetingInstancesUrl(meetingId)
	if err != nil {
		return nil, err
	}

	// Fetch meeting instances for the last month.
	// Docs state this must be done to improve performance.
	timestamp := time.Now()

	url.WithQueryParam("from", datautils.Time.FormatDateOnly(timestamp.Add(-1*month)))
	url.WithQueryParam("to", datautils.Time.FormatDateOnly(timestamp))

	resp, err := base.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	body, ok := resp.Body()
	if !ok {
		return new(string), nil
	}

	instances, err := jsonquery.New(body).ArrayOptional("meetings")
	if err != nil {
		return nil, err
	}

	if len(instances) == 0 {
		return new(string), nil
	}

	latestInstance := instances[0]

	id, err := jsonquery.New(latestInstance).StrWithDefault("uuid", "")
	if err != nil {
		return nil, err
	}

	return new(id), nil
}
