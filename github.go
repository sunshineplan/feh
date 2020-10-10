package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/github"
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

func commit(name, content string) error {
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
		}, 3, 10)
}
