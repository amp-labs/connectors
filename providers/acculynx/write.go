package acculynx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/providers/acculynx/metadata"
)

// AccuLynx write API references (all JSON; file-upload endpoints are excluded
// — see metadata/README.md for the multipart exclusion rationale):
//
//   Create contact:                https://apidocs.acculynx.com/reference/postcontact
//   Create job:                    https://apidocs.acculynx.com/reference/postjobs
//   Update contact CF value:       https://apidocs.acculynx.com/reference/putcontactcustomfields
//   Update job CF value:           https://apidocs.acculynx.com/reference/putjobcustomfield
//   Set job AR/Sales/Company rep:  https://apidocs.acculynx.com/reference/postjobsarrepresentative
//   Update initial appointment:    https://apidocs.acculynx.com/reference/putinitialappointmentforjob
//   Set job insurance company:     https://apidocs.acculynx.com/reference/putjobsinsurance
//   Add contact phone number:      https://apidocs.acculynx.com/reference/postcontactphonenumber
//   Add contact log:               https://apidocs.acculynx.com/reference/postcontactlog
//   Create external job reference: https://apidocs.acculynx.com/reference/postjobsexternalreferences
//   Send job message:              https://apidocs.acculynx.com/reference/postjobmessage
//   Reply to job message:          https://apidocs.acculynx.com/reference/postjobmessagereply
//   Add payment (expense/paid/received):
//     https://apidocs.acculynx.com/reference/postjobsadditionalexpense
//     https://apidocs.acculynx.com/reference/postjobspaymentpaid
//     https://apidocs.acculynx.com/reference/postjobspaymentreceived

// Path-template parent-id keys carried in RecordData for nested writes.
const (
	parentIDKeyJobID     = "jobId"
	parentIDKeyContactID = "contactId"
	parentIDKeyMessageID = "messageId"
)

var errMissingParentID = errors.New("acculynx: missing parent id in record data")

func (c *Connector) buildWriteRequest(ctx context.Context, params common.WriteParams) (*http.Request, error) {
	if err := params.ValidateParams(); err != nil {
		return nil, err
	}

	if err := validateWriteParams(params); err != nil {
		return nil, err
	}

	record, err := common.RecordDataToMap(params.RecordData)
	if err != nil {
		return nil, err
	}

	url, method, err := c.buildWriteURL(params, record)
	if err != nil {
		return nil, err
	}

	body, err := json.Marshal(record)
	if err != nil {
		return nil, fmt.Errorf("marshal record data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// validateWriteParams rejects unsupported objects and operations early.
// AccuLynx exposes no top-level updates and no general deletes.
//
//nolint:cyclop
func validateWriteParams(params common.WriteParams) error {
	switch params.ObjectName {
	case objectContacts, objectJobs:
		if params.IsUpdate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	case "contacts/custom-fields", "jobs/custom-fields":
		if params.IsCreate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	case
		// Job-singleton PUTs (set a slot on a job).
		"jobs/initial-appointment",
		"jobs/insurance/insurance-company",
		// Job-rep POSTs (assign someone to a slot).
		"jobs/representatives/ar-owner",
		"jobs/representatives/sales-owner",
		"jobs/representatives/company",
		// Contact sub-resource POSTs.
		"contacts/phone-numbers",
		"contacts/logs",
		// Job sub-resource POSTs.
		"jobs/external-references",
		"jobs/messages",
		"jobs/messages/replies",
		"jobs/payments/expense",
		"jobs/payments/paid",
		"jobs/payments/received":
		if params.IsUpdate() {
			return common.ErrOperationNotSupportedForObject
		}

		return nil
	default:
		return common.ErrOperationNotSupportedForObject
	}
}

//nolint:cyclop,funlen
func (c *Connector) buildWriteURL(
	params common.WriteParams, record map[string]any,
) (*urlbuilder.URL, string, error) {
	baseURL := c.ProviderInfo().BaseURL

	switch params.ObjectName {
	// Top-level POSTs.
	case objectContacts:
		u, err := urlbuilder.New(baseURL, c.modulePath(), "contacts")

		return u, http.MethodPost, err

	case objectJobs:
		u, err := urlbuilder.New(baseURL, c.modulePath(), "jobs")

		return u, http.MethodPost, err

	case "jobs/external-references":
		u, err := urlbuilder.New(baseURL, c.modulePath(), "jobs", "external-references")

		return u, http.MethodPost, err

	// Contact-nested PUTs/POSTs.
	case "contacts/custom-fields":
		u, err := c.buildContactNestedURL(record, "custom-fields", params.RecordId)

		return u, http.MethodPut, err

	case "contacts/phone-numbers":
		u, err := c.buildContactNestedURL(record, "phone-numbers")

		return u, http.MethodPost, err

	case "contacts/logs":
		u, err := c.buildContactNestedURL(record, "logs")

		return u, http.MethodPost, err

	// Job-nested PUTs.
	case "jobs/custom-fields":
		u, err := c.buildJobNestedURL(record, "custom-fields", params.RecordId)

		return u, http.MethodPut, err

	case "jobs/initial-appointment":
		u, err := c.buildJobNestedURL(record, "initial-appointment")

		return u, http.MethodPut, err

	case "jobs/insurance/insurance-company":
		u, err := c.buildJobNestedURL(record, "insurance", "insurance-company")

		return u, http.MethodPut, err

	// Job-nested POSTs (representatives + messages + payments).
	case "jobs/representatives/ar-owner":
		u, err := c.buildJobNestedURL(record, "representatives", "ar-owner")

		return u, http.MethodPost, err

	case "jobs/representatives/sales-owner":
		u, err := c.buildJobNestedURL(record, "representatives", "sales-owner")

		return u, http.MethodPost, err

	case "jobs/representatives/company":
		u, err := c.buildJobNestedURL(record, "representatives", "company")

		return u, http.MethodPost, err

	case "jobs/messages":
		u, err := c.buildJobNestedURL(record, "messages")

		return u, http.MethodPost, err

	case "jobs/messages/replies":
		messageID, ok := stringFromRecord(record, parentIDKeyMessageID)
		if !ok {
			return nil, "", fmt.Errorf("%w: %s", errMissingParentID, parentIDKeyMessageID)
		}

		u, err := c.buildJobNestedURL(record, "messages", messageID, "replies")

		return u, http.MethodPost, err

	case "jobs/payments/expense":
		u, err := c.buildJobNestedURL(record, "payments", "expense")

		return u, http.MethodPost, err

	case "jobs/payments/paid":
		u, err := c.buildJobNestedURL(record, "payments", "paid")

		return u, http.MethodPost, err

	case "jobs/payments/received":
		u, err := c.buildJobNestedURL(record, "payments", "received")

		return u, http.MethodPost, err

	default:
		return nil, "", common.ErrOperationNotSupportedForObject
	}
}

// stringFromRecord pulls a parent-id from RecordData and removes it from the
// map so it is not echoed back into the request body.
func stringFromRecord(record map[string]any, key string) (string, bool) {
	raw, ok := record[key]
	if !ok {
		return "", false
	}

	str, ok := raw.(string)
	if !ok || str == "" {
		return "", false
	}

	delete(record, key)

	return str, true
}

// buildJobNestedURL extracts jobId from the record and builds a URL of the form
// /api/v2/jobs/{jobId}/<segments...>. Returns errMissingParentID if jobId is
// missing or empty in the record.
func (c *Connector) buildJobNestedURL(record map[string]any, segments ...string) (*urlbuilder.URL, error) {
	jobID, ok := stringFromRecord(record, parentIDKeyJobID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errMissingParentID, parentIDKeyJobID)
	}

	path := append([]string{c.modulePath(), "jobs", jobID}, segments...)

	return urlbuilder.New(c.ProviderInfo().BaseURL, path...)
}

// modulePath returns the module's URL prefix (e.g. "/api/v2") from schemas.json
// root.path. Used by URL builders here that compose paths with dynamic record
// IDs and therefore can't go through metadata.Schemas.LookupURLPath(objectName).
func (c *Connector) modulePath() string {
	return metadata.Schemas.LookupModuleURLPath(c.ProviderContext.Module())
}

// buildContactNestedURL extracts contactId from the record and builds a URL of
// the form /api/v2/contacts/{contactId}/<segments...>.
func (c *Connector) buildContactNestedURL(record map[string]any, segments ...string) (*urlbuilder.URL, error) {
	contactID, ok := stringFromRecord(record, parentIDKeyContactID)
	if !ok {
		return nil, fmt.Errorf("%w: %s", errMissingParentID, parentIDKeyContactID)
	}

	path := append([]string{c.modulePath(), "contacts", contactID}, segments...)

	return urlbuilder.New(c.ProviderInfo().BaseURL, path...)
}

func (c *Connector) parseWriteResponse(
	_ context.Context,
	params common.WriteParams,
	_ *http.Request,
	response *common.JSONHTTPResponse,
) (*common.WriteResult, error) {
	body, ok := response.Body()
	if !ok {
		// 204 No Content (PUT custom-fields). Echo the caller's RecordId so
		// downstream code can correlate the update.
		return &common.WriteResult{
			Success:  true,
			RecordId: params.RecordId,
		}, nil
	}

	recordID, err := jsonquery.New(body).StrWithDefault("id", params.RecordId)
	if err != nil {
		return nil, err
	}

	data, err := jsonquery.Convertor.ObjectToMap(body)
	if err != nil {
		return nil, err
	}

	return &common.WriteResult{
		Success:  true,
		RecordId: recordID,
		Data:     data,
	}, nil
}
