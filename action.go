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

func update() {
	const (
		title = "FEH 投票大戦第%d回 %s - %s"
		body  = "%s\n\n%s"
	)
	event, round, fullScoreboard := scrape()
	newScoreboard := record(fullScoreboard)
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
		mailConfig := getSubscribe()
		c := make(chan int, 1)
		if extra != 0 {
			go func() {
				if err := retry.Do(
					func() error {
						err := mail.SendMail(
							&mailConfig,
							fmt.Sprintf(title, event, Round[extra], time.Now().Format("20060102 15:00:00")),
							fmt.Sprintf(body, strings.Join(extraContent, "\n"), time.Now().Format("20060102 15:00:00")),
						)
						return err
					},
					retry.Attempts(attempts),
					retry.Delay(delay),
					retry.LastErrorOnly(lastErrorOnly),
					retry.OnRetry(func(n uint, err error) {
						log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
					}),
				); err != nil {
					log.Fatal("Mail result failed.")
				}
				c <- 1
			}()
		} else {
			c <- 1
		}
		if err := retry.Do(
			func() error {
				err := mail.SendMail(
					&mailConfig,
					fmt.Sprintf(title, event, Round[round], time.Now().Format("20060102 15:00:00")),
					fmt.Sprintf(body, strings.Join(content, "\n"), time.Now().Format("20060102 15:00:00")),
				)
				return err
			},
			retry.Attempts(attempts),
			retry.Delay(delay),
			retry.LastErrorOnly(lastErrorOnly),
			retry.OnRetry(func(n uint, err error) {
				log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
			}),
		); err != nil {
			log.Fatal("Mail result failed.")
		}
		<-c
	}
	fmt.Println("Update FEH done.")
}

func backup() {
	file := dump()
	defer os.Remove(file)
	mailConfig := getSubscribe()
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
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.LastErrorOnly(lastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Mail delivery failed. #%d: %s\n", n+1, err)
		}),
	)
	if err != nil {
		return
	}
	fmt.Println("Backup FEH done.")
}

func upload(e int) {
	var detail, summary string
	if e == 0 {
		e, _, _ = scrape()
		detail, summary = result(e)
		if detail == "" {
			log.Fatal("No data in database.")
		}
	} else {
		detail, summary = result(e)
		if detail == "" {
			fmt.Printf("No result for event %d. Use last event result instead.\n", e)
			upload(0)
			return
		}
	}
	c := make(chan int)
	go func() {
		if err := commit(fmt.Sprintf("FEH 投票大戦第%d回", e), detail); err == nil {
			fmt.Printf("FEH 投票大戦第%d回.json uploaded.\n", e)
		}
		c <- 1
	}()
	if err := commit(fmt.Sprintf("FEH 投票大戦第%d回結果一覧", e), summary); err == nil {
		fmt.Printf("FEH 投票大戦第%d回結果一覧.json uploaded.\n", e)
	}
	<-c
}
