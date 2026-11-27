package twenty

import (
	"bytes"
	"errors"
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testconn"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

const (
	vendorContractID = "fe3cd1c7-d4e6-4eca-ade0-d98a865ec2e6"
	alphaCompanyID   = "c6f82506-260c-4ee5-a880-1cb4cce21178"
)

// metadataCase answers the schema request every write makes first.
func metadataCase(t *testing.T) mockserver.Case {
	t.Helper()

	return mockserver.Case{
		If: mockcond.And{
			mockcond.MethodGET(),
			mockcond.Path("/rest/metadata/objects"),
		},
		Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "metadata-objects.json")),
	}
}

// metadataWithoutWritabilityCase answers with the same schema, minus the writability attribute.
func metadataWithoutWritabilityCase(t *testing.T) mockserver.Case {
	t.Helper()

	schema := bytes.ReplaceAll(testutils.DataFromFile(t, "metadata-objects.json"), []byte(`"writability": "OPEN",`), nil)

	return mockserver.Case{
		If:   mockcond.Path("/rest/metadata/objects"),
		Then: mockserver.Response(http.StatusOK, schema),
	}
}

func TestWrite(t *testing.T) { // nolint:funlen
	t.Parallel()

	responseCreate := testutils.DataFromFile(t, "write-vendor-contract-create.json")
	responseUpdate := testutils.DataFromFile(t, "write-vendor-contract-update.json")
	responseUnauthenticated := testutils.DataFromFile(t, "error-unauthenticated.json")

	tests := []testconn.TestCaseWrite{
		{
			Name:         "Write object must be included",
			Input:        common.WriteParams{RecordData: map[string]any{"name": "x"}},
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name: "Create sends Twenty's field names and returns the new id",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordData: map[string]any{
					"name":           "VC-001 Fixture",
					"renewalNotes":   "Renew in spring",
					"billingAddress": map[string]any{"addressCity": "Lyon"},
					"contractStatus": "DRAFT",
					"regions":        []any{"EUROPE"},
					"vendorId":       alphaCompanyID,
				},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataCase(t), {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/rest/vendorContracts"),
						mockcond.QueryParam("depth", "0"),
						mockcond.Body(`{"name":"VC-001 Fixture","renewalNotes":"Renew in spring",` +
							`"billingAddress":{"addressCity":"Lyon"},"contractStatus":"DRAFT","regions":["EUROPE"],` +
							`"vendorId":"` + alphaCompanyID + `"}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data: map[string]any{
					"id":           vendorContractID,
					"renewalNotes": "Renew in spring",
					"vendorId":     alphaCompanyID,
				},
			},
		},
		{
			Name: "Update patches the record by id",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordId:   vendorContractID,
				RecordData: map[string]any{"renewalNotes": "Renewal signed", "contractStatus": "SIGNED"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataCase(t), {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/rest/vendorContracts/" + vendorContractID),
						mockcond.Body(`{"renewalNotes":"Renewal signed","contractStatus":"SIGNED"}`),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data:     map[string]any{"renewalNotes": "Renewal signed", "contractStatus": "SIGNED"},
			},
		},
		{
			Name: "A schema without writability does not empty the payload",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordId:   vendorContractID,
				RecordData: map[string]any{"renewalnotes": "Renewal signed", "contractstatus": "SIGNED"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataWithoutWritabilityCase(t), {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Body(`{"renewalNotes":"Renewal signed","contractStatus":"SIGNED"}`),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data:     map[string]any{"renewalNotes": "Renewal signed"},
			},
		},
		{
			Name: "An inactive field keeps its Twenty name",
			Input: common.WriteParams{
				ObjectName: "companies",
				RecordId:   alphaCompanyID,
				RecordData: map[string]any{"legacyrating": 3},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataCase(t), {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Path("/rest/companies/" + alphaCompanyID),
						mockcond.Body(`{"legacyRating":3}`),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data:     map[string]any{"renewalNotes": "Renewal signed"},
			},
		},
		{
			Name: "An inactive object's payload is resolved too: managed fields dropped",
			Input: common.WriteParams{
				ObjectName: "archivedThings",
				RecordData: map[string]any{"name": "x", "updatedat": "2020-01-01T00:00:00.000Z"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataCase(t), {
					If: mockcond.And{
						mockcond.MethodPOST(),
						mockcond.Path("/rest/archivedThings"),
						mockcond.Body(`{"name":"x"}`),
					},
					Then: mockserver.Response(http.StatusCreated, responseCreate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data:     map[string]any{"renewalNotes": "Renew in spring"},
			},
		},
		{
			Name: "The legacy metadata envelope is read too",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordId:   vendorContractID,
				RecordData: map[string]any{"renewalnotes": "Renewal signed"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{{
					If: mockcond.And{
						mockcond.Path("/rest/metadata/objects"),
						mockcond.QueryParam("starting_after", "6762a14f-e73b-4ccb-a130-7641a1b20770"),
					},
					Then: mockserver.Response(http.StatusOK, legacyEnvelope(t, testutils.DataFromFile(t, "metadata-objects-page-2.json"))),
				}, {
					If:   mockcond.Path("/rest/metadata/objects"),
					Then: mockserver.Response(http.StatusOK, legacyEnvelope(t, testutils.DataFromFile(t, "metadata-objects-page-1.json"))),
				}, {
					If: mockcond.And{
						mockcond.MethodPATCH(),
						mockcond.Body(`{"renewalNotes":"Renewal signed"}`),
					},
					Then: mockserver.Response(http.StatusOK, responseUpdate),
				}},
			}.Server(),
			Comparator: testconn.ComparatorSubsetWrite,
			Expected: &common.WriteResult{
				Success:  true,
				RecordId: vendorContractID,
				Data:     map[string]any{"renewalNotes": "Renewal signed"},
			},
		},
		{
			Name: "A key given twice under different cases is refused before any write",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordData: map[string]any{"renewalNotes": "a", "renewalnotes": "b"},
			},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{metadataCase(t)},
			}.Server(),
			ExpectedErrs: []error{common.ErrCaller, errDuplicateField},
		},
		{
			Name: "The write fails when the schema cannot be read",
			Input: common.WriteParams{
				ObjectName: "vendorContracts",
				RecordData: map[string]any{"name": "x"},
			},
			Server: mockserver.Fixed{
				Setup:  mockserver.ContentJSON(),
				Always: mockserver.Response(http.StatusUnauthorized, responseUnauthenticated),
			}.Server(),
			ExpectedErrs: []error{common.ErrAccessToken},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (testconn.TestableWriter, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

// A record read, edited and written back must reach Twenty with Twenty's camelCase names (ReadResult.Fields
// keys are lowercase), with structured values kept whole, and without the fields Twenty manages: writing the
// updatedAt that was read would set it back and hide the edit from the next incremental read.
func TestReadFieldsRoundTripToWrite(t *testing.T) {
	t.Parallel()

	expectedPatch := `{"name":"VC-001 Fixture","renewalNotes":"Renewal signed","contractStatus":"SIGNED",` +
		`"billingAddress":{"addressStreet1":"1 Rue Exemple","addressStreet2":"","addressCity":"Paris",` +
		`"addressPostcode":"","addressState":"","addressCountry":"France","addressLat":null,"addressLng":null},` +
		`"regions":["EUROPE"],"vendorId":"` + alphaCompanyID + `"}`

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{metadataCase(t), {
			If: mockcond.And{
				mockcond.MethodGET(),
				mockcond.Path("/rest/vendorContracts"),
			},
			Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "read-vendor-contracts.json")),
		}, {
			If: mockcond.And{
				mockcond.MethodPATCH(),
				mockcond.Path("/rest/vendorContracts/" + vendorContractID),
				mockcond.Body(expectedPatch),
			},
			Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "write-vendor-contract-update.json")),
		}},
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	read, err := conn.Read(t.Context(), common.ReadParams{
		ObjectName: "vendorContracts",
		Fields: connectors.Fields("id", "name", "renewalNotes", "billingAddress", "contractStatus", "regions",
			"vendorId", "createdAt", "updatedAt", "deletedAt", "createdBy", "updatedBy", "position", "searchVector"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(read.Data) != 1 {
		t.Fatalf("read %d rows, want 1", len(read.Data))
	}

	record := read.Data[0].Fields
	if _, ok := record["renewalnotes"]; !ok {
		t.Fatalf("expected ReadResult.Fields to use lowercase keys, got %v", record)
	}

	record["renewalnotes"] = "Renewal signed"
	record["contractstatus"] = "SIGNED"
	address := record["billingaddress"].(map[string]any) // nolint:forcetypeassert
	address["addressCity"] = "Paris"

	written, err := conn.Write(t.Context(), common.WriteParams{
		ObjectName: "vendorContracts",
		RecordId:   read.Data[0].Id,
		RecordData: record,
	})
	if err != nil {
		t.Fatalf("write of the read record was not sent as expected: %v", err)
	}

	stored := written.Data["billingAddress"].(map[string]any) // nolint:forcetypeassert
	if written.Data["renewalNotes"] != "Renewal signed" || stored["addressCity"] != "Paris" ||
		stored["addressCountry"] != "France" {
		t.Errorf("response does not hold the written values: %v", written.Data)
	}
}

// Twenty allows two fields whose names differ only by case. A lowercase key, as ReadResult.Fields gives it,
// could stand for either, so it is refused and nothing is written; the exact camelCase name is accepted.
func TestWriteRefusesCaseCollidingKey(t *testing.T) {
	t.Parallel()

	schema := []byte(`{"data":[{"nameSingular":"vendorContract","namePlural":"vendorContracts",` +
		`"labelPlural":"Vendor contracts","isActive":true,"fields":[` +
		`{"id":"1","type":"TEXT","name":"renewalNotes","label":"Renewal notes","isActive":true,"writability":"OPEN"},` +
		`{"id":"2","type":"TEXT","name":"renewalnotes","label":"Renewal notes (lowercase)","isActive":true,` +
		`"writability":"OPEN"}]}],"totalCount":1,"pageInfo":{"hasNextPage":false,"endCursor":"1"}}`)

	server := mockserver.Switch{
		Setup: mockserver.ContentJSON(),
		Cases: []mockserver.Case{{
			If:   mockcond.Path("/rest/metadata/objects"),
			Then: mockserver.Response(http.StatusOK, schema),
		}, {
			If: mockcond.And{
				mockcond.MethodPATCH(),
				mockcond.Body(`{"renewalNotes":"A"}`),
			},
			Then: mockserver.Response(http.StatusOK, testutils.DataFromFile(t, "write-vendor-contract-update.json")),
		}},
	}.Server()
	defer server.Close()

	conn, err := constructTestConnector(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = conn.Write(t.Context(), common.WriteParams{
		ObjectName: "vendorContracts",
		RecordId:   vendorContractID,
		RecordData: map[string]any{"renewalnotes": "B"},
	})
	if !errors.Is(err, errCaseCollision) {
		t.Fatalf("a lowercase key matching two fields must be refused, got %v", err)
	}

	if _, err = conn.Write(t.Context(), common.WriteParams{
		ObjectName: "vendorContracts",
		RecordId:   vendorContractID,
		RecordData: map[string]any{"renewalNotes": "A"},
	}); err != nil {
		t.Fatalf("the exact camelCase name must be written, got %v", err)
	}
}
