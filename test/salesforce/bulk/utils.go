package bulk

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"log/slog"
	"strings"
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

func GetRecordIDsForJob(ctx context.Context, conn *salesforce.Connector, jobId string)  ([]byte, error) {
	// Get the successful results to get the ids to use for the deletion
	successRes, err := conn.GetSuccessfulJobResults(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("error getting successfult write results: %w", err)
	}

	successBody, err := io.ReadAll(successRes.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading success results body: %w", err)
	}

	defer func() {
		if successRes != nil && successRes.Body != nil {
			if closeErr := successRes.Body.Close(); closeErr != nil {
				slog.Warn("unable to close response body", "error", closeErr)
			}
		}
	}()

	headers, rows, err := csvBytesToSlice(successBody)
	if err != nil {
		return nil, fmt.Errorf("error parsing CSV: %w", err)
	}

	// remove all columns except the id column
	csvRecords, err := filterIds(headers, rows)
	if err != nil {
		return nil, fmt.Errorf("error filtering ids: %w", err)
	}

	var b []byte
	buf := bytes.NewBuffer(b)
	w := csv.NewWriter(buf)

	for _, record := range csvRecords {
		if err := w.Write(record); err != nil {
			log.Fatalln("error writing record to csv:", err)
		}
	}

	w.Flush()

	if err := w.Error(); err != nil {
		log.Fatalln(err)
	}

	return buf.Bytes(), nil
}

func filterIds(headers []string, rows [][]string) ([][]string, error) {
	// filter ids
	idIndex := -1

	for i, header := range headers {
		if header == "sf__Id" {
			idIndex = i
			break
		}
	}

	if idIndex == -1 {
		return nil, fmt.Errorf("sf__Id not found in successfulResults headers")
	}

	csvRecords := [][]string{{"id"}}

	for _, row := range rows {
		csvRecords = append(csvRecords, []string{strings.Trim(row[idIndex], " ")})
	}

	return csvRecords, nil
}

func csvBytesToSlice(b []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewBuffer(b))

	records := make([][]string, 0)

	headers, err := reader.Read()
	if err != nil {
		return nil, nil, fmt.Errorf("error reading CSV headers: %w", err)
	}

	for {
		row, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, nil, fmt.Errorf("error reading CSV row: %w", err)
		}

		records = append(records, row)
	}

	return headers, records, nil
}
