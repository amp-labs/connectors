package deep

import (
	"github.com/amp-labs/connectors/common/paramsbuilder"
)

type Assignable[A any] interface {
	CopyFrom(assignable *A)
}


type Recipe struct {
	Params         []paramsbuilder.ParamAssurance
	DefaultOptions func(params paramsbuilder.ParamAssurance)
}

//func ConnectorBuilder(recipe Recipe) {
//	for i, param := range recipe.Params {
//		if p, ok := param.(paramsbuilder.Client) {
//			httpClient := p.Caller
//		}
//
//
//		if catalogVar, ok := param.(paramsbuilder.CatalogVariable) {
//			// Instead we should collect all catalog variables before applying them.
//			providerInfo, err := providers.ReadInfo(conn.Provider(), &catalogVar)
//			if err != nil {
//				return nil, err
//			}
//		}
//	}
//
//
//}
