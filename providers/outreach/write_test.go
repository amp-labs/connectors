package outreach

import (
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

// nolint
func TestWrite(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	sequencesResponse := testutils.DataFromFile(t, "sequences.json")
	unsupportedResponse := testutils.DataFromFile(t, "unsupported.json")
	updateDealsResponse := testutils.DataFromFile(t, "emailaddress.json")

	tests := []testroutines.Write{
		{
			Name:         "Object Name is required",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "RecordData is required",
			Input:        common.WriteParams{ObjectName: "leads"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingRecordData},
		},

		{
			Name:  "Unsupported object",
			Input: common.WriteParams{ObjectName: "arsenal", RecordData: map[string]any{"test": "value"}},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusNotFound, unsupportedResponse),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrRetryable,
				errors.New(string(unsupportedResponse)),
			},
		},
		{
			Name:         "RecordID must be convertible to integers",
			Input:        common.WriteParams{ObjectName: "prospects", RecordData: map[string]any{}, RecordId: "xseedesrt"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{ErrIdMustInt},
		},
		{
			Name: "Successful creation of a sequence",
			Input: common.WriteParams{
				ObjectName: "sequences",
				RecordData: map[string]any{
					"description": "A test sequence",
					"type":        "sequence",
					"name":        "string",
					"tags":        []string{"sequence", "coffee"},
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPOST(),
				Then:  mockserver.Response(http.StatusOK, sequencesResponse),
			}.Server(),
			Comparator: testroutines.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "18",
				Data: map[string]any{
					"attributes": map[string]any{
						"automationPercentage":        float64(0),
						"bounceCount":                 float64(0),
						"clickCount":                  float64(0),
						"createdAt":                   "2025-01-13T12:47:31.000Z",
						"deliverCount":                float64(0),
						"description":                 "A test sequence",
						"durationInDays":              float64(0),
						"enabled":                     false,
						"enabledAt":                   nil,
						"engagementScore":             nil,
						"failureCount":                float64(0),
						"finishOnReply":               true,
						"interestedCount":             nil,
						"lastUsedAt":                  "2025-01-13T12:47:31.000Z",
						"locked":                      false,
						"lockedAt":                    nil,
						"maxActivations":              float64(150),
						"name":                        "string",
						"negativeReplyCount":          float64(0),
						"neutralReplyCount":           float64(0),
						"numContactedProspects":       float64(0),
						"numRepliedProspects":         float64(0),
						"openCount":                   float64(0),
						"optOutCount":                 float64(0),
						"positiveReplyCount":          float64(0),
						"primaryReplyAction":          "finish",
						"primaryReplyPauseDuration":   nil,
						"replyCount":                  float64(0),
						"salesMotion":                 nil,
						"scheduleCount":               float64(0),
						"scheduleIntervalType":        "calendar",
						"secondaryReplyAction":        "finish",
						"secondaryReplyPauseDuration": nil,
						"sequenceStepCount":           float64(0),
						"sequenceType":                "interval",
						"shareType":                   "shared",
						"tags": []any{
							"sequence",
							"coffee",
						},
						"throttleCapacity":      nil,
						"throttleMaxAddsPerDay": nil,
						"throttlePaused":        false,
						"throttlePausedAt":      nil,
						"transactional":         false,
						"updatedAt":             "2025-01-13T12:47:31.000Z",
					},
					"id": float64(18),
					"links": map[string]any{
						"self": "https://api.outreach.io/api/v2/sequences/18",
					},

					"type": "sequence",
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name: "Successfully update an EmailAddress",
			Input: common.WriteParams{
				ObjectName: "emailAddresses",
				RecordId:   "5",
				RecordData: map[string]any{
					"email": "groverstiedemann@lehner.io",
					"order": 6,
				},
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.MethodPATCH(),
				Then:  mockserver.Response(http.StatusOK, updateDealsResponse),
			}.Server(),
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: "5",
				Data: map[string]any{
					"attributes": map[string]any{
						"bouncedAt":       nil,
						"createdAt":       "2024-04-28T09:07:09.000Z",
						"email":           "groverstiedemann@lehner.io",
						"emailType":       nil,
						"order":           float64(6),
						"status":          nil,
						"statusChangedAt": nil,
						"unsubscribedAt":  nil,
						"updatedAt":       "2025-01-13T12:34:39.000Z",
					},
					"id": float64(5),
					"links": map[string]any{
						"self": "https://api.outreach.io/api/v2/emailAddresses/5",
					},
					"relationships": map[string]any{
						"prospect": map[string]any{
							"data": nil,
						},
					},
					"type": "emailAddress",
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		// nolint:varnamelen
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.WriteConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
