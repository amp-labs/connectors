package credscanning

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/iancoleman/strcase"
)

// LoadPath will give path to creds.json.
// For provider called `dynamicsCRM` the file location will either be
// * value of DYNAMICS_CRM_CRED_FILE env var, or
// * ./dynamics-crm-creds.json.
//
// Suffixes are optional to specify.
// Primarily, this is useful when the provider has multiple files, each designated to a connector module.
func LoadPath(providerName string, suffixes ...string) string {
	var filePath string

	provider := strcase.ToKebab(providerName)
	suffix := createFileSuffix(suffixes)

	// Infer file name from the environment variable or
	// construct the expected file path.
	filePath = pathFromENV(provider, suffix)
	if filePath == "" {
		filePath = fileInOS(provider, suffix)
	}

	slog.Debug("loading credentials file", "path", filePath)

	return filePath
}

func fileInOS(provider string, suffix string) string {
	filePath := fmt.Sprintf("./%v-creds.json", provider+suffix)

	if !fileExists(filePath) {
		// Fallback to file path without suffixes.
		oldPath := filePath
		filePath = fmt.Sprintf("./%v-creds.json", provider)
		slog.Warn("credentials file does not exist, using fallback", "path", oldPath, "newPath", filePath)
	}

	return filePath
}

func pathFromENV(provider string, suffix string) string {
	filePath := os.Getenv(
		fmt.Sprintf("%v_CRED_FILE", envNameFormat(provider+suffix)),
	)

	if filePath != "" {
		return filePath
	}

	// Fallback to env var without suffixes.
	return os.Getenv(
		fmt.Sprintf("%v_CRED_FILE", envNameFormat(provider)),
	)
}

func createFileSuffix(suffixes []string) string {
	parts := make([]string, 0, len(suffixes))
	for _, part := range suffixes {
		parts = append(parts, strings.ToLower(part))
	}

	suffix := strings.Join(parts, "-")
	if suffix != "" {
		suffix = fmt.Sprintf("-%v", suffix)
	}

	return suffix
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return !errors.Is(err, os.ErrNotExist)
}
