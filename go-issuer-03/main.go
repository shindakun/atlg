package issuer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/google/go-github/v25/github"
	"github.com/shindakun/envy"
	"golang.org/x/oauth2"
)

// RepoOwner is the owner of the repo we want to open an issue in
var RepoOwner string

// IssueRepo is the repo we want to open this new issue in.
var IssueRepo string

// ProjectColumn is the TODO column number of the project we want to add the issue to
var ProjectColumn int64

// Token is the GitHub Personal Access Token
var Token string

// Secret is used to validate payloads
var Secret string

// Payload of GitHub webhook
type Payload struct {
	Action string `json:"action"`
	Issue  struct {
		URL           string `json:"url"`
		RepositoryURL string `json:"repository_url"`
		Number        int    `json:"number"`
		Title         string `json:"title"`
		Body          string `json:"body"`
	} `json:"issue"`
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
}

// HandleWebhook is our entry point, it recieves the POST from GitHub
func HandleWebhook(res http.ResponseWriter, req *http.Request) {
	var Payload Payload
	defer req.Body.Close()

	p, err := github.ValidatePayload(req, []byte(Secret))
	if err != nil {
		http.Error(res, "bad request: "+err.Error(), 400)
		log.Printf("bad request: %v", err.Error())
		return
	}

	err = json.Unmarshal(p, &Payload)
	if err != nil {
		http.Error(res, "bad request: "+err.Error(), 400)
		log.Printf("bad request: %v", err.Error())
		return
	}

	err = createNewIssue(&Payload)
	if err != nil {
		log.Printf("bad request: %v", err.Error())
		return
	}
}

func createNewIssue(p *Payload) error {
	log.Printf("Creating New Issue.\n")
	log.Printf("  Name: %#v\n", p.Repository.Name)
	log.Printf("  Title: %#v\n", p.Issue.Title)
	log.Printf("  Body: %#v\n", p.Issue.Body)
	log.Printf("  URL: %#v\n", p.Issue.URL)

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: Token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	title := fmt.Sprintf("[%s] %s", p.Repository.Name, p.Issue.Title)
	body := fmt.Sprintf("%s\n%s/%s#%d", p.Issue.Body, RepoOwner, p.Repository.Name, p.Issue.Number)

	issue := &github.IssueRequest{
		Title: &title,
		Body:  &body,
	}

	ish, _, err := client.Issues.Create(ctx, RepoOwner, IssueRepo, issue)
	if err != nil {
		log.Printf("error: %v", err)
		return err
	}

	id := *ish.ID
	card := &github.ProjectCardOptions{
		ContentID:   id,
		ContentType: "Issue",
	}

	_, _, err = client.Projects.CreateProjectCard(ctx, ProjectColumn, card)
	if err != nil {
		log.Printf("error: %v", err)
		return err
	}

	return nil
}

func init() {
	log.Println("Issuer")
	var err error
	Token, err = envy.Get("GITHUBTOKEN")
	if err != nil || Token == "" {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	Secret, err = envy.Get("SECRET")
	if err != nil || Secret == "" {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	RepoOwner, err = envy.Get("REPOOWNER")
	if err != nil || RepoOwner == "" {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	IssueRepo, err = envy.Get("ISSUEREPO")
	if err != nil || IssueRepo == "" {
		log.Printf("error: %v", err)
		os.Exit(1)
	}

	n, err := envy.Get("PROJECTCOLUMN")
	if err != nil || n == "" {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
	ProjectColumn, err = strconv.ParseInt(n, 10, 64)
	if err != nil || ProjectColumn == 0 {
		log.Printf("error: %v", err)
		os.Exit(1)
	}
}
