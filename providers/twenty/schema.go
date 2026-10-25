package twenty

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/urlbuilder"
)

// Twenty has no fixed schema: each workspace generates its objects and fields, standard and custom alike,
// and serves them at GET /rest/metadata/objects, one entry per object with its fields inline.
// https://docs.twenty.com/developers/extend/api ("There is no static API reference for Twenty.")
// Objects are named by namePlural, the segment of their REST path (/rest/companies, /rest/vendorContracts).
const (
	metadataObjectsPath = "rest/metadata/objects"
	metadataPageSize    = 200
)

var errMetadataEnvelope = errors.New("unexpected object metadata response")

// objectsResponse takes both envelopes Twenty v2.50.0 serves, chosen per workspace by the feature flag
// IS_REST_METADATA_API_NEW_FORMAT_DIRECT: {"data": [...]} where it is on (the default for new workspaces),
// {"data": {"objects": [...]}} where it is off (object-metadata.controller.ts and
// to-legacy-object-metadata-response.util.ts in Twenty's source; the flag's description says the legacy
// format is discontinued on November 15, 2026). Both were seen live by switching the flag.
type objectsResponse struct {
	Data     objectList `json:"data"`
	PageInfo pageInfo   `json:"pageInfo"`
}

type objectList []objectMetadata

func (l *objectList) UnmarshalJSON(data []byte) error {
	var direct []objectMetadata
	if err := json.Unmarshal(data, &direct); err == nil {
		*l = direct

		return nil
	}

	var legacy struct {
		Objects *[]objectMetadata `json:"objects"`
	}

	if err := json.Unmarshal(data, &legacy); err != nil || legacy.Objects == nil {
		return fmt.Errorf("%w: data is neither a list of objects nor {\"objects\": [...]}", errMetadataEnvelope)
	}

	*l = *legacy.Objects

	return nil
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type objectMetadata struct {
	NameSingular string          `json:"nameSingular"`
	NamePlural   string          `json:"namePlural"`
	LabelPlural  string          `json:"labelPlural"`
	IsActive     bool            `json:"isActive"`
	Fields       []fieldMetadata `json:"fields"`
}

type fieldMetadata struct {
	ID           string         `json:"id"`
	Type         string         `json:"type"`
	Name         string         `json:"name"`
	Label        string         `json:"label"`
	IsActive     bool           `json:"isActive"`
	IsSystem     bool           `json:"isSystem"`
	IsUIReadOnly bool           `json:"isUIReadOnly"`
	IsNullable   bool           `json:"isNullable"`
	Writability  string         `json:"writability"`
	DefaultValue any            `json:"defaultValue"`
	Options      []fieldOption  `json:"options"`
	Settings     *fieldSettings `json:"settings"`
}

type fieldOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type fieldSettings struct {
	RelationType   string `json:"relationType"`
	JoinColumnName string `json:"joinColumnName"`
}

// recordKey is the key under which a record read with depth=0 carries the field. A many-to-one relation is
// carried as its join column (vendor -> vendorId); a one-to-many relation is not carried at all.
// https://docs.twenty.com/developers/extend/api ("A many-to-one relation field returns its join column
// (e.g. companyId) with depth=0. One-to-many relation fields are only returned with depth=1.").
func (f fieldMetadata) recordKey() (string, bool) {
	if f.Type != "RELATION" && f.Type != "MORPH_RELATION" {
		return f.Name, true
	}

	if f.Settings == nil || f.Settings.JoinColumnName == "" {
		return "", false
	}

	return f.Settings.JoinColumnName, true
}

// readOnly reports fields Twenty manages itself: system fields locked in its UI (id, createdAt, updatedAt,
// deletedAt, createdBy, updatedBy) and fields whose writability is set to something other than OPEN
// (searchVector is SYSTEM, and the API refuses to write it). A missing writability, as an instance older or
// newer than v2.50.0 may send, is read as writable, so writes are not emptied of every field.
func (f fieldMetadata) readOnly() bool {
	return (f.IsSystem && f.IsUIReadOnly) || (f.Writability != "" && f.Writability != "OPEN")
}

// fetchObjects reads every page of the object metadata. The list is not filtered by name in the API, so a
// call returns every object of the workspace with all its fields.
// A page that says there is more without a new cursor fails the call, rather than loop or stop short.
// Its errors name the metadata request, since reads and writes make it too (a key without the data-model
// permission is refused there, not on the record request).
func (c *Connector) fetchObjects(ctx context.Context) (map[string]objectMetadata, error) {
	objects := make(map[string]objectMetadata)
	seen := make(map[string]bool)
	cursor := ""

	for {
		page, err := c.fetchObjectsPage(ctx, cursor)
		if err != nil {
			return nil, fmt.Errorf("twenty object metadata: %w", err)
		}

		for _, object := range page.Data {
			objects[object.NamePlural] = object
		}

		if !page.PageInfo.HasNextPage {
			return objects, nil
		}

		cursor = page.PageInfo.EndCursor
		if cursor == "" || seen[cursor] {
			return nil, fmt.Errorf("%w: object metadata has a next page without a new endCursor (%q)",
				common.ErrNextPageInvalid, cursor)
		}

		seen[cursor] = true
	}
}

func (c *Connector) fetchObjectsPage(ctx context.Context, cursor string) (*objectsResponse, error) {
	pageURL, err := urlbuilder.New(c.ProviderInfo().BaseURL, metadataObjectsPath)
	if err != nil {
		return nil, err
	}

	// The metadata list pages like the record lists, at most 200 objects a page (default 60), so a workspace
	// of up to 200 objects is one request: the OpenAPI document each instance serves at /rest/open-api/metadata
	// (v2.50.0), and seen live (limit=10 returned 10 of 44 objects with hasNextPage).
	pageURL.WithQueryParam("limit", strconv.Itoa(metadataPageSize))

	if cursor != "" {
		pageURL.WithQueryParam("starting_after", cursor)
	}

	response, err := c.JSONHTTPClient().Get(ctx, pageURL.String())
	if err != nil {
		return nil, err
	}

	page, err := common.UnmarshalJSON[objectsResponse](response)
	if err != nil {
		return nil, err
	}

	if page == nil {
		return nil, common.ErrEmptyJSONHTTPResponse
	}

	return page, nil
}

// activeObject returns the object named by its REST path segment, if it is active. The name is matched
// exactly: Twenty's paths are case sensitive (/rest/workspacemembers is a 400, /rest/workspaceMembers is not).
func activeObject(objects map[string]objectMetadata, objectName string) (objectMetadata, bool) {
	object, ok := objects[objectName]
	if !ok || !object.IsActive {
		return objectMetadata{}, false
	}

	return object, true
}

// hasRecordKey reports whether a field of the object, active or not, is carried under this key, ignoring
// case. Twenty still returns the column of an inactive field (seen live on v2.50.0).
func (o objectMetadata) hasRecordKey(name string) bool {
	for _, field := range o.Fields {
		if key, ok := field.recordKey(); ok && lowercase(key) == lowercase(name) {
			return true
		}
	}

	return false
}

func lowercase(name string) string {
	return strings.ToLower(name)
}
