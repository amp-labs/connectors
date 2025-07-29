package suiteql

type suiteQLQueryBody struct {
	Query string `json:"q"`
}

// nolint:tagliatelle
type suiteQLResponse struct {
	Links        []suiteQLLink    `json:"links"`
	Count        int              `json:"count"`
	HasMore      bool             `json:"hasMore"`
	Items        []map[string]any `json:"items"`
	Offset       int              `json:"offset"`
	TotalResults int              `json:"totalResults"`
}

type suiteQLLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}
