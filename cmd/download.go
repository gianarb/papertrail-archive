package cmd

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var urlTemplate = "https://papertrailapp.com/api/v1/archives/%s/download"

var from TimeVar
var to TimeVar

type TimeVar struct {
	time *time.Time
}

func (t *TimeVar) String() string {
	if t.time == nil {
		return ""
	}
	return t.time.Format(time.RFC3339)
}

func (t *TimeVar) Time() *time.Time {
	return t.time
}

func (t *TimeVar) Set(v string) error {
	tt, err := time.Parse(time.RFC3339, v)
	if err != nil {
		return err
	}
	t.time = &tt
	return nil
}

func (t *TimeVar) Type() string {
	return "time"
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().Var(&from, "from", "")
	downloadCmd.PersistentFlags().Var(&to, "to", "")
}

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download archives logs from papertrail",
	Long: `The download command can be used to download one or more archive
specifying --from and --to. The format of those flags follows RFC3339 with "T".
example: 2019-06-30T00:23:02+02:00.
Both of them are converted to UTC format because that's the timezone used by Papertrail.
If you are on Linux you can use the "date" cli tool to generate the data. Here some examples.

$ papertrail-archive download --token your-token --from $( date -d "4 days ago" --rfc-3339=seconds  | sed 's/ /T/')

This command downloads a single archive from four days ago. The hour used
in this case is the one you was when you triggered the command (plus
timezone conversion).

Let's suppose you starte the command "2019-06-30T00:23:02+02:00" the start
time will be 2019-06-29T22:23:02Z UTC.

"--from" and "--token" are required, when "--to" is not specified it uses
the "--from" value. That's why this command downloads only on hour of data.

To download more hours or days you can combine --from and --to

$ ./papertrail-archive download --token your-token --from "2019-06-30T00:23:02+02:00" --to "2019-06-30T3:23:02+02:00"
from: 2019-06-29T22:23:02Z
to: 2019-06-30T01:23:02Z
downloading 2019-06-29-22`,
	Run: func(cmd *cobra.Command, args []string) {
		if token == "" {
			println("A token is required in order to persist this action. Please pass the --token flag.")
			os.Exit(1)
		}
		fromTime := from.Time().UTC()
		var toTime time.Time
		if to.Time() == nil {
			toTime = fromTime
		} else {
			toTime = to.Time().UTC()
		}
		println("from: " + fromTime.Format(time.RFC3339))
		println("to: " + toTime.Format(time.RFC3339))
		client := http.Client{}
		for {
			println("downloading " + fmt.Sprintf("%d-%02d-%02d-%02d", fromTime.Year(), fromTime.Month(), fromTime.Day(), fromTime.Hour()))

			out, err := os.Create(fmt.Sprintf("%d-%02d-%02d-%02d.tsv.gz", fromTime.Year(), fromTime.Month(), fromTime.Day(), fromTime.Hour()))
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			defer out.Close()

			req := newRequest()
			u, err := url.Parse(fmt.Sprintf(urlTemplate, fmt.Sprintf("%d-%02d-%02d-%02d", fromTime.Year(), fromTime.Month(), fromTime.Day(), fromTime.Hour())))
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			req.URL = u
			//b, _ := httputil.DumpRequest(req, true)
			//fmt.Printf("%s", b)
			resp, err := client.Do(req)
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}
			defer resp.Body.Close()
			_, err = io.Copy(out, resp.Body)
			if err != nil {
				println(err.Error())
				os.Exit(1)
			}

			fromTime = fromTime.Add(1 * time.Hour)
			if !fromTime.Before(toTime) {
				break
			}
		}
	},
}

func newRequest() *http.Request {
	r, _ := http.NewRequest("GET", "", nil)
	r.Header.Add("X-Papertrail-Token", token)
	r.Header.Add("User-Agent", "papertrail-archive/undefined")
	return r
}
