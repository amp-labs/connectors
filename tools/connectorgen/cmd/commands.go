package cmd

import (
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
)

var baseCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "base",
	Short: "Sets up boilerplate for a base auth connector",
	Long:  "Provides a template of connector with sample struct, constructor, params, error handler",
	Run: func(cmd *cobra.Command, args []string) {
		recipe := GetRecipe()
		applyTemplatesFromDirectory("base", recipe,
			filepath.Join(recipe.Output, recipe.Package),
		)
		completed(recipe)
	},
}

var readCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "read objectName <objectName, ex: contact, user>",
	Short: "Create read method",
	Long: "Provides a template to start implementing a read method. " +
		"Manual test will have read template for objectName",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recipe := GetRecipe()
		applyTemplatesFromDirectory("read", recipe,
			filepath.Join(recipe.Output, recipe.Package),
		)
		createManualTest(recipe, args[0], "read")
		completed(recipe)
	},
}

var writeCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "write objectName <objectName, ex: contact, user>",
	Short: "Create write method",
	Long: "Provides a template to start implementing a write method. " +
		"Manual test will have create/update/delete template for objectName",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recipe := GetRecipe()
		applyTemplatesFromDirectory("write", recipe,
			filepath.Join(recipe.Output, recipe.Package),
		)
		createManualTest(recipe, args[0], "write-delete")
		completed(recipe)
	},
}

var deleteCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "delete objectName <objectName, ex: contact, user>",
	Short: "Create delete method",
	Long: "Provides a template to start implementing a delete method. " +
		"Manual test will have create/update/delete template for objectName",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recipe := GetRecipe()
		applyTemplatesFromDirectory("delete", recipe,
			filepath.Join(recipe.Output, recipe.Package),
		)
		createManualTest(recipe, args[0], "write-delete")
		completed(recipe)
	},
}

var metadataCmd = &cobra.Command{ //nolint:gochecknoglobals
	Use:   "metadata objectName <objectName, ex: contact, user>",
	Short: "Create metadata method",
	Long: "Provides a template to start implementing a list object metadata method. " +
		"Manual test will have template for objectName",
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		recipe := GetRecipe()
		applyTemplatesFromDirectory("metadata", recipe,
			filepath.Join(recipe.Output, recipe.Package),
		)
		createManualTest(recipe, args[0], "metadata")
		completed(recipe)
	},
}

func createManualTest(recipe *Recipe, objectName string, directory string) {
	applyTemplatesFromDirectory("test", recipe,
		filepath.Join(recipe.Output, "test", recipe.Package),
	)

	type ManualTestParams struct {
		*Recipe
		ObjectName string
	}

	data := &ManualTestParams{
		Recipe:     recipe,
		ObjectName: objectName,
	}
	applyTemplatesFromDirectory(filepath.Join("test", directory), data,
		filepath.Join(recipe.Output, "test", recipe.Package, directory),
	)
}

func completed(recipe *Recipe) {
	log.Printf("Template generation completed.\nLocate output at [%v] directory\n", recipe.Output)
}

func init() {
	rootCmd.AddCommand(baseCmd)
	rootCmd.AddCommand(readCmd)
	rootCmd.AddCommand(writeCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(metadataCmd)
}
