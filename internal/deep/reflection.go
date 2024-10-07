package deep

import (
	"errors"
	"github.com/amp-labs/connectors/common"
	"github.com/amp-labs/connectors/common/paramsbuilder"
	"reflect"
)

// TODO revisit every method in this file

// ExtractCatalogVariables
// TODO since we are using reflection, ensure the method is safe from panics.
// TODO parameters that are catalog variables SHOULDN'T be pointers. They could be both.
func ExtractCatalogVariables(structure any) ([]paramsbuilder.CatalogVariable, error) {
	v := reflect.ValueOf(structure)

	if v.Kind() == reflect.Pointer {
		v = reflect.Indirect(v)
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("Expected a struct")
	}

	catalogVariables := make([]paramsbuilder.CatalogVariable, 0)
	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)

		if value.Kind() == reflect.Struct {
			if !value.CanAddr() {
				return nil, errors.New("cannot get address to struct")
			}
			value = value.Addr()
		}

		if !value.CanInterface() {
			continue
		}

		a := value.Interface()
		catalogVar, ok := a.(paramsbuilder.CatalogVariable)
		if ok {
			catalogVariables = append(catalogVariables, catalogVar)
		}
	}

	return catalogVariables, nil
}

func ExtractHTTPClient(structure any) (*common.HTTPClient, error) {
	v := reflect.ValueOf(structure)

	if v.Kind() == reflect.Pointer {
		v = reflect.Indirect(v)
	}

	if v.Kind() != reflect.Struct {
		return nil, errors.New("Expected a struct")
	}

	for i := 0; i < v.NumField(); i++ {
		value := v.Field(i)

		if !value.CanInterface() {
			continue
		}

		a := value.Interface()
		client, ok := a.(paramsbuilder.Client)
		if ok {
			return client.Caller, nil
		}
	}

	return nil, errors.New("http client not found on struct")
}
