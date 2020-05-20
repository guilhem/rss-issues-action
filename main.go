package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
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

	// Parse repository in form owner/name
	repo := strings.Split(os.Getenv("GITHUB_REPOSITORY"), "/")

	// Parse limit time option
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

	// Parse Labels
	labels := strings.Split(gha.GetInput("labels"), ",")
	gha.Debug(fmt.Sprintf("labels %v", labels), ghaLogOption)

	ctx := context.Background()

	// Instanciate GitHub client
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: gha.GetInput("repo-token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Instanciate feed parser
	fp := gofeed.NewParser()
	feed, err := fp.ParseURLWithContext(gha.GetInput("feed"), ctx)
	if err != nil {
		gha.Error(fmt.Sprintf("Cannot parse feed '%s': '%s'", gha.GetInput("feed"), err), ghaLogOption)
		os.Exit(1)
	}
	gha.Info(feed.Title)

	// Instanciate HTML to markdown
	converter := md.NewConverter("", true, nil)

	// Remove old items in feed
	feed.Items = funk.Filter(feed.Items, func(x *gofeed.Item) bool {
		return x.PublishedParsed.After(limitTime)
	}).([]*gofeed.Item)

	// Get all issues
	IssueListByRepoOption := &github.IssueListByRepoOptions{
		State:  "all",
		Labels: labels,
	}

	issues, _, err := client.Issues.ListByRepo(ctx, repo[0], repo[1], IssueListByRepoOption)
	if err != nil {
		gha.Error(fmt.Sprint(err), ghaLogOption)
		os.Exit(1)
	}
	gha.Debug(fmt.Sprintf("%d issues", len(issues)), ghaLogOption)

	var createdIssues []string

	// Iterate
	for _, item := range feed.Items {
		title := strings.Join([]string{gha.GetInput("prefix"), item.Title}, " ")
		gha.Debug(fmt.Sprintf("Issue '%s'", title), ghaLogOption)

		if issue := funk.Find(issues, func(x *github.Issue) bool {
			return *x.Title == title
		}); issue != nil {
			gha.Warning("Issue already exists", ghaLogOption)
			continue
		}

		// Issue Content

		content := item.Content
		if content == "" {
			content = item.Description
		}

		markdown, err := converter.ConvertString(content)
		if err != nil {
			gha.Error(fmt.Sprintf("Fail to convert HTML to markdown: '%s'", err), ghaLogOption)
			continue
		}

		// Execute the template with a map as context
		context := map[string]string{
			"Link":    item.Link,
			"Content": markdown,
		}

		const issue = `
{{if .Link}}
[{{ .Link }}]({{ .Link }})

{{end}}
{{if .Content}}
{{ .Content }}
{{end}}
`
		var tpl bytes.Buffer
		if err := template.Must(template.New("issue").Parse(issue)).Execute(&tpl, context); err != nil {
			gha.Warning(fmt.Sprintf("Cannot render issue: '%s'", err), ghaLogOption)
			continue
		}

		body := tpl.String()

		// Create Issue

		issueRequest := &github.IssueRequest{
			Title:  &title,
			Body:   &body,
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
