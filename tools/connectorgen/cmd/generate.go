package cmd

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"text/template" // nosemgrep:go.lang.security.audit.xss.import-text-template.import-text-template

	"github.com/amp-labs/connectors/common/naming"
	"github.com/iancoleman/strcase"
)

var customFunctions = template.FuncMap{ // nolint:gochecknoglobals
	"camel":     strcase.ToCamel,
	"loweCamel": strcase.ToLowerCamel,
	"snake":     strcase.ToSnake,
	"upper":     strings.ToUpper,
	"kebab":     strcase.ToKebab,
	"singular": func(text string) string {
		return naming.NewSingularString(text).String()
	},
	"plural": func(text string) string {
		return naming.NewPluralString(text).String()
	},
}

func applyTemplatesFromDirectory(directoryName string, data any, dirOutput string) {
	for _, fileName := range getTemplateNames(filepath.Join("../template", directoryName)) {
		path := resolveRelativePath(filepath.Join("../template", directoryName, fileName))
		if isDir(path) {
			continue
		}

		tmpl, err := template.New(fileName).Funcs(customFunctions).ParseFiles(path)
		if err != nil {
			log.Fatalf("failed parsing template %v\n", err)
		}

		writeOutputFile(tmpl, data, dirOutput, fileName)
	}
}

func writeOutputFile(tmpl *template.Template, data any, outputDirectoryName, templateFileName string) {
	if err := os.MkdirAll(outputDirectoryName, os.ModePerm); err != nil {
		log.Fatalf("failed creating output directory %v\n", err)
	}

	outputFileName, _ := strings.CutSuffix(templateFileName, ".tmpl")

	file, err := os.Create(filepath.Join(outputDirectoryName, outputFileName))
	if err != nil {
		log.Fatalf("failed creating output file %v\n", err)
	}

	err = tmpl.Execute(file, data)
	if err != nil {
		_ = file.Close()

		log.Fatalf("failed applying template %v\n", err)
	}

	_ = file.Close()
}

func getTemplateNames(dirName string) []string {
	files, err := os.ReadDir(resolveRelativePath(dirName))
	if err != nil {
		log.Fatal(err)
	}

	templateNames := make([]string, len(files))
	for i, file := range files {
		templateNames[i] = file.Name()
	}

	return templateNames
}

func resolveRelativePath(filename string) string {
	_, thisMethodsLocation, _, _ := runtime.Caller(0) // nolint:dogsled
	localDir := filepath.Dir(thisMethodsLocation)

	return filepath.Join(localDir, filename)
}

func isDir(name string) bool {
	if fi, err := os.Stat(name); err == nil {
		if fi.Mode().IsDir() {
			return true
		}
	}

	return false
}
