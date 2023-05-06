package utils

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"feh"

	"github.com/sunshineplan/database/mongodb"
	"github.com/sunshineplan/database/mongodb/driver"
	"github.com/sunshineplan/utils/mail"
	"github.com/sunshineplan/utils/retry"
)

func Update(dialer *mail.Dialer, to mail.Receipts, tz *time.Location, db mongodb.Client) error {
	const (
		title = "FEH 投票大戦第%d回 %s - %s"
		body  = "%s\n\n%s"
	)

	var event, round int
	var fullScoreboard, newScoreboard []feh.Scoreboard
	if err := retry.Do(
		func() (err error) {
			event, round, fullScoreboard, err = feh.Scrape()
			if err != nil {
				if err == feh.ErrEventNotOpen {
					err = retry.ErrNoMoreRetry(err.Error())
					return
				}
				log.Print(err)
				return
			}
			newScoreboard, err = record(fullScoreboard, tz, db)
			if err != nil {
				log.Print(err)
			}
			return
		}, 5, 60); err != nil {
		return err
	}

	if dialer != nil && newScoreboard != nil {
		var content []string
		var extra int
		var extraContent []string
		for _, item := range newScoreboard {
			score := item.Formatter()
			fmt.Printf("第%d回 %s: %s\n", event, feh.Round[item.Round], score)
			if item.Round == round {
				content = append(content, score)
			} else {
				extra = item.Round
				extraContent = append(extraContent, score)
			}
		}

		c := make(chan error, 1)
		if extra != 0 {
			go func() {
				c <- retry.Do(
					func() error {
						return dialer.Send(
							&mail.Message{
								To: to,
								Subject: fmt.Sprintf(title, event, feh.Round[extra],
									time.Now().In(tz).Format("20060102 15:00:00")),
								Body: fmt.Sprintf(body, strings.Join(extraContent, "\n"),
									time.Now().In(tz).Format("20060102 15:04:05")),
							})
					}, 3, 10)
			}()
		} else {
			c <- nil
		}

		if err := retry.Do(
			func() error {
				return dialer.Send(
					&mail.Message{
						To: to,
						Subject: fmt.Sprintf(title, event, feh.Round[round],
							time.Now().In(tz).Format("20060102 15:00:00")),
						Body: fmt.Sprintf(body, strings.Join(content, "\n"),
							time.Now().In(tz).Format("20060102 15:04:05")),
					})
			}, 3, 10); err != nil {
			return err
		}

		if err := <-c; err != nil {
			return err
		}
	}

	return nil
}

func Backup(dialer *mail.Dialer, to mail.Receipts, tz *time.Location, db *driver.Client) error {
	file := "backup.tmp"
	if err := retry.Do(
		func() error {
			return db.Backup(file)
		}, 3, 60); err != nil {
		return err
	}
	defer os.Remove(file)

	return retry.Do(
		func() error {
			return dialer.Send(
				&mail.Message{
					To: to,
					Subject: fmt.Sprintf("FEH Backup-%s",
						time.Now().In(tz).Format("20060102")),
					Attachments: []*mail.Attachment{{Path: file, Filename: "database"}},
				})
		}, 3, 10)
}

func Result(event int, tz *time.Location, db mongodb.Client) (int, string, string, error) {
	var detail, summary string
	if event == 0 {
		if err := retry.Do(
			func() (err error) {
				event, _, _, err = feh.Scrape()
				if err != nil {
					return
				}
				detail, summary, err = result(event, tz, db)
				return
			}, 5, 60); err != nil {
			return 0, "", "", err
		}
		if detail == "" {
			return 0, "", "", fmt.Errorf("no data in database")
		}
	} else {
		if err := retry.Do(
			func() (err error) {
				detail, summary, err = result(event, tz, db)
				return
			}, 5, 60); err != nil {
			return 0, "", "", err
		}
		if detail == "" {
			log.Printf("No result for event %d. Use last event result instead.", event)
			return Result(0, tz, db)
		}
	}

	return event, detail, summary, nil
}
