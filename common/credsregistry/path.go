package credsregistry

import (
	"fmt"
	"os"

	"github.com/iancoleman/strcase"
)

// LoadPath will give path to creds.json.
// For provider called `dynamicsCRM` the file location will either be
// * value of DYNAMICS_CRM_CRED_FILE env var, or
// * ./dynamics-crm-creds.json.
func LoadPath(providerName string) string {
	filePath := os.Getenv(
		fmt.Sprintf("%v_CRED_FILE", envNameFormat(providerName)),
	)
	if len(filePath) == 0 {
		filePath = strcase.ToKebab(providerName)
		filePath = fmt.Sprintf("./%v-creds.json", filePath)
	}

	return filePath
}
