package databricks

import (
	"context"
	"fmt"
	"strings"

	"github.com/amp-labs/connectors/common"
)

// listObjectMetadata describes each table from the catalog's information_schema.
func (c *Connector) listObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result := common.NewListObjectMetadataResult()

	for _, objectName := range objectNames {
		fields, err := c.describeTable(ctx, objectName)
		if err != nil {
			result.AppendError(objectName, err)

			continue
		}

		result.Result[objectName] = *common.NewObjectMetadata(objectName, fields)
	}

	return result, nil
}

func (c *Connector) describeTable(ctx context.Context, table string) (common.FieldsMetadata, error) {
	statement := fmt.Sprintf(
		"SELECT column_name, data_type FROM %s.information_schema.columns "+
			"WHERE table_schema = :schema AND table_name = :table ORDER BY ordinal_position",
		quoteIdentifier(c.catalog),
	)

	records, err := c.executeStatement(ctx, c.newStatementRequest(
		statement,
		statementParameter{Name: "schema", Value: c.schema, Type: "STRING"},
		statementParameter{Name: "table", Value: table, Type: "STRING"},
	))
	if err != nil {
		return nil, err
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("%w: %s", errObjectNotFound, table)
	}

	fields := make(common.FieldsMetadata, len(records))

	for _, record := range records {
		name, _ := record["column_name"].(string)
		dataType, _ := record["data_type"].(string)

		fields[name] = common.FieldMetadata{
			DisplayName:  name,
			ValueType:    valueTypeOf(dataType),
			ProviderType: dataType,
		}
	}

	return fields, nil
}

func valueTypeOf(dataType string) common.ValueType {
	switch strings.ToUpper(dataType) {
	case "STRING", "CHAR", "VARCHAR":
		return common.ValueTypeString
	case "BOOLEAN":
		return common.ValueTypeBoolean
	case "BYTE", "TINYINT", "SHORT", "SMALLINT", "INT", "INTEGER", "LONG", "BIGINT":
		return common.ValueTypeInt
	case "FLOAT", "DOUBLE", "DECIMAL", "NUMERIC":
		return common.ValueTypeFloat
	case "DATE":
		return common.ValueTypeDate
	case "TIMESTAMP", "TIMESTAMP_NTZ":
		return common.ValueTypeDateTime
	default:
		return common.ValueTypeOther
	}
}
