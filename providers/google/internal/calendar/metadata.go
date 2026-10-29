package calendar

import (
	"context"
	_ "embed"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/logging"
	"github.com/amp-labs/connectors/common/naming"
	"github.com/amp-labs/connectors/internal/datautils"
	"github.com/amp-labs/connectors/internal/jsonquery"
	"github.com/amp-labs/connectors/internal/staticschema"
	"github.com/amp-labs/connectors/tools/scrapper"
	"github.com/spyzhov/ajson"
)

// nolint:gochecknoglobals
var (
	// Static file containing a list of object metadata is embedded and can be served.
	//
	//go:embed schemas.json
	schemas []byte

	// Schemas is cached data.
	Schemas = scrapper.NewReader[staticschema.FieldMetadataMapV2](schemas).MustLoadSchemas()
)

const (
	fieldExtendedProperties = "extendedProperties"

	// extendedPropertiesSampleSize is how many primary-calendar events are inspected to
	// discover extendedProperties keys. It is the maximum events.list allows on one page.
	// https://developers.google.com/workspace/calendar/api/v3/reference/events/list
	extendedPropertiesSampleSize = 2500
)

// extendedPropertyBuckets are the two maps Google keeps free-form event properties in.
var extendedPropertyBuckets = []string{"private", "shared"} // nolint:gochecknoglobals

// ListObjectMetadata serves the static schema. For events it additionally discovers the
// free-form extendedProperties keys in use by sampling one page of primary-calendar events,
// and exposes each as "$['extendedProperties']['<bucket>']['<key>']" so they can be picked like any field.
//
// The keys are customer-defined, so they cannot live in the static schema. Discovery is best
// effort: when sampling fails the static schema is returned unchanged and a warning is logged.
func (a *Adapter) ListObjectMetadata(
	ctx context.Context, objectNames []string,
) (*common.ListObjectMetadataResult, error) {
	result, err := a.SchemaProvider.ListObjectMetadata(ctx, objectNames)
	if err != nil {
		return nil, err
	}

	events, ok := result.Result[objectNameEvents]
	if !ok {
		return result, nil
	}

	keys, err := a.discoverExtendedPropertyKeys(ctx)
	if err != nil {
		logging.Logger(ctx).Warn(
			"events metadata: could not discover extendedProperties keys, serving static schema only",
			"error", err,
		)

		return result, nil
	}

	if events.Fields == nil {
		events.Fields = make(common.FieldsMetadata)
	}

	for _, bucket := range extendedPropertyBuckets {
		for _, key := range keys[bucket] {
			if strings.Contains(key, "'") {
				// Cannot be represented in bracket notation; the field name would not parse back.
				continue
			}

			fieldName := jsonPathField(fieldExtendedProperties, bucket, key)

			events.AddFieldMetadata(fieldName, common.FieldMetadata{
				DisplayName:  fmt.Sprintf("Extended Properties %v %v", naming.CapitalizeFirstLetterEveryWord(bucket), key),
				ValueType:    common.ValueTypeString,
				ProviderType: "string",
				IsCustom:     new(true),
			})
		}
	}

	result.Result[objectNameEvents] = events

	return result, nil
}

// discoverExtendedPropertyKeys reads one page of primary-calendar events, requesting only the
// extendedProperties part of each item, and returns the sorted distinct keys per bucket.
//
// events.list without a time window or ordering returns events in the provider's default
// order, so on calendars with more than one page of history the sample may not include the
// most recently created events, and keys used only on other calendars are not seen.
func (a *Adapter) discoverExtendedPropertyKeys(ctx context.Context) (map[string][]string, error) {
	url, err := a.getURL(objectNameEvents)
	if err != nil {
		return nil, err
	}

	url.WithQueryParam("maxResults", strconv.Itoa(extendedPropertiesSampleSize))
	// Partial response: only the data needed for discovery is transferred.
	// https://developers.google.com/workspace/calendar/api/guides/performance#partial-response
	url.WithQueryParam("fields", "items("+fieldExtendedProperties+")")

	resp, err := a.JSONHTTPClient().Get(ctx, url.String())
	if err != nil {
		return nil, err
	}

	keys := make(map[string]datautils.StringSet, len(extendedPropertyBuckets))
	for _, bucket := range extendedPropertyBuckets {
		keys[bucket] = datautils.NewStringSet()
	}

	body, ok := resp.Body()
	if !ok {
		return sortedBucketKeys(keys), nil
	}

	items, err := jsonquery.New(body).ArrayOptional("items")
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if err := collectExtendedPropertyKeys(item, keys); err != nil {
			return nil, err
		}
	}

	return sortedBucketKeys(keys), nil
}

// collectExtendedPropertyKeys adds the extendedProperties keys of one event to the per-bucket sets.
func collectExtendedPropertyKeys(item *ajson.Node, keys map[string]datautils.StringSet) error {
	properties, err := jsonquery.New(item).ObjectOptional(fieldExtendedProperties)
	if err != nil {
		return err
	}

	if properties == nil {
		return nil
	}

	for _, bucket := range extendedPropertyBuckets {
		bucketNode, err := jsonquery.New(properties).ObjectOptional(bucket)
		if err != nil {
			return err
		}

		if bucketNode != nil {
			keys[bucket].Add(bucketNode.Keys())
		}
	}

	return nil
}

func sortedBucketKeys(keys map[string]datautils.StringSet) map[string][]string {
	result := make(map[string][]string, len(keys))

	for bucket, set := range keys {
		list := set.List()
		slices.Sort(list)
		result[bucket] = list
	}

	return result
}
