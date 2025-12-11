package scrapper

import (
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/amp-labs/connectors/internal/datautils"
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

func (r *ModelURLRegistry) AddModelByName(name, url string) {
	url, _ = strings.CutSuffix(url, "/")

	r.ModelDocs = append(r.ModelDocs, ModelDocLink{
		DisplayName: name,
		Name:        name,
		URL:         url,
	})
}

func (r *ModelURLRegistry) Sort() {
	sort.Slice(r.ModelDocs, func(i, j int) bool {
		return r.ModelDocs[i].Name < r.ModelDocs[j].Name
	})
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

// CalculateQueryParamStats produces statistics on objects and their query parameters.
// queryParamRegistry - holds query parameter name to the list of object names that use it.
func CalculateQueryParamStats(queryParamRegistry datautils.NamedLists[string]) *QueryParamStats {
	objects := datautils.NewStringSet()
	for _, objectNames := range queryParamRegistry {
		objects.Add(objectNames)
	}

	totalUniqueObject := len(objects)

	stats := NewQueryParamStats(totalUniqueObject)
	queryParams := queryParamRegistry.GetBuckets()

	// sort query parameters, where most occurred come first
	sort.SliceStable(queryParams, func(i, j int) bool {
		a := queryParams[i]
		b := queryParams[j]
		l1 := len(queryParamRegistry[a])
		l2 := len(queryParamRegistry[b])

		if l1 == l2 {
			return a < b
		}

		return l1 > l2
	})

	for _, queryParam := range queryParams {
		stats.SaveParameterStats(queryParam, queryParamRegistry[queryParam])
	}

	return stats
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

func (s *QueryParamStats) SaveParameterStats(queryParamName string, objectNames []string) {
	num := len(objectNames)
	freq := float64(num) / float64(s.Meta.TotalObjects)

	sort.Strings(objectNames)

	s.Data = append(s.Data, queryParamObjectStats{
		Name:         queryParamName,
		Frequency:    roundFloat(freq, 4), // nolint:mnd
		TotalObjects: num,
		Objects:      objectNames,
	})
}

func roundFloat(f float64, decPlaces int) float64 {
	target := math.Pow(10, float64(decPlaces)) // nolint:mnd

	return float64(int(f*target)) / target
}

type DateTime struct { //nolint:recvcheck
	Time time.Time
}

func (s DateTime) MarshalJSON() ([]byte, error) {
	format := s.Time.Format(time.DateOnly)

	return fmt.Appendf(nil, `"%v"`, format), nil
}

func (s *DateTime) UnmarshalJSON(bytes []byte) error {
	str := string(bytes)
	// remove string quotes
	if len(str) < 2 { // nolint:mnd
		return errors.New("date time has no quotes") // nolint:err113
	}

	format := str[1 : len(str)-1]

	t, err := time.Parse(time.DateTime, format)
	if err != nil {
		return err
	}

	s.Time = t

	return nil
}
