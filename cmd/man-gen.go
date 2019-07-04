package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var basedir string

func init() {
	rootCmd.AddCommand(mangenCmd)
	mangenCmd.PersistentFlags().StringVar(&basedir, "basedir", "", "The directory where man files will be placed. (default .)")
}

var mangenCmd = &cobra.Command{
	Use:   "man-gen",
	Short: "Generate man",
	Run: func(cmd *cobra.Command, args []string) {
		header := &doc.GenManHeader{
			Title:   "MINE",
			Section: "3",
		}
		err := doc.GenManTree(rootCmd, header, basedir)
		if err != nil {
			log.Fatal(err)
		}
	},
}
