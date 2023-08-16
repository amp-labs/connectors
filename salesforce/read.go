package salesforce

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/spyzhov/ajson"
)

func (s *Connector) Read(ctx context.Context, config common.ReadParams) (*common.ReadResult, error) {
	var data *ajson.Node
	var err error

	if len(config.Fields) == 0 {
		return nil, errors.New("no fields specified")
	}

	if len(config.NextPage) > 0 {
		data, err = s.get(ctx, fmt.Sprintf("https://%s%s", s.Domain, config.NextPage))
	} else if config.Since.IsZero() {
		fields := strings.Join(config.Fields, ",")
		soql := fmt.Sprintf("SELECT %s FROM %s", fields, config.ObjectName)

		qp := url.Values{}
		qp.Add("q", soql)
		data, err = s.get(ctx, s.BaseURL+"/query/?"+qp.Encode())
	} else {
		fields := strings.Join(config.Fields, ",")
		soql := fmt.Sprintf("SELECT %s FROM %s WHERE SystemModstamp > %s", fields, config.ObjectName, config.Since.Format("2006-01-02T15:04:05Z"))

		qp := url.Values{}
		qp.Add("q", soql)
		data, err = s.get(ctx, s.BaseURL+"/query/?"+qp.Encode())
	}

	if err != nil {
		return nil, err
	}

	ts, err := getTotalSize(data)
	if err != nil {
		return nil, err
	}

	records, err := getRecords(data)
	if err != nil {
		return nil, err
	}

	nextPage, err := getNextRecordsUrl(data)
	if err != nil {
		return nil, err
	}

	done, err := getDone(data)
	if err != nil {
		return nil, err
	}

	return &common.ReadResult{
		Rows:     ts,
		Data:     records,
		NextPage: nextPage,
		Done:     done,
	}, nil
}

func getRecords(node *ajson.Node) ([]map[string]interface{}, error) {
	records, err := node.GetKey("records")
	if err != nil {
		return nil, err
	}

	if !records.IsArray() {
		return nil, errors.New("records isn't an array")
	}

	arr := records.MustArray()
	var out []map[string]interface{}

	for _, v := range arr {
		if !v.IsObject() {
			return nil, errors.New("record isn't an object")
		}

		data, err := v.Unpack()
		if err != nil {
			return nil, err
		}

		m, ok := data.(map[string]interface{})
		if !ok {
			return nil, errors.New("record isn't an object")
		}

		out = append(out, m)
	}

	return out, nil
}

func getNextRecordsUrl(node *ajson.Node) (string, error) {
	var nextPage string
	if node.HasKey("nextRecordsUrl") {
		next, err := node.GetKey("nextRecordsUrl")
		if err != nil {
			return "", err
		}

		if !next.IsString() {
			return "", errors.New("nextRecordsUrl isn't a string")
		}

		nextPage = next.MustString()
	}
	return nextPage, nil
}

func getDone(node *ajson.Node) (bool, error) {
	var done bool
	if node.HasKey("done") {
		doneNode, err := node.GetKey("done")
		if err != nil {
			return false, err
		}

		if !doneNode.IsBool() {
			return false, errors.New("done isn't a boolean")
		}

		done = doneNode.MustBool()
	}
	return done, nil
}

func getTotalSize(node *ajson.Node) (int, error) {
	node, err := node.GetKey("totalSize")
	if err != nil {
		return 0, err
	}

	if !node.IsNumeric() {
		return 0, errors.New("totalSize isn't numeric")
	}

	return int(node.MustNumeric()), nil
}
