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
				return
			}
			newScoreboard, err = record(fullScoreboard)
			return
		}, 5, 60); err != nil {
		log.Fatal(err)
	}

	if newScoreboard != nil {
		var content []string
		var extra int
		var extraContent []string
		for _, item := range newScoreboard {
			if item.Round == round {
				content = append(content, item.Formatter())
			} else {
				extra = item.Round
				extraContent = append(extraContent, item.Formatter())
			}
		}

		dialer, to := getSubscribe()
		c := make(chan error, 1)
		if extra != 0 {
			go func() {
				c <- utils.Retry(
					func() error {
						return dialer.Send(
							&mail.Message{
								To:      to,
								Subject: fmt.Sprintf(title, event, feh.Round[extra], time.Now().Format("20060102 15:00:00")),
								Body:    fmt.Sprintf(body, strings.Join(extraContent, "\n"), time.Now().Format("20060102 15:00:00")),
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
						To:      to,
						Subject: fmt.Sprintf(title, event, feh.Round[round], time.Now().Format("20060102 15:00:00")),
						Body:    fmt.Sprintf(body, strings.Join(content, "\n"), time.Now().Format("20060102 15:00:00")),
					})
			}, 3, 10); err != nil {
			log.Fatal(err)
		}

		if err := <-c; err != nil {
			log.Fatal(err)
		}
	}
	log.Print("Update FEH done.")
}

func backup() {
	file := "backup.tmp"
	if err := utils.Retry(
		func() error {
			return db.Backup(file)
		}, 3, 60); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(file)

	dialer, to := getSubscribe()
	if err := utils.Retry(
		func() error {
			return dialer.Send(
				&mail.Message{
					To:          to,
					Subject:     fmt.Sprintf("FEH Backup-%s", time.Now().Format("20060102")),
					Attachments: []*mail.Attachment{{Path: file, Filename: "database"}},
				})
		}, 3, 10); err != nil {
		log.Fatal(err)
	}
	log.Print("Backup FEH done.")
}

func upload(e int) {
	var detail, summary string
	if e == 0 {
		if err := utils.Retry(
			func() (err error) {
				e, _, _, err = feh.Scrape()
				if err != nil {
					return
				}
				detail, summary, err = result(e)
				return
			}, 5, 60); err != nil {
			log.Fatal(err)
		}
		if detail == "" {
			log.Fatal("No data in database.")
		}
	} else {
		if err := utils.Retry(
			func() (err error) {
				detail, summary, err = result(e)
				return
			}, 5, 60); err != nil {
			log.Fatal(err)
		}
		if detail == "" {
			log.Printf("No result for event %d. Use last event result instead.", e)
			upload(0)
			return
		}
	}

	c := make(chan error)
	go func() {
		err := commit(fmt.Sprintf("FEH 投票大戦第%d回", e), detail)
		if err == nil {
			log.Printf("FEH 投票大戦第%d回.json uploaded.", e)
		}
		c <- err
	}()

	if err := commit(fmt.Sprintf("FEH 投票大戦第%d回結果一覧", e), summary); err == nil {
		log.Printf("FEH 投票大戦第%d回結果一覧.json uploaded.", e)
	} else {
		log.Fatal(err)
	}

	if err := <-c; err != nil {
		log.Fatal(err)
	}
}
