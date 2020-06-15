package main

import (
	"context"
	"fmt"
	"time"

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

// Commit single new file to repository
func Commit(name, content string) {
	config := GetGithub()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	token := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: config.Token})
	tc := oauth2.NewClient(ctx, token)

	client := github.NewClient(tc)

	opts := &github.RepositoryContentFileOptions{
		Message: github.String(name),
		Content: []byte(content)}
	_, _, err := client.Repositories.CreateFile(
		ctx,
		config.User,
		config.Repository,
		fmt.Sprintf("%s/%s.json", config.Path, name),
		opts)
	if err != nil {
		fmt.Println(err)
	}
}
