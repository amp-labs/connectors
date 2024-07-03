package scrapper

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type ModelDocLinks []ModelDocLink

func (l ModelDocLinks) FindByName(name string) (ModelDocLink, bool) {
	for _, link := range l {
		if link.Name == name {
			return link, true
		}
	}

	return ModelDocLink{}, false
}

type ModelURLRegistry struct {
	ModelDocs ModelDocLinks `json:"data"`
}

type ModelDocLink struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	URL         string `json:"url"`
}

func NewModelURLRegistry() *ModelURLRegistry {
	return &ModelURLRegistry{
		ModelDocs: make([]ModelDocLink, 0),
	}
}

func (r *ModelURLRegistry) Add(displayName, url string) {
	if len(displayName) == 0 || len(url) == 0 {
		// Trying to add URL with no display name or missing URL
		return
	}

	url, _ = strings.CutSuffix(url, "/")
	parts := strings.Split(url, "/")
	name := parts[len(parts)-1]

	r.ModelDocs = append(r.ModelDocs, ModelDocLink{
		DisplayName: displayName,
		Name:        name,
		URL:         url,
	})
}

func (r *ModelURLRegistry) Sort() {
	sort.Slice(r.ModelDocs, func(i, j int) bool {
		return r.ModelDocs[i].Name < r.ModelDocs[j].Name
	})
}

type ObjectMetadataResult struct {
	// Result is a map of object names to object metadata
	Result map[string]ObjectMetadata `json:"data"`
}

type ObjectMetadata struct {
	// Provider's display name for the object
	DisplayName string `json:"displayName"`

	// FieldsMap is a map of field names to field display names
	FieldsMap map[string]string `json:"fields"`
}

func NewObjectMetadataResult() *ObjectMetadataResult {
	return &ObjectMetadataResult{
		Result: make(map[string]ObjectMetadata),
	}
}

func (r *ObjectMetadataResult) Add(objectName string, objectDisplayName string, fieldName string) {
	data, ok := r.Result[objectName]
	if !ok {
		data = ObjectMetadata{
			DisplayName: objectDisplayName,
			FieldsMap:   make(map[string]string),
		}
		r.Result[objectName] = data
	}

	data.FieldsMap[fieldName] = fieldName
}

type QueryParamStats struct {
	Meta queryParamStatsMeta     `json:"meta"`
	Data []queryParamObjectStats `json:"queryParams"`
}

type queryParamStatsMeta struct {
	TotalObjects int      `json:"totalObjects"`
	CollectedAt  DateTime `json:"collectedAt"`
}

type queryParamObjectStats struct {
	Name         string   `json:"name"`
	Frequency    float64  `json:"frequency"`
	TotalObjects int      `json:"totalObjects"`
	Objects      []string `json:"objects"`
}

func NewQueryParamStats(totalObjects int) *QueryParamStats {
	return &QueryParamStats{
		Meta: queryParamStatsMeta{
			TotalObjects: totalObjects,
			CollectedAt:  DateTime{Time: time.Now()},
		},
		Data: make([]queryParamObjectStats, 0),
	}
}

func (s *QueryParamStats) Add(queryParamName string, objectNames []string) {
	num := len(objectNames)
	freq := float64(num) / float64(s.Meta.TotalObjects)
	s.Data = append(s.Data, queryParamObjectStats{
		Name:         queryParamName,
		Frequency:    roundFloat(freq, 4), // nolint:gomnd
		TotalObjects: num,
		Objects:      objectNames,
	})
}

func roundFloat(f float64, decPlaces int) float64 {
	target := math.Pow(10, float64(decPlaces)) // nolint:gomnd

	return float64(int(f*target)) / target
}

type DateTime struct {
	Time time.Time
}

func (s DateTime) MarshalJSON() ([]byte, error) {
	format := s.Time.Format(time.DateOnly)

	return []byte(fmt.Sprintf(`"%v"`, format)), nil
}

func (s *DateTime) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	// remove string quotes
	if len(str) < 2 { // nolint:gomnd
		return errors.New("date time has no quotes") // nolint:goerr113
	}

	format := str[1 : len(str)-1]

	t, err := time.Parse(time.DateTime, format)
	if err != nil {
		return err
	}

	s.Time = t

	return nil
}
