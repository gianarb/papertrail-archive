package cmd

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"sync"
	"time"

	"compress/gzip"

	"github.com/gosuri/uiprogress"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const gzipRatio = 10

var urlTemplate = "https://papertrailapp.com/api/v1/archives/%s/download"

var from TimeVar
var to TimeVar
var noInteractive bool
var noGunzip bool
var parallel int

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

type WriteCounter struct {
	Total   uint64
	current uint64
	Bar     *uiprogress.Bar
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	wc.current += uint64(len(p))
	wc.Bar.Set(int((wc.current * 100) / wc.Total))
	return len(p), nil
}

func init() {
	rootCmd.AddCommand(downloadCmd)
	downloadCmd.PersistentFlags().Var(&from, "from", "")
	downloadCmd.PersistentFlags().Var(&to, "to", "")
	downloadCmd.Flags().BoolVarP(&noGunzip, "no-gunzip", "", false, "By default the archive gets decompressed. With this flag it stays compressed.")
	downloadCmd.Flags().BoolVarP(&noInteractive, "no-interactive", "y", false, "Do not ask any question and download.")
	downloadCmd.Flags().StringVarP(&basedir, "basedir", "", "/tmp", "directory where to store the archives.")
	downloadCmd.Flags().IntVar(&parallel, "parallel", 3, "How many downloads to start in parallel")
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
			token = viper.GetString("token")
		}
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
		if basedir == "" {
			basedir = "."
		}
		println("from: " + fromTime.Format(time.RFC3339))
		println("to: " + toTime.Format(time.RFC3339))
		println("file will be stored to directory (--basedir to change location): " + basedir)
		client := http.Client{}
		downloadsProspect := []string{}
		for {
			downloadsProspect = append(downloadsProspect, fmt.Sprintf("%d-%02d-%02d-%02d", fromTime.Year(), fromTime.Month(), fromTime.Day(), fromTime.Hour()))
			fromTime = fromTime.Add(1 * time.Hour)
			if !fromTime.Before(toTime) {
				break
			}
		}

		println("The archives that will be downloaded are:")
		for _, v := range downloadsProspect {
			println("\t" + v)
		}

		if !noInteractive {
			println("do you wan't to proceed with the download? (y)")
			var response string
			_, err := fmt.Scanln(&response)
			if err != nil {
				log.Fatal(err)
			}
			if response != "y" {
				println("You decided to not proceed. The only way to proceed is answering y")
				os.Exit(1)
			}
		}

		progressBars := map[string]*uiprogress.Bar{}
		uiprogress.Start()
		for _, v := range downloadsProspect {
			progressBars[v] = uiprogress.AddBar(100).AppendCompleted().PrependElapsed().PrependFunc(func(vv string) func(b *uiprogress.Bar) string {
				return func(b *uiprogress.Bar) string { return fmt.Sprintf("Archive %s", vv) }
			}(v))
		}

		c := make(chan string)
		var wg sync.WaitGroup
		wg.Add(parallel)
		for ii := 0; ii < parallel; ii++ {
			go func(c chan string) {
				for {
					v, more := <-c
					if more == false {
						wg.Done()
						return
					}
					err := downloadArchive(client, v, progressBars[v])
					if err != nil {
						println(err.Error())
						os.Exit(1)
					}
				}
			}(c)
		}
		for _, v := range downloadsProspect {
			c <- v
		}
		close(c)
		wg.Wait()
		println("I am done.")
	},
}

func downloadArchive(client http.Client, v string, bar *uiprogress.Bar) error {
	tmpFileName := fmt.Sprintf("%s/%s.tsv.gz.tmp", basedir, v)
	out, err := os.Create(tmpFileName)
	if err != nil {
		return err
	}
	defer out.Close()

	req := newRequest()
	u, err := url.Parse(fmt.Sprintf(urlTemplate, v))
	if err != nil {
		return err
	}
	req.URL = u
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	totSize, err := strconv.Atoi(resp.Header.Get("Content-Length"))
	if err != nil {
		return err
	}
	counter := &WriteCounter{
		Total: uint64(totSize),
		Bar:   bar,
	}
	reader := resp.Body
	destFileName := fmt.Sprintf("%s/%s.tsv", basedir, v)
	if !noGunzip {
		// 4 is the std compression ration for gz. I do not know a better way
		// to do it atm
		counter.Total = counter.Total * gzipRatio
		reader, err = gzip.NewReader(resp.Body)
		defer reader.Close()
	} else {
		destFileName = destFileName + ".gz"
	}
	_, err = io.Copy(out, io.TeeReader(reader, counter))
	if err != nil {
		return err
	}
	err = os.Rename(tmpFileName, destFileName)
	if err != nil {
		return err
	}
	// At this point everyting is done, set the bar to complete
	bar.Set(int(counter.Total))
	return nil
}

func newRequest() *http.Request {
	r, _ := http.NewRequest("GET", "", nil)
	r.Header.Add("X-Papertrail-Token", token)
	r.Header.Add("User-Agent", "papertrail-archive/undefined")
	return r
}
