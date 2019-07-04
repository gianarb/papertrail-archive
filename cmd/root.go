package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var token string

func init() {
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "the papertrail authentication token.")
}

var rootCmd = &cobra.Command{
	Use:   "papertrail-archive",
	Short: "papertrai-archive is a cli tool that provides utils to download and manage logs archived by papertrail.",
	Long: `Papertrail is an as a service logging platform. Based on your plan
it archives old logs as tabular files per hour. This cli util is made to
provide shortcuts to download and manage logs.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
