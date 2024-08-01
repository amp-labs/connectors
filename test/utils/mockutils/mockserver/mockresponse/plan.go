package mockresponse

import (
	"net/http"
	"net/url"

	"github.com/amp-labs/connectors/test/utils/mockutils"
)

// Plan will call on success method only if all conditions were met.
type Plan struct {
	body               *string
	header             *http.Header
	methodName         *string
	queryParamsPresent *url.Values
	queryParamsMissing *[]string
	onSuccess          func(w http.ResponseWriter, r *http.Request)
}

// If is a constructor that allows to build a conditional mock response.
// Ex: mockresponse.If().MethodIs("PATCH").QueryParamsAre(url.Values{}).Do(...)
func If() *Plan {
	return &Plan{}
}

func (r *Plan) Check(req *http.Request) bool {
	if r.methodName != nil {
		if req.Method != *r.methodName {
			return false
		}
	}

	if r.body != nil {
		if !mockutils.BodiesMatch(req.Body, *r.body) {
			return false
		}
	}

	if r.header != nil {
		if _, ok := mockutils.HeaderIsSubset(req.Header, *r.header); !ok {
			return false
		}
	}

	if r.queryParamsPresent != nil {
		if _, ok := mockutils.QueryParamsAreSubset(req.URL.Query(), *r.queryParamsPresent); !ok {
			return false
		}
	}

	if r.queryParamsMissing != nil {
		if _, ok := mockutils.QueryParamsMissing(req.URL.Query(), *r.queryParamsMissing); !ok {
			return false
		}
	}

	return true
}

func (r *Plan) BodyIs(body string) *Plan {
	r.body = &body

	return r
}

func (r *Plan) HeaderIs(header http.Header) *Plan {
	r.header = &header
	return r
}

func (r *Plan) MethodIs(methodName string) *Plan {
	r.methodName = &methodName
	return r
}

func (r *Plan) QueryParamsAre(queryParams url.Values) *Plan {
	r.queryParamsPresent = &queryParams
	return r
}

func (r *Plan) QueryParamsMissing(queryParams []string) *Plan {
	r.queryParamsMissing = &queryParams
	return r
}

func (r *Plan) Do(onSuccess func(w http.ResponseWriter, r *http.Request)) *Plan {
	r.onSuccess = onSuccess

	return r
}

func (r *Plan) OnSuccess(w http.ResponseWriter, res *http.Request) {
	if r.onSuccess != nil {
		r.onSuccess(w, res)
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
