package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var baseDir string

func init() {
	rootCmd.AddCommand(docgenCmd)
	docgenCmd.PersistentFlags().StringVar(&baseDir, "basedir", "", "The directory where doc file will be placed. (default .)")
}

var docgenCmd = &cobra.Command{
	Use:   "doc-gen",
	Short: "Generate markdown documentation",
	Run: func(cmd *cobra.Command, args []string) {
		if baseDir == "" {
			baseDir = "."
		}
		err := doc.GenMarkdownTree(rootCmd, baseDir)
		if err != nil {
			println(err.Error())
			os.Exit(1)
		}
	},
}
