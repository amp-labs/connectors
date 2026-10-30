package output

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

func TestExtractObjectMetadata(t *testing.T) {
	objectUsers := `{
		"objectName": "users",
		"displayName": "Users",
		"urlPath": "/api/v1/users",
		"fields": {
			"id": {
				"name": "id",
				"displayName": "User ID",
				"type": "string",
				"valueType": "string"
			}
		}
	}`
	objectsProjects := `{
		"objectName": "projects",
		"displayName": "Projects",
		"urlPath": "/api/v1/projects",
		"fields": {
			"id": {
				"name": "id",
				"displayName": "Project ID",
				"type": "string",
				"valueType": "string"
			}
		}
	}`
	objectUsersCustom := `"users": {
		"displayName": "Custom Users",
		"path": "/api/v2/users",
		"fields": {
			"id": {
				"displayName": "Custom User ID",
				"valueType": "string",
				"providerType": "string"
			}
		}
	}`
	objectTasks := `"tasks": {
		"displayName": "Custom Tasks",
		"path": "/api/v1/tasks",
		"fields": {
			"id": {
				"displayName": "Task ID",
				"valueType": "string",
				"providerType": "string"
			}
		}
	}`

	mainJSON := []byte(fmt.Sprintf("[%v,%v]", objectUsers, objectsProjects))
	manualJSON := []byte(fmt.Sprintf("{%v, %v}", objectUsersCustom, objectTasks))

	var schemas []spec.Schema
	if err := json.Unmarshal(mainJSON, &schemas); err != nil {
		t.Fatalf("failed to unmarshal main JSON: %v", err)
	}

	pipe := pipeline.NewSchemaPipe(schemas)
	meta, err := extractObjectMetadata(
		createMetadataParams([]MetadataOption{
			WithManualObjects(manualJSON),
		}),
		pipe,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(meta) != 3 {
		t.Fatalf("expected 3 objects, got %d", len(meta))
	}

	// Verify manual JSON overrides the keys (users object is overwritten by manual JSON)
	usersMeta, ok := meta["users"]
	if !ok {
		t.Fatalf("expected users object in metadata")
	}
	if usersMeta.DisplayName != "Custom Users" || usersMeta.Path != "/api/v2/users" {
		t.Errorf("expected users object to be overridden by manual JSON, got: %+v", usersMeta)
	}
	if len(usersMeta.Fields) != 1 {
		t.Errorf("expected 1 field in users, got %d", len(usersMeta.Fields))
	}
	if idField, ok := usersMeta.Fields["id"]; !ok || idField.DisplayName != "Custom User ID" {
		t.Errorf("unexpected field in overridden users object: %+v", idField)
	}

	// Verify main JSON object is preserved when not overridden (projects object)
	projMeta, ok := meta["projects"]
	if !ok {
		t.Fatalf("expected projects object in metadata")
	}
	if projMeta.DisplayName != "Projects" || projMeta.Path != "/api/v1/projects" {
		t.Errorf("unexpected projects metadata: %+v", projMeta)
	}
	if len(projMeta.Fields) != 1 {
		t.Errorf("expected 1 field in projects, got %d", len(projMeta.Fields))
	}
	if idField, ok := projMeta.Fields["id"]; !ok || idField.DisplayName != "Project ID" {
		t.Errorf("unexpected field in projects object: %+v", idField)
	}

	// Verify manual JSON adds new objects (tasks object)
	tasksMeta, ok := meta["tasks"]
	if !ok {
		t.Fatalf("expected tasks object in metadata")
	}
	if tasksMeta.DisplayName != "Custom Tasks" || tasksMeta.Path != "/api/v1/tasks" {
		t.Errorf("unexpected tasks metadata: %+v", tasksMeta)
	}
	if len(tasksMeta.Fields) != 1 {
		t.Errorf("expected 1 field in tasks, got %d", len(tasksMeta.Fields))
	}
	if idField, ok := tasksMeta.Fields["id"]; !ok || idField.DisplayName != "Task ID" {
		t.Errorf("unexpected field in tasks object: %+v", idField)
	}
}
