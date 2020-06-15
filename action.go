package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Update feh scoreboard
func Update() {
	const (
		title = "FEH 投票大戦第%d回 %s - %s"
		body  = "%s\n\n%s"
	)
	event, round, fullScoreboard, status := Scrape()
	var content []string
	for _, item := range fullScoreboard {
		content = append(content, item.Formatter())
	}
	Record(event, round, fullScoreboard)
	if status == 1 {
		mailConfig := GetSubscribe()
		if err := Mail(
			&mailConfig,
			fmt.Sprintf(title, event, Round[round], time.Now().Format("20060102 15:00:00")),
			fmt.Sprintf(body, strings.Join(content, "\n"), time.Now().Format("20060102 15:00:00")),
			nil,
		); err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Update FEH done.")
}

// Backup feh database
func Backup() {
	file := Dump()
	defer os.Remove(file)
	mailConfig := GetSubscribe()
	if err := Mail(
		&mailConfig,
		fmt.Sprintf("FEH Backup-%s", time.Now().Format("20060102")),
		"",
		&Attachment{FilePath: file, Filename: "database"},
	); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Backup FEH done.")
}

// Upload feh results to github
func Upload(e int) {
	var detail, summary string
	if e == 0 {
		e, _, _, _ = Scrape()
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
		if err := Commit(fmt.Sprintf("FEH 投票大戦第%d回", e), detail); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("FEH 投票大戦第%d回.json uploaded\n", e)
		}
		c <- 1
	}()
	if err := Commit(fmt.Sprintf("FEH 投票大戦第%d回結果一覧", e), summary); err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("FEH 投票大戦第%d回結果一覧.json uploaded\n", e)
	}
	<-c
}
