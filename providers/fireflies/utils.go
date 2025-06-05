package fireflies

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/spyzhov/ajson"
)

const (
	defaultPageSize              = 50
	usersObjectName              = "users"
	transcriptsObjectName        = "transcripts"
	bitesObjectName              = "bites"
	objectNameLiveMeeting        = "liveMeeting"
	objectNameCreateBite         = "bite"
	objectNameSetUserRole        = "userRole"
	objectNameUploadAudio        = "audio"
	objectNameUpdateMeetingTitle = "meetingTitle"
	objectNamedeleteTranscript   = "transcript"
)

var supportLimitAndSkip = datautils.NewSet( //nolint:gochecknoglobals
	transcriptsObjectName,
	bitesObjectName,
)

func makeNextRecordsURL(params common.ReadParams, count int) func(*ajson.Node) (string, error) {
	return func(node *ajson.Node) (string, error) {
		if !supportLimitAndSkip.Has(params.ObjectName) {
			return "", nil
		}

		if count < defaultPageSize {
			return "", nil
		}

		var (
			currentPage int
			err         error
		)

		if params.NextPage != "" {
			currentPage, err = strconv.Atoi(params.NextPage.String())
			if err != nil {
				return "", err
			}
		}

		nextPage := currentPage + count

		return strconv.Itoa(nextPage), nil
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
