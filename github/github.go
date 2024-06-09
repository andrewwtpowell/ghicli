// Provides Go API for the GitHub issue tracker
package github

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const IssuesURL = "https://api.github.com/search/issues"
const BaseURL = "https://api.github.com/"

var Username string

type IssuesSearchResult struct {
    TotalCount  int `json:"total_count"`
    Items       []*Issue
}

type Issue struct {
    Number      int
    HTMLURL     string `json:"html_url"`
    Title       string
    State       string
    User        *User
    CreatedAt   time.Time `json:"created_at"`
    Body        string
}

type IssueRequest struct {
    Title       string
    Body        string
    Assignees   []string
    State       string
}

type User struct {
    Login   string
    HTMLURL string `json:html_url"`
}

// SearchIssues queries the GitHub Issue Tracker
func SearchIssues(repo string, terms []string) (*IssuesSearchResult, error) {

    repo = "repo:" + repo
    if len(terms) > 0 {
        terms = append(terms[:1], terms[0:]...)
        terms[0] = repo
    } else {
        terms = append(terms, repo)
    }

    q := url.QueryEscape(strings.Join(terms, " "))
    fmt.Printf("Sending query to %s\n", IssuesURL + "?q=" + q)
    dest := IssuesURL + "?q=" + q
    req, err := http.NewRequest("GET", dest, nil)
    if err != nil {
        return nil, err
    }

    req.Header.Add("Accept", "application/vnd.github+json")
    req.Header.Add("X-Github-Api-Version", "2022-11-28")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return nil, err
    }

    defer resp.Body.Close()
        
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("search query failed: %s\n", resp.Status)
    }

    var result IssuesSearchResult
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result, nil
}

// List issues queried from the GitHub Issue Tracker
func ListIssues(repo string, terms []string) {

    result, err := SearchIssues(repo, terms)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%d issues:\n", result.TotalCount)
    for _, item := range result.Items {
        fmt.Printf("#%-5d %9.9s %.55s\n", item.Number, item.User.Login, item.Title)
    }
}

// Create issue in the GitHub Issue Tracker
func CreateIssue(repo string, issue IssueRequest) error {

    var terms []string
    terms = append(terms, issue.Title)
    result, err := SearchIssues(repo, terms)
    if err != nil {
        return err
    }

    if result.TotalCount > 0 {
        return errors.New("Issue already exists")
    }

    issue.State = "open"
    issue.Assignees = append(issue.Assignees, Username)

    issueJSON, err := json.Marshal(issue)
    if err != nil {
        return err
    }

    fmt.Printf("Marshalled issue: \n%s\n", issueJSON)

    dest := BaseURL + "repos/" + repo + "/issues"
    fmt.Printf("Making request to %s\n", dest)
    req, err := http.NewRequest("POST", dest, bytes.NewBuffer(issueJSON))
    if err != nil {
        return err
    }

    req.Header.Add("Accept", "application/json")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }

    body, err := io.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil {
        return err
    }

    fmt.Printf("Received response %s\n", body)

    var res Issue
    err = json.Unmarshal([]byte(body), &res)
    if err != nil {
        return err
    }

    if issue.Title != res.Title {
        return errors.New("Response title not equivalent to posted title")
    }

    fmt.Printf("Created issue at %s\n", dest)
    fmt.Printf("#%-5d %9.9s %.55s\n", res.Number, res.User.Login, res.Title)

    return nil
}
