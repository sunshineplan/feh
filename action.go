package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/sunshineplan/utils/mail"
)

// Update feh scoreboard
func Update() {
	const (
		title = "FEH 投票大戦第%d回 %s - %s"
		body  = "%s\n\n%s"
	)
	event, round, fullScoreboard := Scrape()
	newScoreboard := Record(event, round, fullScoreboard)
	if newScoreboard != nil {
		var content []string
		for _, item := range newScoreboard {
			content = append(content, item.Formatter())
		}
		mailConfig := GetSubscribe()
		err := retry.Do(
			func() error {
				err := mail.SendMail(
					&mailConfig,
					fmt.Sprintf(title, event, Round[round], time.Now().Format("20060102 15:00:00")),
					fmt.Sprintf(body, strings.Join(content, "\n"), time.Now().Format("20060102 15:00:00")),
					nil,
				)
				return err
			},
			retry.Attempts(Attempts),
			retry.Delay(Delay),
			retry.LastErrorOnly(LastErrorOnly),
			retry.OnRetry(func(n uint, err error) {
				log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
			}),
		)
		if err != nil {
			return
		}
	}
	fmt.Println("Update FEH done.")
}

// Backup feh database
func Backup() {
	file := Dump()
	defer os.Remove(file)
	mailConfig := GetSubscribe()
	err := retry.Do(
		func() error {
			err := mail.SendMail(
				&mailConfig,
				fmt.Sprintf("FEH Backup-%s", time.Now().Format("20060102")),
				"",
				&mail.Attachment{FilePath: file, Filename: "database"},
			)
			return err
		},
		retry.Attempts(Attempts),
		retry.Delay(Delay),
		retry.LastErrorOnly(LastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
		}),
	)
	if err != nil {
		return
	}
	fmt.Println("Backup FEH done.")
}

// Upload feh results to github
func Upload(e int) {
	var detail, summary string
	if e == 0 {
		e, _, _ = Scrape()
		detail, summary = Result(e)
		if detail == "" {
			log.Fatal("No data in database.")
		}
	} else {
		detail, summary = Result(e)
		if detail == "" {
			fmt.Printf("No result for event %d. Use last event result instead.\n", e)
			Upload(0)
			return
		}
	}
	c := make(chan int)
	go func() {
		if err := Commit(fmt.Sprintf("FEH 投票大戦第%d回", e), detail); err == nil {
			fmt.Printf("FEH 投票大戦第%d回.json uploaded.\n", e)
		}
		c <- 1
	}()
	if err := Commit(fmt.Sprintf("FEH 投票大戦第%d回結果一覧", e), summary); err == nil {
		fmt.Printf("FEH 投票大戦第%d回結果一覧.json uploaded.\n", e)
	}
	<-c
}
