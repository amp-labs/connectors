package salesforce

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
)

type ListIngestJobsResult struct {
	Done           bool               `json:"done"`
	NextRecordsURL string             `json:"nextRecordsUrl"`
	Records        []GetJobInfoResult `json:"records"`
}

// ListIngestJobsInfo returns information about Ingest Jobs. If jobIds are provided, only those jobs are returned.
// It is possible to get information about all current ingest jobs by not providing any jobIds. Note that Salesforce
// returns terminal state jobs from a maximum of 7 days ago.
func (c *Connector) ListIngestJobsInfo(ctx context.Context, jobIds ...string) ([]GetJobInfoResult, error) {
	url, err := c.getRestApiURL("jobs/ingest")
	if err != nil {
		return nil, err
	}

	// If we have jobIds, we need to keep track of which ones have been matched, so we can break early
	var jobMatched map[string]bool
	if len(jobIds) > 0 {
		jobMatched = make(map[string]bool)
		for _, id := range jobIds {
			jobMatched[id] = false
		}
	}

	// Filters out BigObjects / BulkAPI v1 jobs
	url.WithQueryParam("jobType", "V2Ingest")

	// To keep track of pages
	location := url.String()

	// Collect all jobs
	var jobsInfo []GetJobInfoResult

	for {
		res, err := c.Client.Get(ctx, location)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to list ingest jobs info: %w",
				errors.Join(err, common.ErrRequestFailed),
			)
		}

		result, err := common.UnmarshalJSON[ListIngestJobsResult](res)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal list ingest jobs info: %w",
				errors.Join(err, common.ErrParseError),
			)
		}

		for _, job := range result.Records {
			// If there's no filter, or the job is in the filter, add it to the list
			if len(jobIds) == 0 || contains[string](jobIds, job.Id) {
				jobsInfo = append(jobsInfo, job)

				// Match it if we have a filter to help break early
				if len(jobIds) > 0 {
					jobMatched[job.Id] = true
				}
			}
		}

		if !result.Done {
			// If we have a filter, and all jobs have been matched, we can break early
			if len(jobIds) > 0 {
				allMatched := true

				for _, matched := range jobMatched {
					if !matched {
						allMatched = false
						break
					}
				}

				if allMatched {
					break
				}
			}

			// Else, get the next page
			domain, err := c.getDomainURL()
			if err != nil {
				return nil, err
			}

			// getDomainURL escapes some of the characters in the nextRecordsURL, using the URL directly
			location = domain.String() + result.NextRecordsURL
		} else {
			// If we're done, break
			break
		}
	}

	return jobsInfo, nil
}

// GetBulkQueryInfo returns information status about a Query Job,
// which was created via BulkRead or BulkQuery.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/query_get_one_job.htm
func (c *Connector) GetBulkQueryInfo(ctx context.Context, jobId string) (*GetJobInfoResult, error) {
	location, err := c.getRestApiURL("jobs/query", jobId)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, location.String())
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get bulk query info for job '%s': %w",
			jobId,
			errors.Join(err, common.ErrRequestFailed),
		)
	}

	return common.UnmarshalJSON[GetJobInfoResult](res)
}

// GetBulkQueryResults returns completed data from bulk query.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/query_get_job_results.htm
func (c *Connector) GetBulkQueryResults(ctx context.Context, jobId string) (*http.Response, error) {
	location, err := c.getRestApiURL("jobs/query/", jobId, "/results")
	if err != nil {
		return nil, err
	}

	req, err := common.MakeJSONGetRequest(ctx, location.String(), []common.Header{
		{
			Key:   "Accept",
			Value: "text/csv",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get results for bulk query %s: %w", jobId, err)
	}

	// Get the connector's JSONHTTPClient, which is a special HTTPClient that handles JSON responses,
	// and use it's underlying http.Client to make the request.
	return c.Client.HTTPClient.Client.Do(req)
}

// GetJobInfo returns information status about an Ingest Job,
// which was created via BulkWrite or BulkDelete.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/get_job_info.htm
func (c *Connector) GetJobInfo(ctx context.Context, jobId string) (*GetJobInfoResult, error) {
	location, err := c.getRestApiURL("jobs/ingest", jobId)
	if err != nil {
		return nil, err
	}

	rsp, err := c.Client.Get(ctx, location.String())
	if err != nil {
		return nil, fmt.Errorf("getGetInfo failed: %w", errors.Join(err, common.ErrRequestFailed))
	}

	info, err := common.UnmarshalJSON[GetJobInfoResult](rsp)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling result of getGetInfo failed: %w", errors.Join(err, common.ErrParseError))
	}

	return info, nil
}

// GetJobResults returns explanation on Ingest Job status.
// In case of success, only metadata marking such state is returned.
// In case of failure, reasons for error are collected and returned. For more on failures refer to the docs below.
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/get_job_failed_results.htm
func (c *Connector) GetJobResults(ctx context.Context, jobId string) (*JobResults, error) {
	jobInfo, err := c.GetJobInfo(ctx, jobId)
	if err != nil {
		return nil, fmt.Errorf("failed to get job information: %w", err)
	}

	if jobInfo.State != JobStateComplete {
		// Take care of failed, aborted, in progress, and upload complete cases
		// We don't need to query Salesforce for these cases
		return &JobResults{
			JobId:   jobInfo.Id,
			State:   jobInfo.State,
			JobInfo: jobInfo,
			Message: getIncompleteJobMessage(jobInfo),
		}, nil
	}

	if jobInfo.State == JobStateComplete && jobInfo.NumberRecordsFailed == 0 {
		// Complete success case, no need to query Salesforce
		return &JobResults{
			JobId:   jobInfo.Id,
			State:   jobInfo.State,
			JobInfo: jobInfo,
		}, nil
	}

	return c.getPartialFailureDetails(ctx, jobInfo)
}

// GetSuccessfulJobResults returns completed data from ingest job.
// If you know that Job was successful from running "info" methods, by calling this method
// you can get record results for this operation (write or delete).
// https://developer.salesforce.com/docs/atlas.en-us.api_asynch.meta/api_asynch/get_job_successful_results.htm
func (c *Connector) GetSuccessfulJobResults(ctx context.Context, jobId string) (*http.Response, error) {
	location, err := c.getRestApiURL("jobs/ingest", jobId, "successfulResults")
	if err != nil {
		return nil, err
	}

	req, err := common.MakeJSONGetRequest(ctx, location.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get request: %w", err)
	}

	// Get the connector's JSONHTTPClient, which is a special HTTPClient that handles JSON responses,
	// and use it's underlying http.Client to make the request.
	return c.Client.HTTPClient.Client.Do(req)
}

// contains is a quick helper function to check if a slice contains a specific element.
func contains[T comparable](s []T, e T) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}

	return false
}

// all is a quick helper function to check if all elements in a slice satisfy a predicate.
func all[T any](s []T, f func(T) bool) bool {
	for _, a := range s {
		if !f(a) {
			return false
		}
	}

	return true
}
