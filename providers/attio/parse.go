// nolint
package attio

import (
	"github.com/spyzhov/ajson"
)

var dummyNextPageFunc = func(*ajson.Node) (string, error) {
	return "", nil
}
