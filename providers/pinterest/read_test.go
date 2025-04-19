package pinterest

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestRead(t *testing.T) { // nolint:funlen,gocognit,cyclop
	t.Parallel()

	pinsResponse := testutils.DataFromFile(t, "pins.json")
	boardsResponse := testutils.DataFromFile(t, "boards.json")
	mediaResponse := testutils.DataFromFile(t, "media.json")

	tests := []testroutines.Read{
		{
			Name:         "Read object must be included",
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Read list of all pins",
			Input: common.ReadParams{ObjectName: "pins", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v5/pins"),
				Then:  mockserver.Response(http.StatusOK, pinsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":               "813744226420795884",
							"created_at":       "2020-01-01T20:10:40Z",
							"link":             "https://www.pinterest.com/",
							"title":            "string",
							"description":      "string",
							"dominant_color":   "#6E7874",
							"alt_text":         "string",
							"creative_type":    "REGULAR",
							"board_id":         "string",
							"board_section_id": "string",
							"board_owner": map[string]any{
								"username": "string",
							},
							"is_owner": false,
							"media": map[string]any{
								"media_type": "string",
								"images": map[string]any{
									"150x150": map[string]any{
										"width":  float64(150),
										"height": float64(150),
										"url":    "https://i.pinimg.com/150x150/0d/f6/f1/0df6f1f0bfe7aaca849c1bbc3607a34b.jpg",
									},
									"400x300": map[string]any{
										"width":  float64(400),
										"height": float64(300),
										"url":    "https://i.pinimg.com/400x300/0d/f6/f1/0df6f1f0bfe7aaca849c1bbc3607a34b.jpg",
									},
									"600x": map[string]any{
										"width":  float64(600),
										"height": float64(600),
										"url":    "https://i.pinimg.com/600x/0d/f6/f1/0df6f1f0bfe7aaca849c1bbc3607a34b.jpg",
									},
									"1200x": map[string]any{
										"width":  float64(1200),
										"height": float64(1200),
										"url":    "https://i.pinimg.com/1200x/0d/f6/f1/0df6f1f0bfe7aaca849c1bbc3607a34b.jpg",
									},
								},
							},
							"parent_pin_id":     "string",
							"is_standard":       false,
							"has_been_promoted": false,
							"note":              "string",
							"pin_metrics": map[string]any{
								"90d": map[string]any{
									"pin_click":    float64(7),
									"impression":   float64(2),
									"clickthrough": float64(3),
								},
								"lifetime_metrics": map[string]any{
									"pin_click":    float64(7),
									"impression":   float64(2),
									"clickthrough": float64(3),
									"reaction":     float64(10),
									"comment":      float64(2),
								},
							},
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v5/pins?bookmark=P2MxMDU0MzM0OTYyNzU0OTg0NjUzLjE3MTUzMzgzMDksLTF8NHwxMjE0NzA2NjAzMTA1Mz&page_size=250", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all boards",
			Input: common.ReadParams{ObjectName: "boards", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v5/boards"),
				Then:  mockserver.Response(http.StatusOK, boardsResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"id":                     "549755885175",
							"created_at":             "2020-01-01T20:10:40Z",
							"board_pins_modified_at": "2020-01-01T20:10:40Z",
							"name":                   "Summer Recipes",
							"description":            "My favorite summer recipes",
							"collaborator_count":     float64(17),
							"pin_count":              float64(5),
							"follower_count":         float64(13),
							"media": map[string]any{
								"image_cover_url": "https://i.pinimg.com/400x300/fd/cd/d5/fdcdd5a6d8a80824add0d054125cd957.jpg",
								"pin_thumbnail_urls": []any{
									"https://i.pinimg.com/150x150/b4/57/10/b45710f1ede96af55230f4b43935c4af.jpg",
									"https://i.pinimg.com/150x150/dd/ff/46/ddff4616e39c1935cd05738794fa860e.jpg",
									"https://i.pinimg.com/150x150/84/ac/59/84ac59b670ccb5b903dace480a98930c.jpg",
									"https://i.pinimg.com/150x150/4c/54/6f/4c546f521be85e30838fb742bfff6936.jpg",
								},
							},
							"owner": map[string]any{
								"username": "string",
							},
							"privacy": "PUBLIC",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v5/boards?bookmark=P2MxMDU0MzM0OTYyNzU0OTg0NjUzLjE3MTUzMzgzMDksLTF8NHwxMjE0NzA2NjAzMTA1Mz&page_size=250", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Read list of all media",
			Input: common.ReadParams{ObjectName: "media", Fields: connectors.Fields("")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.PathSuffix("/v5/media"),
				Then:  mockserver.Response(http.StatusOK, mediaResponse),
			}.Server(),
			Comparator: testroutines.ComparatorPagination,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{
					{
						Fields: map[string]any{},
						Raw: map[string]any{
							"media_id":   "12345",
							"media_type": "video",
							"status":     "succeeded",
						},
					},
				},
				NextPage: testroutines.URLTestServer + "/v5/media?bookmark=P2MxMDU0MzM0OTYyNzU0OTg0NjUzLjE3MTUzMzgzMDksLTF8NHwxMjE0NzA2NjAzMTA1Mz&page_size=250", // nolint:lll
				Done:     false,
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ReadConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}
