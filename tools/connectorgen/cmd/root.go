package cmd

import (
	"errors"
	"fmt"
	"log"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "conn-gen",
	Short: "Connector generator",
	Long:  "Generates the base template to start building connector with",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalf("unhandled error: %v", err)
	}
}

func init() {
	rootCmd.PersistentFlags().StringP("package", "p", "", "Golang package name")
	rootCmd.PersistentFlags().StringP("provider", "n", "",
		"Provider name. By default <package> upper camel case.")
	rootCmd.PersistentFlags().StringP("output", "o", "",
		"Output directory. By default <package>-output-gen")

	if err := errors.Join(
		rootCmd.MarkPersistentFlagRequired("package"),
		viper.BindPFlag("package", rootCmd.PersistentFlags().Lookup("package")),
		viper.BindPFlag("provider", rootCmd.PersistentFlags().Lookup("provider")),
		viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output")),
	); err != nil {
		log.Fatal(err)
	}
}

type Recipe struct {
	Package  string
	Provider string
	Output   string
}

func GetRecipe() *Recipe {
	result := &Recipe{
		Package:  viper.GetString("package"),
		Provider: viper.GetString("provider"),
		Output:   viper.GetString("output"),
	}
	if len(result.Output) == 0 {
		result.Output = fmt.Sprintf("%v-output-gen", result.Package)
	}

	if len(result.Provider) == 0 {
		result.Provider = strcase.ToCamel(result.Package)
	}

	return result
}
