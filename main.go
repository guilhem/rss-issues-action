package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/go-github/v31/github"
	"golang.org/x/oauth2"

	funk "github.com/thoas/go-funk"

	"github.com/mmcdole/gofeed"

	gha "github.com/haya14busa/go-actions-toolkit/core"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

func main() {

	ghaLogOption := &gha.LogOption{File: "main.go"}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gha.GetInput("repo-token")},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(gha.GetInput("feed"))
	if err != nil {
		gha.Error(fmt.Sprintf("Cannot parse feed '%s': '%s'", gha.GetInput("feed"), err), ghaLogOption)
		os.Exit(1)
	}
	gha.Info(feed.Title)

	IssueListByRepoOption := &github.IssueListByRepoOptions{}
	repo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")
	issues, _, err := client.Issues.ListByRepo(ctx, repo[0], repo[1], IssueListByRepoOption)
	if err != nil {
		gha.Error(fmt.Sprint(err), ghaLogOption)
		os.Exit(1)
	}
	gha.Debug(fmt.Sprintf("Issues %v", issues), ghaLogOption)

	converter := md.NewConverter("", true, nil)

	var createdIssues []string
	var limitTime time.Time
	if d, err := time.ParseDuration(gha.GetInput("lastTime")); err == nil {
		// Make duration negative
		if d > 0 {
			d = -d
		}
		limitTime = time.Now().Add(d)
	} else {
		gha.Debug(fmt.Sprintf("Fail to parse last time %s", gha.GetInput("lastTime")), ghaLogOption)
	}
	gha.Debug(fmt.Sprintf("limitTime %s", limitTime), ghaLogOption)

	labels := strings.Split(gha.GetInput("labels"), ",")
	gha.Debug(fmt.Sprintf("labels %v", labels), ghaLogOption)

	// Iterate
	for _, item := range feed.Items {
		title := strings.Join([]string{gha.GetInput("prefix"), item.Title}, " ")
		gha.Debug(fmt.Sprintf("Issue '%s'", title), ghaLogOption)

		var t *time.Time
		if item.UpdatedParsed != nil {
			t = item.UpdatedParsed
			gha.Debug("Use 'updated' field", ghaLogOption)
		} else if item.PublishedParsed != nil {
			t = item.PublishedParsed
			gha.Debug("Use 'published' field", ghaLogOption)
		} else {
			gha.Warning("No timed field", ghaLogOption)
		}

		if t != nil && t.Before(limitTime) {
			gha.Debug(fmt.Sprintf("Item date '%s' is before limit", t), ghaLogOption)
			continue
		}

		if issue := funk.Find(issues, func(x *github.Issue) bool {
			return *x.Title == title
		}); issue != nil {
			gha.Warning("Issue already exists", ghaLogOption)
			continue
		}

		markdown, err := converter.ConvertString(item.Content)
		if err != nil {
			gha.Error(fmt.Sprintf("Fail to convert HTML to markdown: '%s'", err), ghaLogOption)
			continue
		}

		issueRequest := &github.IssueRequest{
			Title:  &title,
			Body:   &markdown,
			Labels: &labels,
		}

		if dr, err := strconv.ParseBool(gha.GetInput("dry-run")); err != nil || !dr {
			_, _, err := client.Issues.Create(ctx, repo[0], repo[1], issueRequest)
			if err != nil {
				gha.Warning(fmt.Sprintf("Fail create issue %s: %s", *issueRequest.Title, err), ghaLogOption)
				continue
			}
		} else {
			gha.Debug(fmt.Sprintf("Creating Issue '%s' with content '%s'", *issueRequest.Title, *issueRequest.Body), ghaLogOption)
		}
		createdIssues = append(createdIssues, *issueRequest.Title)
	}

	gha.SetOutput("issues", strings.Join(createdIssues, ","))
}
