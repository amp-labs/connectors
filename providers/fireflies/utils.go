package fireflies

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize              = 50
	usersObjectName              = "users"
	transcriptsObjectName        = "transcripts"
	bitesObjectName              = "bites"
	objectNameLiveMeeting        = "addToLiveMeeting"
	objectNameCreateBite         = "createBite"
	objectNameSetUserRole        = "setUserRole"
	objectNameUploadAudio        = "uploadAudio"
	objectNameUpdateMeetingTitle = "updateMeetingTitle"
	objectNamedeleteTranscript   = "deleteTranscript"
)

var supportLimitAndSkip = datautils.NewSet( //nolint:gochecknoglobals
	transcriptsObjectName,
	bitesObjectName,
)

func getRecords(objectName string) func(*ajson.Node) ([]map[string]any, error) {
	return func(node *ajson.Node) ([]map[string]any, error) {
		// First get the data object
		dataNode, err := node.GetKey("data")
		if err != nil {
			return nil, err
		}

		// Then get the array under the object name (e.g., "boards" or "users")
		records, err := jsonquery.New(dataNode).ArrayOptional(objectName)
		if err != nil {
			return nil, err
		}

		return jsonquery.Convertor.ArrayToMap(records)
	}
}

func makeNextRecordsURL(params common.ReadParams, count int) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		if supportLimitAndSkip.Has(params.ObjectName) {
			if count < defaultPageSize {
				return "", nil
			}

			var currentPage int
			if params.NextPage != "" {
				_, err := fmt.Sscanf(string(params.NextPage), "%d", &currentPage)
				if err != nil {
					return "", fmt.Errorf("invalid next page format: %w", err)
				}
			}

			nextPage := currentPage + count

			return strconv.Itoa(nextPage), nil
		}

		return "", nil
	}
}

func supportedOperations() components.EndpointRegistryInput {
	// We support reading everything under schema.json, so we get all the objects and join it into a pattern.
	readSupport := []string{usersObjectName, transcriptsObjectName, bitesObjectName}
	writeSupport := []string{objectNameLiveMeeting, objectNameCreateBite, objectNameSetUserRole, objectNameUploadAudio, objectNameUpdateMeetingTitle} // nolint
	deleteSupport := []string{objectNamedeleteTranscript}

	return components.EndpointRegistryInput{
		common.ModuleRoot: {
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(readSupport, ",")),
				Support:  components.ReadSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(writeSupport, ",")),
				Support:  components.WriteSupport,
			},
			{
				Endpoint: fmt.Sprintf("{%s}", strings.Join(deleteSupport, ",")),
				Support:  components.DeleteSupport,
			},
		},
	}
}
