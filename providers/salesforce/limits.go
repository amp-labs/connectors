package salesforce

import (
	"context"

	"github.com/amp-labs/connectors/common"
)

func (c *Connector) Limits(ctx context.Context) (*LimitsResponse, error) {
	url, err := c.getRestApiURL("limits")
	if err != nil {
		return nil, err
	}

	response, err := c.JSON.Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	return common.UnmarshalJSON[LimitsResponse](response)
}

// nolint:tagliatelle
type LimitsResponse struct {
	ActiveScratchOrgs struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"ActiveScratchOrgs"`
	AnalyticsExternalDataSizeMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"AnalyticsExternalDataSizeMB"`
	ConcurrentAsyncGetReportInstances struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"ConcurrentAsyncGetReportInstances"`
	ConcurrentEinsteinDataInsightsStoryCreation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"ConcurrentEinsteinDataInsightsStoryCreation"`
	ConcurrentEinsteinDiscoveryStoryCreation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"ConcurrentEinsteinDiscoveryStoryCreation"`
	ConcurrentSyncReportRuns struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"ConcurrentSyncReportRuns"`
	DailyAnalyticsDataflowJobExecutions struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyAnalyticsDataflowJobExecutions"`
	DailyAnalyticsUploadedFilesSizeMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyAnalyticsUploadedFilesSizeMB"`
	DailyFunctionsAPICallLimit struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyFunctionsApiCallLimit"`
	DailyAPIRequests struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyApiRequests"`
	DailyAsyncApexExecutions struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyAsyncApexExecutions"`
	DailyAsyncApexTests struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyAsyncApexTests"`
	DailyBulkAPIBatches struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyBulkApiBatches"`
	DailyBulkV2QueryFileStorageMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyBulkV2QueryFileStorageMB"`
	DailyBulkV2QueryJobs struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyBulkV2QueryJobs"`
	DailyDeliveredPlatformEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyDeliveredPlatformEvents"`
	DailyDurableGenericStreamingAPIEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyDurableGenericStreamingApiEvents"`
	DailyDurableStreamingAPIEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyDurableStreamingApiEvents"`
	DailyEinsteinDataInsightsStoryCreation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyEinsteinDataInsightsStoryCreation"`
	DailyEinsteinDiscoveryPredictAPICalls struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyEinsteinDiscoveryPredictAPICalls"`
	DailyEinsteinDiscoveryPredictionsByCDC struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyEinsteinDiscoveryPredictionsByCDC"`
	DailyEinsteinDiscoveryStoryCreation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyEinsteinDiscoveryStoryCreation"`
	DailyGenericStreamingAPIEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyGenericStreamingApiEvents"`
	DailyScratchOrgs struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyScratchOrgs"`
	DailyStandardVolumePlatformEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyStandardVolumePlatformEvents"`
	DailyStreamingAPIEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyStreamingApiEvents"`
	DailyWorkflowEmails struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DailyWorkflowEmails"`
	DataStorageMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DataStorageMB"`
	DurableStreamingAPIConcurrentClients struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"DurableStreamingApiConcurrentClients"`
	FileStorageMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"FileStorageMB"`
	HourlyAsyncReportRuns struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyAsyncReportRuns"`
	HourlyDashboardRefreshes struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyDashboardRefreshes"`
	HourlyDashboardResults struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyDashboardResults"`
	HourlyDashboardStatuses struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyDashboardStatuses"`
	HourlyLongTermIDMapping struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyLongTermIdMapping"`
	HourlyManagedContentPublicRequests struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyManagedContentPublicRequests"`
	HourlyODataCallout struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyODataCallout"`
	HourlyPublishedPlatformEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyPublishedPlatformEvents"`
	HourlyPublishedStandardVolumePlatformEvents struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyPublishedStandardVolumePlatformEvents"`
	HourlyShortTermIDMapping struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyShortTermIdMapping"`
	HourlySyncReportRuns struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlySyncReportRuns"`
	HourlyTimeBasedWorkflow struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"HourlyTimeBasedWorkflow"`
	MassEmail struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"MassEmail"`
	MonthlyEinsteinDiscoveryStoryCreation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"MonthlyEinsteinDiscoveryStoryCreation"`
	Package2VersionCreates struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"Package2VersionCreates"`
	Package2VersionCreatesWithoutValidation struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"Package2VersionCreatesWithoutValidation"`
	PermissionSets struct {
		Max          int `json:"Max"`
		Remaining    int `json:"Remaining"`
		CreateCustom struct {
			Max       int `json:"Max"`
			Remaining int `json:"Remaining"`
		} `json:"CreateCustom"`
	} `json:"PermissionSets"`
	PlatformEventTriggersWithParallelProcessing struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"PlatformEventTriggersWithParallelProcessing"`
	PrivateConnectOutboundCalloutHourlyLimitMB struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"PrivateConnectOutboundCalloutHourlyLimitMB"`
	SingleEmail struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"SingleEmail"`
	StreamingAPIConcurrentClients struct {
		Max       int `json:"Max"`
		Remaining int `json:"Remaining"`
	} `json:"StreamingApiConcurrentClients"`
}
