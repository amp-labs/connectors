package core

import (
	"errors"
	"fmt"
	"os"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/test/utils"
)

// PrintError prints error and returns true if error is not nil.
func PrintError(err error) bool {
	if err == nil {
		return false
	}

	fmt.Println("[test failed]", "error", err.Error())
	if httpError, ok := errors.AsType[*common.HTTPError](err); ok {
		utils.DumpJSON(httpError.Body, os.Stdout)
		return true
	}

	return true
}
