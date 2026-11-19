package search

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/readhelper"
	"github.com/amp-labs/connectors/common/searchhelper"
	"github.com/amp-labs/connectors/common/urlbuilder"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/providers/connectwise/internal/reader"
)

const RawTypeFilterConditions = "conditions"

func (s Strategy) Search(ctx context.Context, params *common.SearchParams) (*common.SearchResult, error) {
	if err := params.ValidateParams(false); err != nil {
		return nil, err
	}

	url, err := s.buildSearchUrl(params)
	if err != nil {
		return nil, err
	}

	resp, err := s.base.JSONHTTPClient().Get(ctx, url.String(), s.base.ClientIdHeader())
	if err != nil {
		return nil, err
	}

	return common.ParseResult(
		resp,
		// "" is used, because root level of JSON is right away an array.
		common.MakeRecordsFunc(""),
		reader.NextRecordsURL(resp),
		readhelper.MakeMarshaledDataFuncWithId(
			reader.RecordTransformer(params.ObjectName),
			readhelper.IdFieldQuery{Field: "id"},
		),
		params.Fields,
	)
}

func (s Strategy) buildSearchUrl(params *common.SearchParams) (*urlbuilder.URL, error) {
	if params.NextPage != "" {
		return urlbuilder.New(params.NextPage.String())
	}

	url, err := s.base.GetUrl(params.ObjectName)
	if err != nil {
		return nil, err
	}

	conditions, err := getConditionsQueryParam(params)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("conditions", conditions)
	url.WithQueryParam("pageSize", searchhelper.PageSizeWithDefaultStr(params, reader.DefaultPageSize))

	return url, nil
}

// Conditions query param can be provided via RawFilter or must be built by iterating through FieldFilters.
func getConditionsQueryParam(params *common.SearchParams) (string, error) {
	rawCondition, err := getOptionalRawCondition(params)
	if err != nil {
		return "", err
	}

	if rawCondition != "" {
		return rawCondition, nil
	}

	return buildConditionsQueryParam(params)
}

func getOptionalRawCondition(params *common.SearchParams) (string, error) {
	if params.RawFilter != nil {
		filterType, err := searchhelper.ValidateRawFilterType(params, RawTypeFilterConditions)
		if err != nil {
			return "", err
		}

		if filterType == RawTypeFilterConditions {
			return common.StringFilterValue(params)
		}
	}

	return "", nil
}

func buildConditionsQueryParam(params *common.SearchParams) (string, error) {
	conditions := make([]string, 0, len(params.Filter.FieldFilters))
	for _, filter := range params.Filter.FieldFilters {
		switch filter.Operator {
		case common.FilterOperatorEQ:
			value, err := translateEqualityFieldValue(filter.Value)
			if err != nil {
				return "", err
			}

			conditions = append(conditions, fmt.Sprintf(`%s = %s`, filter.FieldName, value))
		case common.FilterOperatorIn:
			value, err := translateInclusionFieldValue(filter.Value)
			if err != nil {
				return "", err
			}

			conditions = append(conditions, fmt.Sprintf(`%s in (%s)`, filter.FieldName, value))
		default:
			return "", common.ErrSearchFilterOperatorUnsupported
		}
	}

	return strings.Join(conditions, " and "), nil
}

func translateEqualityFieldValue(value any) (string, error) {
	switch data := value.(type) {
	case string:
		return strconv.Quote(data), nil
	case bool:
		// No quotes for bool. Must be either True or False according to docs.
		if data {
			return "True", nil
		}

		return "False", nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		// No quotes for numbers.
		return fmt.Sprint(data), nil
	case time.Time:
		// Time must be in square brackets.
		return "[" + datautils.Time.FormatRFC3339inUTC(data) + "]", nil
	default:
		return "", fmt.Errorf("%w: unsupported equality value type %T",
			common.ErrSearchFilterValueUnsupported, value)
	}
}

func translateInclusionFieldValue(value any) (string, error) {
	data, ok := value.([]string)
	if ok {
		return strings.Join(data, ","), nil
	}

	return "", fmt.Errorf("%w: unsupported inclusion value type %T",
		common.ErrSearchFilterValueUnsupported, value)
}
