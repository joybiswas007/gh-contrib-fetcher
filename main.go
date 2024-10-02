package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

type application struct {
	user           string
	resultsPerPage int
}

func main() {
	var app application

	flag.StringVar(&app.user, "user", "", "The handle for the GitHub user account")
	flag.IntVar(&app.resultsPerPage, "results-per-page", 30, "The number of results per page (max 100)")

	flag.Parse()

	if app.user == "" {
		log.Println("github username is required to fetch user activity")
		os.Exit(0)
	}

	events, err := app.fetchUserActivity()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Output:")
	if len(events) == 0 {
		log.Printf("No contributions found for the GitHub user '%s'. It appears they have not made any public contributions recently or their activity is private.\n", app.user)
		os.Exit(0)
	}
	for _, event := range events {
		switch event.Type {
		case "PushEvent":
			fmt.Printf("- Pushed %d commits to %s\n", len(event.Payload.Commits), event.Repo.Name)
		case "PullRequestEvent":
			fmt.Printf("- Pulled request from %s branch into %s\n", event.Payload.PullRequest.User.Login, event.Repo.Name)
		case "WatchEvent":
			fmt.Printf("- Starred %s\n", event.Repo.Name)
		case "IssuesEvent":
			fmt.Printf("- Opened a new issue in %s\n", event.Repo.Name)
		case "ForkEvent":
			fmt.Printf("- Forked %s\n", event.Repo.Name)
		case "IssueCommentEvent":
			fmt.Printf("- %s commented on an issue in the %s\n", event.Payload.Issue.User.Login, event.Repo.Name)
		default:
			fmt.Println("- ...")
		}
	}
}

func (app application) fetchUserActivity() (Events, error) {
	values := url.Values{}
	values.Add("per_page", fmt.Sprintf("%d", app.resultsPerPage))
	urlWithParams := fmt.Sprintf("https://api.github.com/users/%s/events?", app.user) + values.Encode()

	client := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, urlWithParams, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var eventErr EventError
		err := json.NewDecoder(resp.Body).Decode(&eventErr)
		if err != nil {
			return nil, err
		}
		msg := fmt.Sprintf("error: %s (statusCode: %s), check documentation: %s", eventErr.Message, eventErr.Status, eventErr.DocumentationURL)
		return nil, errors.New(msg)
	}

	var events Events
	err = json.NewDecoder(resp.Body).Decode(&events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

type Events []Event

type Event struct {
	ID           string       `json:"id"`
	Type         string       `json:"type"`
	User         User         `json:"actor"`
	Repo         Repo         `json:"repo"`
	Payload      Payload      `json:"payload"`
	Public       bool         `json:"public"`
	CreatedAt    time.Time    `json:"created_at"`
	Organization Organization `json:"org,omitempty"`
}

type User struct {
	ID           int    `json:"id"`
	Login        string `json:"login"`
	NodeID       string `json:"node_id"`
	DisplayLogin string `json:"display_login,omitempty"`
	GravatarID   string `json:"gravatar_id,omitempty"`
	URL          string `json:"url,omitempty"`
	AvatarURL    string `json:"avatar_url,omitempty"`
}

type Payload struct {
	Action       string      `json:"action,omitempty"`
	RepositoryID int         `json:"repository_id,omitempty"`
	PushID       int64       `json:"push_id,omitempty"`
	Size         int         `json:"size,omitempty"`
	DistinctSize int         `json:"distinct_size,omitempty"`
	Ref          string      `json:"ref,omitempty"`
	Head         string      `json:"head,omitempty"`
	Before       string      `json:"before,omitempty"`
	Commits      []Commit    `json:"commits,omitempty"`
	Issue        Issue       `json:"issue,omitempty"`
	PullRequest  PullRequest `json:"pull_request,omitempty"`
}

type Commit struct {
	Sha      string `json:"sha"`
	Author   Author `json:"author"`
	Message  string `json:"message"`
	Distinct bool   `json:"distinct"`
	ApiURL   string `json:"url"`
}

type PullRequest struct {
	User User `json:"user"`
}

type Issue struct {
	User User `json:"user"`
}

type Author struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Repo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Organization struct {
	ID         int    `json:"id"`
	Login      string `json:"login"`
	GravatarID string `json:"gravatar_id"`
	URL        string `json:"url"`
	AvatarURL  string `json:"avatar_url"`
}

type EventError struct {
	Message          string `json:"message"`
	DocumentationURL string `json:"documentation_url"`
	Status           string `json:"status"`
}
