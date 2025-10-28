package outplay

import (
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/internal/components"
	"github.com/amp-labs/connectors/internal/datautils"
)

//nolint:gochecknoglobals
var (
	ObjectNameProspect        = "prospect"
	ObjectNameProspectAccount = "prospectaccount"
	ObjectNameSequence        = "sequence"
	ObjectNameCall            = "call"
	ObjectNameTask            = "task"
	ObjectNameCallAnalysis    = "callanalysis"
	ObjectNameProspectMails   = "prospectmails"
	ObjectNameNote            = "note"
)

var objectAPIPath = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	ObjectNameProspect:        "prospect/search",
	ObjectNameProspectAccount: "prospectaccount/search",
	ObjectNameSequence:        "sequence/search",
	ObjectNameCall:            "call/search",
	ObjectNameTask:            "task/list",
	ObjectNameCallAnalysis:    "callanalysis/list",
	ObjectNameProspectMails:   "prospectmails/list",
}, func(objectName string) string {
	return objectName
})

var writeObjectAPIPath = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	ObjectNameSequence: "sequence/create",
	ObjectNameNote:     "note/create",
	ObjectNameTask:     "task/create",
}, func(objectName string) string {
	return objectName
})

var writeObjectResponseIDField = datautils.NewDefaultMap(datautils.Map[string, string]{ //nolint:gochecknoglobals
	ObjectNameProspectAccount: "accountid",
	ObjectNameNote:            "noteId",
}, func(objectName string) string {
	// Default ID field pattern
	return objectName + "id"
})

func supportedOperations() components.EndpointRegistryInput {
	readSupport := []string{
		ObjectNameProspect,
		ObjectNameProspectAccount,
		ObjectNameSequence,
		ObjectNameCall,
		ObjectNameTask,
		ObjectNameCallAnalysis,
		ObjectNameProspectMails,
	}

	writeSupport := []string{
		ObjectNameProspect,
		ObjectNameProspectAccount,
		ObjectNameSequence,
		ObjectNameNote,
		ObjectNameTask,
	}

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
		},
	}
}
