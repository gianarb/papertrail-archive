package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tipsCmd)
}

var tipsCmd = &cobra.Command{
	Use:   "tips",
	Short: "This command prints a set of tips to read and manipulate archives when they have been downloaded",
	Long: `
A set of tips and tricks to manipulate read archives when they have been
download.

---

$ gzip -cd 2019-05-01-* | grep <search_param>

It looks for a specific string inside one or more archives. You do not need to
gunzip them.

---

$ gzip -cd 2019-05-14-11.tsv.gz | grep <search_param> -n

This command searches a string inside one or multiple archives. it also prints
the lines where it find the occurrences.

---

$ gunzip -k 2019-05-14-11.tsv.gz

it extract the content from the compressed archive.

---

$ gunzip -k 2019-05-14-11.tsv.gz
$ cat ./2019-05-14-11.tsv | grep <search_param> -n
$ less +<line_you_got_from_previous_command> 2019-05-14-11.tsv

This set of commands brings you to the line you was looking for.

---

$ awk '{ print $7 }' ./2019-02-26-00.tsv | grep <app_name>

The default format prints in the column 7 the appname where the logs come from.
With this command you can filter logs based on their appname.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print(cmd.Long)
	},
}
