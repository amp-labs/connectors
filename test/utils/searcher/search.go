package searcher

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/jsonquery"
	"github.com/amp-labs/connectors/test/utils"
	"github.com/spyzhov/ajson"
)

type Type string

const (
	Object  Type = "object"
	Array   Type = "array"
	String  Type = "string"
	Integer Type = "integer"
)

type Key struct {
	Type  Type
	At    string
	Index int // relevant if type is array
}

func Find[T any](res *common.ReadResult, keys []Key, value T) map[string]any {
	for _, data := range res.Data {
		bytes, err := json.Marshal(data.Fields)
		if err != nil {
			slog.Warn("couldn't marshal data fields")

			continue
		}

		node, err := ajson.Unmarshal(bytes)
		if err != nil {
			slog.Warn("couldn't unmarshal into JSON")

			continue
		}

		// node variable is written to several times in the loop.
		// This accomplishes zooming into JSON.
		for _, key := range keys {
			switch key.Type {
			case String:
				actual, err := jsonquery.New(node).Str(key.At, false)
				if err != nil {
					slog.Warn("string", "error", err)

					continue
				}

				if fmt.Sprintf("%v", *actual) == fmt.Sprintf("%v", value) {
					return data.Fields
				}
			case Integer:
				actual, err := jsonquery.New(node).Integer(key.At, false)
				if err != nil {
					slog.Warn("integer", "error", err)

					continue
				}

				if fmt.Sprintf("%v", *actual) == fmt.Sprintf("%v", value) {
					return data.Fields
				}
			case Array:
				var nodes []*ajson.Node

				nodes, err = jsonquery.New(node).Array(key.At, false)
				if err != nil {
					slog.Warn("array", "error", err)

					continue
				}

				node = nodes[key.Index]
			case Object:
				fallthrough
			default:
				node, err = jsonquery.New(node).Object(key.At, false)
				if err != nil {
					slog.Warn("object", "error", err)

					continue
				}
			}
		}
	}

	utils.Fail("error finding object, it is not found")

	return nil
}
