package twenty

import (
	"errors"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const (
	// pageInfo.endCursor of read-companies-first-page.json, as Twenty returned it.
	firstPageEndCursor = "eyJpZCI6IjkyODBkMDMyLTk0ZjItNDc0OC05MzZmLTYyZTM0NTQwZWVmNyJ9"

	idGamma = "64f8acbb-ee41-4877-ab83-3a1f33931628"
	idBeta  = "9280d032-94f2-4748-936f-62e34540eef7"
	idAlpha = "c6f82506-260c-4ee5-a880-1cb4cce21178"
)

func TestRead(t *testing.T) { //nolint:funlen,maintidx
	t.Parallel()

	responseFirstPage := testutils.DataFromFile(t, "read-companies-first-page.json")
	responseLastPage := testutils.DataFromFile(t, "read-companies-last-page.json")
	responseEmpty := testutils.DataFromFile(t, "read-companies-empty.json")
	responseObjectNotFound := testutils.DataFromFile(t, "error-object-not-found.json")
	// Not captured live: Twenty did not return an empty page with more after it. It pins that Done follows
	// pageInfo.hasNextPage and not the number of rows.
	responseEmptyWithMore := []byte(`{"data":{"companies":[]},"totalCount":3,` +
		`"pageInfo":{"startCursor":null,"endCursor":"` + firstPageEndCursor + `",` +
		`"hasNextPage":true,"hasPreviousPage":false}}`)

	// Bounds equal to the updatedAt of two records of the first page.
	since := time.Date(2026, 10, 25, 13, 7, 55, 279_000_000, time.UTC) // Fixture Beta GmbH
	until := time.Date(2026, 10, 25, 13, 7, 55, 295_000_000, time.UTC) // Fixture Gamma SAS

	tests := []testconn.TestCaseRead{
		{
			Name:         "Read object must be included",
			Input:        common.ReadParams{Fields: connectors.Fields("id")},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:         "At least one field is requested",
			Input:        common.ReadParams{ObjectName: "companies"},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingFields},
		},
		{
			Name:  "Unknown object is a bad request, not retried",
			Input: common.ReadParams{ObjectName: "invoices", Fields: connectors.Fields("id")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusBadRequest, responseObjectNotFound),
			}.Server(),
			ExpectedErrs: []error{
				common.ErrBadRequest,
				testutils.StringError("object 'invoices' not found"),
			},
		},
		{
			Name:  "First page: depth=0, maximum page size, no fields parameter; cursor taken from pageInfo",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name", "domainName")},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.MethodGET(),
					mockcond.Path("/rest/companies"),
					mockcond.QueryParam("depth", "0"),
					mockcond.QueryParam("limit", "200"),
					mockcond.QueryParamsMissing("fields", "starting_after", "filter"),
				},
				Then: mockserver.Response(http.StatusOK, responseFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id: idGamma,
					Fields: map[string]any{
						"name": "Fixture Gamma SAS",
						"domainname": map[string]any{
							"primaryLinkLabel": "",
							"primaryLinkUrl":   "gamma.example.com",
							"secondaryLinks":   []any{},
						},
					},
					Raw: map[string]any{
						"id":        idGamma,
						"updatedAt": "2026-10-25T13:07:55.295Z",
						"address": map[string]any{
							"addressStreet1":  "",
							"addressStreet2":  "",
							"addressCity":     "Nantes",
							"addressPostcode": "",
							"addressState":    "",
							"addressCountry":  "",
							"addressLat":      nil,
							"addressLng":      nil,
						},
						// An inactive field's column is still returned, and Raw keeps it.
						"legacyRating": nil,
					},
				}, {
					Id: idBeta,
				}},
				NextPage: firstPageEndCursor,
				Done:     false,
			},
		},
		{
			Name: "Next page resumes at the cursor and ends when hasNextPage is false",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
				NextPage:   firstPageEndCursor,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/companies"),
					mockcond.QueryParam("starting_after", firstPageEndCursor),
				},
				Then: mockserver.Response(http.StatusOK, responseLastPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 1,
				Data: []common.ReadResultRow{{
					Id:     idAlpha,
					Fields: map[string]any{"name": "Fixture Alpha Ltd"},
					Raw:    map[string]any{"name": "Fixture Alpha Ltd"},
				}},
				NextPage: "",
				Done:     true,
			},
		},
		{
			Name:  "Page size is capped at Twenty's maximum of 200",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("id"), PageSize: 500},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If:    mockcond.QueryParam("limit", "200"),
				Then:  mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name: "Since and Until both go to Twenty, inclusive, at millisecond precision",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
				// Sub-millisecond parts are truncated: Twenty rejects more than three fractional digits.
				Since: since.Add(400 * time.Microsecond),
				Until: until.Add(999 * time.Microsecond),
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.Path("/rest/companies"),
					mockcond.QueryParam("filter",
						`updatedAt[gte]:"2026-10-25T13:07:55.279Z",updatedAt[lte]:"2026-10-25T13:07:55.295Z"`),
				},
				Then: mockserver.Response(http.StatusOK, responseFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				// Records stamped exactly at Since and at Until are both delivered: nothing is trimmed in code.
				Rows: 2,
				Data: []common.ReadResultRow{
					{
						Id:     idGamma,
						Fields: map[string]any{"name": "Fixture Gamma SAS"},
						Raw:    map[string]any{"updatedAt": "2026-10-25T13:07:55.295Z"},
					},
					{Id: idBeta, Raw: map[string]any{"updatedAt": "2026-10-25T13:07:55.279Z"}},
				},
				NextPage: firstPageEndCursor,
				Done:     false,
			},
		},
		{
			Name: "A request without updatedAt among its fields is still filtered, and Raw keeps updatedAt",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.And{
					mockcond.QueryParam("filter", `updatedAt[gte]:"2026-10-25T13:07:55.279Z"`),
					mockcond.QueryParamsMissing("fields"),
				},
				Then: mockserver.Response(http.StatusOK, responseFirstPage),
			}.Server(),
			Comparator: testconn.ComparatorSubsetRead,
			Expected: &common.ReadResult{
				Rows: 2,
				Data: []common.ReadResultRow{{
					Id:     idGamma,
					Fields: map[string]any{"name": "Fixture Gamma SAS"},
					Raw:    map[string]any{"updatedAt": "2026-10-25T13:07:55.295Z"},
				}, {
					Id: idBeta,
				}},
				NextPage: firstPageEndCursor,
				Done:     false,
			},
		},
		{
			Name: "Deleted reads the trash, with the time bounds",
			Input: common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
				Deleted:    true,
				Since:      since,
			},
			Server: mockserver.Conditional{
				Setup: mockserver.ContentJSON(),
				If: mockcond.QueryParam("filter",
					`deletedAt[is]:NOT_NULL,updatedAt[gte]:"2026-10-25T13:07:55.279Z"`),
				Then: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
		{
			Name:  "An empty page with hasNextPage true is not the end",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name"), Since: since},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmptyWithMore),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: firstPageEndCursor, Done: false},
		},
		{
			Name:  "An empty last page is the end",
			Input: common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name")},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, responseEmpty),
			}.Server(),
			Comparator: testconn.ComparatorPagination,
			Expected:   &common.ReadResult{Rows: 0, NextPage: "", Done: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableReader, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// The second call must send the NextPage the first call produced, and the two pages together must hold every
// record once.
func TestReadFollowsCursorAcrossPages(t *testing.T) {
	t.Parallel()

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If:   mockcond.QueryParam("starting_after", firstPageEndCursor),
			Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read-companies-last-page.json")),
		}, {
			If:   mockcond.QueryParamsMissing("starting_after"),
			Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read-companies-first-page.json")),
		}},
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	seen := map[string]int{}
	params := common.ReadParams{ObjectName: "companies", Fields: connectors.Fields("name"), PageSize: 2}
	done := false

	for range 3 {
		result, err := conn.Read(t.Context(), params)
		if err != nil {
			t.Fatal(err)
		}

		for _, row := range result.Data {
			seen[row.Id]++
		}

		if result.Done {
			done = true

			break
		}

		params.NextPage = result.NextPage
	}

	if !done {
		t.Fatal("the read did not reach Done")
	}

	for _, id := range []string{idGamma, idBeta, idAlpha} {
		if seen[id] != 1 {
			t.Errorf("record %s read %d times, want 1 (seen: %v)", id, seen[id], seen)
		}
	}

	if len(seen) != 3 {
		t.Errorf("read %d distinct records, want 3", len(seen))
	}
}

// Without `fields`, Twenty selects at most 200 fields on a wide object, and a morph relation inside the
// selection leaves a record with 199 keys (seen live). A page missing a field the schema has is read again
// with `fields` (requested fields in Twenty's case, a relation by its name, id, an unknown field left out),
// and the next page keeps sending it.
func TestReadSelectsFieldsBeyondDefaultSet(t *testing.T) { //nolint:funlen
	t.Parallel()

	record := `"id":"` + idAlpha + `","updatedAt":"2026-10-25T13:07:55.261Z"`
	for i := range 197 {
		record += `,"customField` + string(rune('A'+i/26)) + string(rune('a'+i%26)) + `":null`
	}

	pageInfo := func(endCursor string, hasNextPage bool) string {
		return `"pageInfo":{"startCursor":null,"endCursor":"` + endCursor + `","hasNextPage":` +
			strconv.FormatBool(hasNextPage) + `,"hasPreviousPage":false}`
	}

	defaultSet := []byte(`{"data":{"bigObjects":[{` + record + `}]},"totalCount":2,` + pageInfo(idAlpha, true) + `}`)
	selected := []byte(`{"data":{"bigObjects":[{"id":"` + idAlpha + `","vendorId":"` + idBeta + `",` +
		`"zzLastField":7}]},"totalCount":2,` + pageInfo(idAlpha, true) + `}`)
	lastSelected := []byte(`{"data":{"bigObjects":[{"id":"` + idGamma + `","vendorId":null,` +
		`"zzLastField":8}]},"totalCount":2,` + pageInfo(idGamma, false) + `}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If:   mockcond.Path("/rest/metadata/objects"),
			Then: mockserver.Response(http.StatusOK, bigObjectSchema),
		}, {
			If: mockcond.And{
				mockcond.Path("/rest/bigObjects"),
				mockcond.QueryParam("fields", "id,vendor,zzLastField"),
				mockcond.QueryParam("starting_after", idAlpha),
			},
			Then: mockserver.Response(http.StatusOK, lastSelected),
		}, {
			If: mockcond.And{
				mockcond.Path("/rest/bigObjects"),
				mockcond.QueryParam("fields", "id,vendor,zzLastField"),
				mockcond.QueryParamsMissing("starting_after"),
			},
			Then: mockserver.Response(http.StatusOK, selected),
		}, {
			If: mockcond.And{
				mockcond.Path("/rest/bigObjects"),
				mockcond.QueryParamsMissing("fields", "starting_after"),
			},
			Then: mockserver.Response(http.StatusOK, defaultSet),
		}},
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	params := common.ReadParams{
		ObjectName: "bigObjects",
		Fields:     connectors.Fields("zzlastfield", "vendorid", "typoField"),
	}

	first, err := conn.Read(t.Context(), params)
	if err != nil {
		t.Fatalf("a field beyond the default set must be read with fields, got %v", err)
	}

	if first.Rows != 1 || first.Data[0].Fields["zzlastfield"] != float64(7) || first.Data[0].Fields["vendorid"] != idBeta {
		t.Fatalf("first page fields %v, want zzlastfield 7 and vendorid %s", first.Data[0].Fields, idBeta)
	}

	if first.Done || first.NextPage != fieldsCursorPrefix+idAlpha {
		t.Fatalf("NextPage %q (Done %v), want the cursor marked to keep sending fields", first.NextPage, first.Done)
	}

	params.NextPage = first.NextPage

	last, err := conn.Read(t.Context(), params)
	if err != nil {
		t.Fatal(err)
	}

	if !last.Done || last.Rows != 1 || last.Data[0].Fields["zzlastfield"] != float64(8) {
		t.Fatalf("last page: Done %v, rows %v, want Done and zzlastfield 8", last.Done, last.Data)
	}
}

// With `fields`, Twenty returns every field it names; if a schema field were still missing, the read fails,
// naming it, as a provider error no retry fixes, not as the caller's. A field the schema does not know is
// only absent.
func TestReadRefusesSchemaFieldMissingFromRecord(t *testing.T) {
	t.Parallel()

	body := []byte(`{"data":{"bigObjects":[{"id":"` + idAlpha + `","updatedAt":"2026-10-25T13:07:55.261Z"}]},` +
		`"totalCount":1,"pageInfo":{"startCursor":null,"endCursor":null,"hasNextPage":false,"hasPreviousPage":false}}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If:   mockcond.Path("/rest/metadata/objects"),
			Then: mockserver.Response(http.StatusOK, bigObjectSchema),
		}, {
			If:   mockcond.Path("/rest/bigObjects"),
			Then: mockserver.Response(http.StatusOK, body),
		}},
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Read(t.Context(), common.ReadParams{
		ObjectName: "bigObjects",
		Fields:     connectors.Fields("customFieldAa", "zzLastField"),
	})
	if !errors.Is(err, errFieldNotReturned) || !errors.Is(err, common.ErrServerNonRetryable) {
		t.Fatalf("a schema field missing even with fields must fail the read, got %v", err)
	}

	if errors.Is(err, common.ErrCaller) {
		t.Errorf("the field is valid, so the error must not be a caller error: %v", err)
	}

	if _, err = conn.Read(t.Context(), common.ReadParams{
		ObjectName: "bigObjects",
		Fields:     connectors.Fields("id", "typoField"),
	}); err != nil {
		t.Fatalf("a field the schema does not know is only absent, got %v", err)
	}
}

// Twenty's default set is taken in name order, so it can keep renewalNotes and cut renewalnotes (seen live on
// v2.50.0). A request for renewalnotes on such a page is refused instead of answered with renewalNotes's value;
// a record too short to have been cut costs no metadata request.
func TestReadRefusesCaseCollisionAcrossDefaultSet(t *testing.T) { //nolint:funlen
	t.Parallel()

	record := `"id":"` + idAlpha + `","renewalNotes":"A"`
	for i := range 197 {
		record += `,"customField` + string(rune('A'+i/26)) + string(rune('a'+i%26)) + `":null`
	}

	wide := []byte(`{"data":{"vendorContracts":[{` + record + `}]},"totalCount":1,` +
		`"pageInfo":{"startCursor":null,"endCursor":null,"hasNextPage":false,"hasPreviousPage":false}}`)
	narrow := []byte(`{"data":{"vendorContracts":[{"id":"` + idAlpha + `","renewalNotes":"A"}]},"totalCount":1,` +
		`"pageInfo":{"startCursor":null,"endCursor":null,"hasNextPage":false,"hasPreviousPage":false}}`)
	schema := []byte(`{"data":[{"nameSingular":"vendorContract","namePlural":"vendorContracts",` +
		`"labelPlural":"Vendor contracts","isActive":true,"fields":[` +
		`{"id":"1","type":"TEXT","name":"renewalNotes","label":"A","isActive":true,"writability":"OPEN"},` +
		`{"id":"2","type":"TEXT","name":"renewalnotes","label":"B","isActive":true,"writability":"OPEN"},` +
		`{"id":"3","type":"NUMBER","name":"customFieldAa","label":"Aa","isActive":true,"writability":"OPEN"}]}],` +
		`"totalCount":1,"pageInfo":{"hasNextPage":false,"endCursor":"1"}}`)

	for _, tt := range []struct {
		name     string
		body     []byte
		metadata int
		fields   []string
		wantErr  error
	}{
		{"cut record, ambiguous field", wide, http.StatusOK, []string{"renewalnotes"}, errCaseCollision},
		{"cut record, unambiguous field", wide, http.StatusOK, []string{"customFieldAa"}, nil},
		{"short record, no metadata request", narrow, http.StatusInternalServerError, []string{"renewalNotes"}, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If:   mockcond.Path("/rest/metadata/objects"),
					Then: mockserver.Response(tt.metadata, schema),
				}, {
					If:   mockcond.Path("/rest/vendorContracts"),
					Then: mockserver.Response(http.StatusOK, tt.body),
				}},
			}.Server()
			defer server.Close()

			conn, err := constructTestConnector(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			_, err = conn.Read(t.Context(), common.ReadParams{
				ObjectName: "vendorContracts",
				Fields:     connectors.Fields(tt.fields...),
			})

			switch {
			case tt.wantErr == nil && err != nil:
				t.Fatalf("unexpected error %v", err)
			case tt.wantErr != nil && (!errors.Is(err, tt.wantErr) || !errors.Is(err, common.ErrCaller)):
				t.Fatalf("error %v, want a caller error %v", err, tt.wantErr)
			}
		})
	}
}

// bigObjectSchema describes bigObjects: two number fields and a many-to-one relation vendor, carried as vendorId.
var bigObjectSchema = []byte(`{"data":[{"nameSingular":"bigObject","namePlural":"bigObjects",` + //nolint:gochecknoglobals
	`"labelPlural":"Big objects","isActive":true,"fields":[` +
	`{"id":"1","type":"NUMBER","name":"customFieldAa","label":"Aa","isActive":true,"writability":"OPEN"},` +
	`{"id":"2","type":"NUMBER","name":"zzLastField","label":"Last","isActive":true,"writability":"OPEN"},` +
	`{"id":"3","type":"RELATION","name":"vendor","label":"Vendor","isActive":true,"writability":"OPEN",` +
	`"settings":{"relationType":"MANY_TO_ONE","joinColumnName":"vendorId"}}]}],` +
	`"totalCount":1,"pageInfo":{"hasNextPage":false,"endCursor":"1"}}`)

// Twenty allows renewalNotes next to renewalnotes. ReadResult.Fields would hold one of them as renewalnotes,
// so a request for either is refused instead of returning a value that may be the other's.
func TestReadRefusesCaseCollidingFields(t *testing.T) {
	t.Parallel()

	body := []byte(`{"data":{"vendorContracts":[{"id":"` + idAlpha + `","renewalNotes":"A","renewalnotes":"B",` +
		`"updatedAt":"2026-10-25T13:07:55.261Z"}]},"totalCount":1,` +
		`"pageInfo":{"startCursor":null,"endCursor":null,"hasNextPage":false,"hasPreviousPage":false}}`)

	server := mockserver.Fixed{
		Setup:  mockserver.ContentJSON(),
		Always: mockserver.Response(http.StatusOK, body),
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Read(t.Context(), common.ReadParams{
		ObjectName: "vendorContracts",
		Fields:     connectors.Fields("id", "renewalNotes"),
	})
	if !errors.Is(err, errCaseCollision) {
		t.Fatalf("expected a case collision error, got %v", err)
	}

	if _, err = conn.Read(t.Context(), common.ReadParams{
		ObjectName: "vendorContracts",
		Fields:     connectors.Fields("id", "updatedAt"),
	}); err != nil {
		t.Fatalf("fields without a collision must read, got %v", err)
	}
}

// Paging stops only on hasNextPage false: a response without it, or a cursor that does not move, fails.
func TestReadRefusesBrokenPaging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		body     string
		nextPage string
	}{
		{
			name: "pageInfo without hasNextPage",
			body: `{"data":{"companies":[]},"pageInfo":{"endCursor":null}}`,
		},
		{
			name:     "cursor equal to the one just sent",
			body:     `{"data":{"companies":[]},"pageInfo":{"hasNextPage":true,"endCursor":"` + firstPageEndCursor + `"}}`,
			nextPage: firstPageEndCursor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusOK, []byte(tt.body)),
			}.Server()
			defer server.Close()

			conn, err := constructTestConnector(server.URL)
			if err != nil {
				t.Fatal(err)
			}

			_, err = conn.Read(t.Context(), common.ReadParams{
				ObjectName: "companies",
				Fields:     connectors.Fields("name"),
				NextPage:   common.NextPageToken(tt.nextPage),
			})
			if err == nil {
				t.Fatal("expected an error, got none")
			}
		})
	}
}
