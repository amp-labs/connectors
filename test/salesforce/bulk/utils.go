package bulk

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/providers/salesforce"
	"github.com/amp-labs/connectors/test/utils"
)

// LoadQueryResults is shared test procedure to wait and get query results.
func LoadQueryResults(ctx context.Context, conn *salesforce.Connector, jobId string) {
	if _, err := getInfoInLoop(ctx, conn, jobId); err != nil {
		utils.Fail("Error getting job results", "error", err)
	}

	slog.Info("Job completed... fetching results")

	// Get the results
	result, err := conn.GetBulkQueryResults(ctx, jobId)
	if err != nil {
		utils.Fail("Error getting query results", "error", err)
	}

	body := common.GetResponseBodyOnce(result)

	slog.Info("Query results")
	fmt.Println(string(body))
}

func GetResultInLoop(
	ctx context.Context, conn *salesforce.Connector, jobId string,
) (*salesforce.JobResults, error) {
	return utils.CycleUntilComplete(2*time.Second, func() (*salesforce.JobResults, error) {
		return conn.GetJobResults(ctx, jobId)
	})
}

func getInfoInLoop(
	ctx context.Context, conn *salesforce.Connector, jobId string,
) (*salesforce.GetJobInfoResult, error) {
	return utils.CycleUntilComplete(2*time.Second, func() (*salesforce.GetJobInfoResult, error) {
		return conn.GetBulkQueryInfo(ctx, jobId)
	})
}
