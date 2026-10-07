package wrike

import (
	"bytes"
	"context"
	"encoding/json"
	"maps"
	"net/http"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
)

const (
	apiPrefix         = "/v4"
	recordPlaceholder = "{record_id}"

	fieldID = "id"
)

// writeSpec is the verified per-object write behaviour.
type writeSpec struct {
	// createPath is the POST path; empty means the object cannot be created.
	createPath string
	// updatePath is the PUT path with {record_id}; empty means no update.
	updatePath string
	// idField is the response field returned as WriteResult.RecordId.
	idField string
}

// writeSpecs documents, per object, the create and update endpoints from the
// Wrike OpenAPI spec.
//
// Payloads are sent as JSON (verified live for POST and PUT, including array
// and object parameters). Wrike rejects any parameter its endpoint does not
// declare, including read-only fields such as id or createdDate, with
// 400 invalid_request; only the record id is stripped on update.
//
// Tasks, folders, comments, approvals, attachments, bookings and timelogs have
// no account-level create in Wrike: they are created under a parent
// (POST /folders/{id}/tasks, POST /tasks/{id}/comments, ...). The connector
// API carries no parent identifier, so those objects are update-only here.
//
// Not writable: access_roles, audit_log, custom_item_types, folder_blueprints,
// placeholders, request_forms, task_blueprints, timelog_categories,
// timesheet_submission_rules, user_types (no per-record create or update).
//
//nolint:gochecknoglobals
var writeSpecs = map[string]writeSpec{
	// Update only: created under a parent in Wrike.
	objectTasks:       {updatePath: "/tasks/{record_id}"},
	objectFolders:     {updatePath: "/folders/{record_id}"},
	objectComments:    {updatePath: "/comments/{record_id}"},
	objectApprovals:   {updatePath: "/approvals/{record_id}"},
	objectAttachments: {updatePath: "/attachments/{record_id}"},
	objectBookings:    {updatePath: "/bookings/{record_id}"},
	objectTimelogs:    {updatePath: "/timelogs/{record_id}"},

	// Account-level create and update.
	objectCustomFields: {createPath: "/customfields", updatePath: "/customfields/{record_id}"},
	objectGroups:       {createPath: "/groups", updatePath: "/groups/{record_id}"},
	objectInvitations:  {createPath: "/invitations", updatePath: "/invitations/{record_id}"},
	objectJobRoles:     {createPath: "/jobroles", updatePath: "/jobroles/{record_id}"},
	objectSpaces:       {createPath: "/spaces", updatePath: "/spaces/{record_id}"},
	objectTimesheets: {
		createPath: "/timesheets", updatePath: "/timesheets/{record_id}", idField: "timesheetId",
	},
	objectUserScheduleExclusions: {
		createPath: "/user_schedule_exclusions", updatePath: "/user_schedule_exclusions/{record_id}",
	},
	objectWorkflows:     {createPath: "/workflows", updatePath: "/workflows/{record_id}"},
	objectWorkschedules: {createPath: "/workschedules", updatePath: "/workschedules/{record_id}"},

	// Update only.
	objectContacts: {updatePath: "/contacts/{record_id}"},
}

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	spec, ok := writeSpecs[params.ObjectName]
	if !ok {
		return nil, common.ErrOperationNotSupportedForObject
	}

	record, err := params.GetRecord()
	if err != nil {
		return nil, err
	}

	// Work on a copy: GetRecord hands back the caller's own map.
	record = maps.Clone(record)

	path, method, err := resolveWritePath(params, spec, record)
	if err != nil {
		return nil, err
	}

	url, err := urlbuilder.New(c.ProviderInfo().BaseURL, apiPrefix+path)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// resolveWritePath picks the endpoint and method and strips the record id,
// which travels in the path, from an update payload.
func resolveWritePath(params common.WriteParams, spec writeSpec, record common.Record) (string, string, error) {
	if params.IsUpdate() {
		if spec.updatePath == "" {
			return "", "", common.ErrOperationNotSupportedForObject
		}

		// Wrike rejects "id" as a parameter; read records carry it.
		delete(record, fieldID)

		return strings.ReplaceAll(spec.updatePath, recordPlaceholder, params.RecordId), http.MethodPut, nil
	}

	if spec.createPath == "" {
		return "", "", common.ErrOperationNotSupportedForObject
	}

	return spec.createPath, http.MethodPost, nil
}

// parseWriteResponse reads the written record from Wrike's {kind, data:[record]} envelope.
func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok || body == nil {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	records, err := jsonquery.New(body).ArrayOptional("data")
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return &common.WriteResult{Success: true, RecordId: params.RecordId}, nil
	}

	idField := writeSpecs[params.ObjectName].idField
	if idField == "" {
		idField = fieldID
	}

	recordID, err := jsonquery.New(records[0]).TextWithDefault(idField, "")
	if err != nil {
		return nil, err
	}

	if recordID == "" {
		recordID = params.RecordId
	}

	data, err := jsonquery.Convertor.ObjectToMap(records[0])
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
