package sageintacct

type SageIntacctMetadataResponse struct {
	Result SageIntacctResult `json:"ia::result"`
	Meta   SageIntacctMeta   `json:"ia::meta"`
}

type SageIntacctResult struct {
	Fields               map[string]SageIntacctFieldDef `json:"fields"`
	Groups               map[string]SageIntacctGroup    `json:"groups,omitempty"`
	HTTPMethods          string                         `json:"httpMethods"`
	Refs                 map[string]SageIntacctRef      `json:"refs,omitempty"`
	Lists                any                            `json:"lists,omitempty"`
	IdempotenceSupported bool                           `json:"idempotenceSupported"`
	Href                 string                         `json:"href,omitempty"`
	Type                 string                         `json:"type,omitempty"`
}

type SageIntacctFieldDef struct {
	Type      string   `json:"type"`
	Mutable   bool     `json:"mutable"`
	Nullable  bool     `json:"nullable"`
	ReadOnly  bool     `json:"readOnly"`
	WriteOnly bool     `json:"writeOnly"`
	Required  bool     `json:"required"`
	Enum      []string `json:"enum,omitempty"`
	Default   any      `json:"default,omitempty"`
}
type SageIntacctGroup struct {
	Fields map[string]SageIntacctFieldDef `json:"fields"`
}

type SageIntacctRef struct {
	APIObject string                         `json:"apiObject"`
	Fields    map[string]SageIntacctFieldDef `json:"fields"`
}

type SageIntacctMeta struct {
	TotalCount   int `json:"totalCount"`
	TotalSuccess int `json:"totalSuccess"`
	TotalError   int `json:"totalError"`
}
