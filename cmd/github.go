package main

import (
	"context"
	"fmt"
	"log"
	"time"

	feh "feh/utils"

	"github.com/google/go-github/v37/github"
	"github.com/sunshineplan/utils/retry"
	"golang.org/x/oauth2"
)

// Github info
type Github struct {
	User       string
	Repository string `json:"repo"`
	Token      string
	Path       string
}

func createFile(name, content string) error {
	config := getGithub()
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(name),
		Content: []byte(content)}
	return retry.Do(
		func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			tc := oauth2.NewClient(ctx, token)
			client := github.NewClient(tc)
			_, _, err := client.Repositories.CreateFile(
				ctx,
				config.User,
				config.Repository,
				fmt.Sprintf("%s/%s.json", config.Path, name),
				opts)
			return err
		}, 3, 10*time.Second)
}

func upload(event int) (err error) {
	var detail, summary string
	event, detail, summary, err = feh.Result(event, time.Local, &db)
	if err != nil {
		return
	}

	c := make(chan error)
	go func() {
		err := createFile(fmt.Sprintf("FEH 投票大戦第%d回", event), detail)
		if err == nil {
			log.Printf("FEH 投票大戦第%d回.json uploaded.", event)
		}
		c <- err
	}()

	if err = createFile(fmt.Sprintf("FEH 投票大戦第%d回結果一覧", event), summary); err == nil {
		log.Printf("FEH 投票大戦第%d回結果一覧.json uploaded.", event)
	} else {
		return
	}

	return <-c
}
