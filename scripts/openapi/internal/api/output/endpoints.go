package output

import (
	"path/filepath"

	"github.com/amp-labs/connectors/scripts/openapi/internal/api/pipeline"
	"github.com/amp-labs/connectors/scripts/openapi/internal/api/spec"
)

const EndpointsFile = "endpoints.json"

func WriteEndpoints(dirName string,
	readPipe, createPipe, updatePipe, deletePipe *pipeline.Pipeline[spec.Schema],
) error {
	pipelines := map[string]*pipeline.Pipeline[spec.Schema]{
		"read":   readPipe,
		"create": createPipe,
		"update": updatePipe,
		"delete": deletePipe,
	}

	for key, value := range pipelines {
		if value == nil {
			delete(pipelines, key)
		}
	}

	for _, pipe := range pipelines {
		if err := validate(*pipe); err != nil {
			return err
		}
	}

	return Write(
		filepath.Join(dirName, EndpointsFile),
		extractEndpoints(pipelines),
	)
}

func extractEndpoints(pipelines map[string]*pipeline.Pipeline[spec.Schema]) any {
	result := make(map[string]*Actions)

	for action, pipe := range pipelines {
		for _, object := range pipe.List() {
			actions, ok := result[object.ObjectName]
			if !ok {
				actions = &Actions{}
				result[object.ObjectName] = actions
			}

			endpoint := &Endpoint{
				URL:         object.URLPath,
				Operation:   object.Operation,
				ResponseKey: object.ResponseKey,
			}

			switch action {
			case "read":
				actions.Read = endpoint
			case "create":
				actions.Create = endpoint
			case "update":
				actions.Update = endpoint
			case "delete":
				actions.Delete = endpoint
			}
		}
	}

	return result
}

type Actions struct {
	Read   *Endpoint `json:"read,omitempty"`
	Create *Endpoint `json:"create,omitempty"`
	Update *Endpoint `json:"update,omitempty"`
	Delete *Endpoint `json:"delete,omitempty"`
}

type Endpoint struct {
	URL         string `json:"url,omitempty"`
	Operation   string `json:"operation,omitempty"`
	ResponseKey string `json:"responseKey,omitempty"`
}
