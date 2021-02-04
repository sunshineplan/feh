package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"feh"

	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/mail"
)

func update() {
	const (
		title = "FEH 投票大戦第%d回 %s - %s"
		body  = "%s\n\n%s"
	)

	var event, round int
	var fullScoreboard, newScoreboard []feh.Scoreboard
	if err := utils.Retry(
		func() (err error) {
			event, round, fullScoreboard, err = feh.Scrape()
			if err != nil {
				log.Print(err)
				return
			}
			newScoreboard, err = record(fullScoreboard)
			if err != nil {
				log.Print(err)
			}
			return
		}, 5, 60); err != nil {
		log.Fatal(err)
	}

	if newScoreboard != nil {
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
				c <- utils.Retry(
					func() error {
						return dialer.Send(
							&mail.Message{
								To: []string{to},
								Subject: fmt.Sprintf(title, event, feh.Round[extra],
									time.Now().In(timezone).Format("20060102 15:00:00")),
								Body: fmt.Sprintf(body, strings.Join(extraContent, "\n"),
									time.Now().In(timezone).Format("20060102 15:04:05")),
							})
					}, 3, 10)
			}()
		} else {
			c <- nil
		}

		if err := utils.Retry(
			func() error {
				return dialer.Send(
					&mail.Message{
						To: []string{to},
						Subject: fmt.Sprintf(title, event, feh.Round[round],
							time.Now().In(timezone).Format("20060102 15:00:00")),
						Body: fmt.Sprintf(body, strings.Join(content, "\n"),
							time.Now().In(timezone).Format("20060102 15:04:05")),
					})
			}, 3, 10); err != nil {
			log.Fatal(err)
		}

		if err := <-c; err != nil {
			log.Fatal(err)
		}
	}
}

func backup() {
	if err := utils.Retry(
		func() error {
			return db.Backup("backup")
		}, 3, 60); err != nil {
		log.Fatal(err)
	}

	if err := utils.Retry(
		func() error {
			return dialer.Send(
				&mail.Message{
					To: []string{to},
					Subject: fmt.Sprintf("FEH Backup-%s",
						time.Now().In(timezone).Format("20060102")),
					Attachments: []*mail.Attachment{{Path: "backup", Filename: "database"}},
				})
		}, 3, 10); err != nil {
		log.Fatal(err)
	}
}

func commit() {
	var event int
	var detail, summary string
	if err := utils.Retry(
		func() (err error) {
			event, _, _, err = feh.Scrape()
			if err != nil {
				return
			}
			detail, summary, err = result(event)
			return
		}, 5, 60); err != nil {
		log.Fatal(err)
	}

	if detail == "" {
		log.Fatal("No data in database.")
	}

	f, err := os.Create("message")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("FEH 投票大戦第%d回", event))
	if err != nil {
		log.Fatal(err)
	}

	c := make(chan error)
	go func() {
		f, err := os.Create(fmt.Sprintf("FEH 投票大戦第%d回.json", event))
		if err != nil {
			c <- err
			return
		}
		defer f.Close()

		_, err = f.WriteString(detail)

		c <- err
	}()

	f, err = os.Create(fmt.Sprintf("FEH 投票大戦第%d回結果一覧.json", event))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	_, err = f.WriteString(summary)
	if err != nil {
		log.Fatal(err)
	}

	if err := <-c; err != nil {
		log.Fatal(err)
	}
}
