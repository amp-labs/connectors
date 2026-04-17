package odoo

import (
	"net/http"
	"testing"

	"github.com/amp-labs/connectors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/goutils"
	"github.com/amp-labs/connectors/test/utils/mockutils"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockcond"
	"github.com/amp-labs/connectors/test/utils/mockutils/mockserver"
	"github.com/amp-labs/connectors/test/utils/testroutines"
	"github.com/amp-labs/connectors/test/utils/testutils"
)

func TestListObjectMetadata(t *testing.T) { //nolint:funlen
	t.Parallel()

	uomModel := testutils.DataFromFile(t, "uom-model.json")
	uomFields := testutils.DataFromFile(t, "uom-model-fields.json")
	calendarModel := testutils.DataFromFile(t, "calendar-model.json")
	calendarFields := testutils.DataFromFile(t, "calendar-model-fields.json")

	tests := []testroutines.Metadata{
		{
			Name:         "At least one object name must be queried",
			Input:        nil,
			Server:       mockserver.Dummy(),
			ExpectedErrs: []error{common.ErrMissingObjects},
		},
		{
			Name:  "Successfully describe uom.uom from ir.model and ir.model.fields search_read",
			Input: []string{"uom.uom"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, uomModel),
					},
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model.fields/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, uomFields),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"uom.uom": {
						DisplayName: "Product Unit of Measure",
						Fields: map[string]common.FieldMetadata{
							"active": {
								DisplayName:  "active",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"create_date": {
								DisplayName:  "create_date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"create_uid": {
								DisplayName:  "create_uid",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.users"},
							},
							"display_name": {
								DisplayName:  "display_name",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"factor": {
								DisplayName:  "factor",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"parent_path": {
								DisplayName:  "parent_path",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"related_uom_ids": {
								DisplayName:  "related_uom_ids",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"uom.uom"},
							},
							"relative_factor": {
								DisplayName:  "relative_factor",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"relative_uom_id": {
								DisplayName:  "relative_uom_id",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"uom.uom"},
							},
							"sequence": {
								DisplayName:  "sequence",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"write_date": {
								DisplayName:  "write_date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"write_uid": {
								DisplayName:  "write_uid",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.users"},
							},
						},
						FieldsMap: map[string]string{
							"active":          "active",
							"create_date":     "create_date",
							"create_uid":      "create_uid",
							"display_name":    "display_name",
							"factor":          "factor",
							"id":              "id",
							"name":            "name",
							"parent_path":     "parent_path",
							"related_uom_ids": "related_uom_ids",
							"relative_factor": "relative_factor",
							"relative_uom_id": "relative_uom_id",
							"sequence":        "sequence",
							"write_date":      "write_date",
							"write_uid":       "write_uid",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Successfully describe resource.calendar from fixtures",
			Input: []string{"resource.calendar"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, calendarModel),
					},
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model.fields/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, calendarFields),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Result: map[string]common.ObjectMetadata{
					"resource.calendar": {
						DisplayName: "Resource Working Time",
						Fields: map[string]common.FieldMetadata{
							"active": {
								DisplayName:  "active",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"attendance_ids": {
								DisplayName:  "attendance_ids",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"resource.calendar.attendance"},
							},
							"attendance_ids_1st_week": {
								DisplayName:  "attendance_ids_1st_week",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"resource.calendar.attendance"},
							},
							"attendance_ids_2nd_week": {
								DisplayName:  "attendance_ids_2nd_week",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"resource.calendar.attendance"},
							},
							"company_id": {
								DisplayName:  "company_id",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.company"},
							},
							"country_code": {
								DisplayName:  "country_code",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"country_id": {
								DisplayName:  "country_id",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.country"},
							},
							"create_date": {
								DisplayName:  "create_date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"create_uid": {
								DisplayName:  "create_uid",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.users"},
							},
							"display_name": {
								DisplayName:  "display_name",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"full_time_required_hours": {
								DisplayName:  "full_time_required_hours",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"global_leave_ids": {
								DisplayName:  "global_leave_ids",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"resource.calendar.leaves"},
							},
							"hours_per_day": {
								DisplayName:  "hours_per_day",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"hours_per_week": {
								DisplayName:  "hours_per_week",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"id": {
								DisplayName:  "id",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"is_fulltime": {
								DisplayName:  "is_fulltime",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"leave_ids": {
								DisplayName:  "leave_ids",
								ValueType:    common.ValueTypeReference,
								ProviderType: "one2many",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"resource.calendar.leaves"},
							},
							"name": {
								DisplayName:  "name",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(true),
							},
							"two_weeks_calendar": {
								DisplayName:  "two_weeks_calendar",
								ValueType:    common.ValueTypeBoolean,
								ProviderType: "boolean",
								ReadOnly:     goutils.Pointer(false),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"two_weeks_explanation": {
								DisplayName:  "two_weeks_explanation",
								ValueType:    common.ValueTypeString,
								ProviderType: "char",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"work_resources_count": {
								DisplayName:  "work_resources_count",
								ValueType:    common.ValueTypeInt,
								ProviderType: "integer",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"work_time_rate": {
								DisplayName:  "work_time_rate",
								ValueType:    common.ValueTypeFloat,
								ProviderType: "float",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"write_date": {
								DisplayName:  "write_date",
								ValueType:    common.ValueTypeDateTime,
								ProviderType: "datetime",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
							},
							"write_uid": {
								DisplayName:  "write_uid",
								ValueType:    common.ValueTypeReference,
								ProviderType: "many2one",
								ReadOnly:     goutils.Pointer(true),
								IsCustom:     goutils.Pointer(false),
								IsRequired:   goutils.Pointer(false),
								ReferenceTo:  []string{"res.users"},
							},
						},
						FieldsMap: map[string]string{
							"active":                   "active",
							"attendance_ids":           "attendance_ids",
							"attendance_ids_1st_week":  "attendance_ids_1st_week",
							"attendance_ids_2nd_week":  "attendance_ids_2nd_week",
							"company_id":               "company_id",
							"country_code":             "country_code",
							"country_id":               "country_id",
							"create_date":              "create_date",
							"create_uid":               "create_uid",
							"display_name":             "display_name",
							"full_time_required_hours": "full_time_required_hours",
							"global_leave_ids":         "global_leave_ids",
							"hours_per_day":            "hours_per_day",
							"hours_per_week":           "hours_per_week",
							"id":                       "id",
							"is_fulltime":              "is_fulltime",
							"leave_ids":                "leave_ids",
							"name":                     "name",
							"two_weeks_calendar":       "two_weeks_calendar",
							"two_weeks_explanation":    "two_weeks_explanation",
							"work_resources_count":     "work_resources_count",
							"work_time_rate":           "work_time_rate",
							"write_date":               "write_date",
							"write_uid":                "write_uid",
						},
					},
				},
				Errors: map[string]error{},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when ir.model search_read returns no rows",
			Input: []string{"uom.uom"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, []byte("[]")),
					},
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model.fields/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, uomFields),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"uom.uom": mockutils.ExpectedSubsetErrors{
						errNoIrModelRow,
					},
				},
			},
			ExpectedErrs: nil,
		},
		{
			Name:  "Returns error when ir.model responds with 500",
			Input: []string{"uom.uom"},
			Server: mockserver.Switch{
				Setup: mockserver.ContentJSON(),
				Cases: []mockserver.Case{
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusInternalServerError),
					},
					{
						If: mockcond.And{
							mockcond.Path("/json/2/ir.model.fields/search_read"),
							mockcond.MethodPOST(),
						},
						Then: mockserver.Response(http.StatusOK, uomFields),
					},
				},
			}.Server(),
			Comparator: testroutines.ComparatorSubsetMetadata,
			Expected: &common.ListObjectMetadataResult{
				Errors: map[string]error{
					"uom.uom": mockutils.ExpectedSubsetErrors{
						common.ErrServer,
					},
				},
			},
			ExpectedErrs: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			t.Parallel()

			tt.Run(t, func() (connectors.ObjectMetadataConnector, error) {
				return constructTestConnector(tt.Server.URL)
			})
		})
	}
}

func constructTestConnector(serverURL string) (*Connector, error) {
	connector, err := NewConnector(common.ConnectorParams{
		AuthenticatedClient: mockutils.NewClient(),
		Metadata: map[string]string{
			"odoo_domain": "mock.example.com",
		},
	})
	if err != nil {
		return nil, err
	}

	connector.SetBaseURL(mockutils.ReplaceURLOrigin(connector.HTTPClient().Base, serverURL))

	return connector, nil
}
