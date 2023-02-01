package cmd

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/mail"
	"os"
	"reflect"
	"runtime"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/google/uuid"
	"github.com/iunary/sesify/internal/compaign"
	"github.com/iunary/sesify/internal/sender"
	"github.com/iunary/sesify/internal/worker"
	"github.com/k0kubun/go-ansi"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

var (
	workers    int64
	recipients string
	from       string
	template   string
	subject    string
	attachment string
	delay      time.Duration
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "A brief description of your command",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// check if template file exists
		if template != "" {
			if _, err := os.Stat(template); err != nil {
				return err
			}
		}
		// check if recipients file exists
		if recipients != "" {
			if _, err := os.Stat(recipients); err != nil {
				return err
			}
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// check workers count
		if workers < 1 || workers > int64(runtime.NumCPU()) {
			workers = 4
		}

		// check attachment file if set and exists
		if attachment != "" {
			if _, err := os.Stat(attachment); err != nil {
				cobra.CheckErr(err)
			}
		}

		// check template
		f, err := os.ReadFile(template)
		if err != nil {
			cobra.CheckErr(err)
		}
		subscribers, err := loadRecipients(recipients)
		cobra.CheckErr(err)

		stats := compaign.Stats{
			Delivered: 0,
			Failed:    0,
		}

		bar := progressbar.NewOptions(len(subscribers),
			progressbar.OptionSetWriter(ansi.NewAnsiStdout()),
			progressbar.OptionEnableColorCodes(true),
			progressbar.OptionSetWidth(50),
			progressbar.OptionSetTheme(progressbar.Theme{
				Saucer:        "[cyan]✓",
				SaucerHead:    "[yellow]-[reset]",
				SaucerPadding: "[white]•",
				BarStart:      "[blue]|[reset]",
				BarEnd:        "[blue]|[reset]",
			}),
			progressbar.OptionOnCompletion(func() {
				log.Println("Done")
			}),
		)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		provider := sender.NewSES(os.Getenv("AWS_REGION"), os.Getenv("AWS_SES_ID"), os.Getenv("AWS_SES_KEY"))

		s := worker.NewWorker(workers, provider)
		go s.Run(ctx)

		go func() {
			for _, subscriber := range subscribers {
				s.Compaigns <- compaign.Compaign{
					Subject:   subject,
					Message:   string(f),
					Recepient: subscriber,
					Delay:     delay,
				}
			}
			close(s.Compaigns)
		}()

		for {
			select {
			case r, ok := <-s.Results:
				if !ok {
					continue
				}

				if r.Delivered {
					bar.Add(1)
					stats.Delivered++
				}

				if r.Error != nil {
					log.Println(r.Error.Error())
					stats.Failed++
				}
			case <-s.Done:
				PrintStats(&stats)
				return nil
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)
	sendCmd.PersistentFlags().StringVarP(&from, "from", "f", "", "email from")
	sendCmd.PersistentFlags().StringVarP(&subject, "subject", "s", "hello there", "email subject")
	sendCmd.PersistentFlags().StringVarP(&recipients, "recipients", "l", "", "csv recipients emails list file with the following format(email,firstname,lastname)")
	sendCmd.PersistentFlags().StringVarP(&template, "template", "t", "", "email message template")
	// sendCmd.PersistentFlags().StringVarP(&attachment, "attachemnt", "a", "", "email attachment file")
	sendCmd.PersistentFlags().Int64VarP(&workers, "workers", "w", 4, "maximum concurrent worker that will attempt to send messages simulaneously (max: 10)")
	sendCmd.PersistentFlags().DurationVarP(&delay, "delay", "d", 500, "email send delay")
	sendCmd.MarkPersistentFlagRequired("from")
	sendCmd.MarkPersistentFlagRequired("recipients")
	sendCmd.MarkPersistentFlagRequired("template")
	if err := sendCmd.MarkPersistentFlagFilename("recipients", "csv"); err != nil {
		log.Fatalln(err.Error())
	}
	if err := sendCmd.MarkPersistentFlagFilename("template", "html"); err != nil {
		log.Fatalln(err.Error())
	}
}

func loadRecipients(filepath string) ([]*compaign.Recipient, error) {
	recipients := make([]*compaign.Recipient, 0)
	f, err := os.Open(filepath)
	if err != nil {
		return recipients, err
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		return recipients, err
	}
	for i, row := range data {
		if i > 0 {
			if ma, err := mail.ParseAddress(row[0]); err == nil {
				recipients = append(recipients, &compaign.Recipient{
					UUID:      uuid.New().String(),
					Email:     ma.Address,
					Firstname: row[1],
					Lastname:  row[2],
				})
			}
		}
	}
	return recipients, nil
}

func PrintStats(stats *compaign.Stats) {
	fmt.Fprintln(os.Stdout, "\n\033[1;33mStats\033[0m")
	writer := tabwriter.NewWriter(os.Stdout, 10, 4, 10, '\t', 10)
	v := reflect.ValueOf(stats).Elem()
	for i := 0; i < v.NumField(); i++ {
		fmt.Fprintln(writer, strings.Join([]string{v.Type().Field(i).Name, fmt.Sprint(v.Field(i).Int())}, "\t"))
	}
	writer.Flush()
}
