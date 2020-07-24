package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// Github info
type Github struct {
	User       string
	Repository string
	Token      string
	Path       string
}

func commit(name, content string) error {
	config := getGithub()
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(name),
		Content: []byte(content)}
	err := retry.Do(
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
		},
		retry.Attempts(attempts),
		retry.Delay(delay),
		retry.LastErrorOnly(lastErrorOnly),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("File commit failed. #%d: %s\n", n+1, err)
		}),
		retry.RetryIf(func(err error) bool {
			if strings.Contains(err.Error(), "sha") {
				log.Printf("File commit failed. %s.json already exists.\n%s\n", name, err)
				return false
			}
			return true
		}),
	)
	return err
}
