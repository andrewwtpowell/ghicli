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

type User struct {
    Login   string
    HTMLURL string `json:html_url"`
}

// SearchIssues queries the GitHub Issue Tracker
func SearchIssues(repo string, terms []string, token string) (*IssuesSearchResult, error) {

    repoTerm := fmt.Sprintf("repo:%s", repo)
    copy(terms[1:], terms[0:])
    terms[0] = repoTerm

    q := url.QueryEscape(strings.Join(terms, " "))
    fmt.Printf("Sending query to %s\n", IssuesURL + "?q=" + q)
    dest := IssuesURL + "?q=" + q
    req, err := http.NewRequest("GET", dest, nil)
    if err != nil {
        return nil, err
    }

    auth := "Bearer " + token
    req.Header.Add("Authorization", auth)
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
func ListIssues(repo string, terms []string, token string) {

    result, err := SearchIssues(repo, terms, token)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("%d issues:\n", result.TotalCount)
    for _, item := range result.Items {
        fmt.Printf("#%-5d %9.9s %.55s\n", item.Number, item.User.Login, item.Title)
    }
}

// Create issue in the GitHub Issue Tracker
func CreateIssue(repo string, issue Issue, token string) error {

    result, err := SearchIssues(repo, strings.Split(issue.Title, ""))
    if err != nil {
        return err
    }

    if result.TotalCount > 0 {
        return errors.New("Issue already exists")
    }

    issueJSON, err := json.Marshal(issue)
    if err != nil {
        return err
    }

    dest := url.PathEscape(BaseURL + "repos/" + repo + "/issues")
    req, err := http.NewRequest("POST", dest, bytes.NewBuffer(issueJSON))
    if err != nil {
        return err
    }

    req.Header.Add("Accept", "application/json")
    auth := fmt.Sprintf("Bearer %s", token)
    req.Header.Add("Authorization", auth)

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }

    body, err := io.ReadAll(resp.Body)
    defer resp.Body.Close()
    if err != nil {
        return err
    }

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
