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

var Token string

func redirectPolicyFunc(req *http.Request, via []*http.Request) error {
    req.Header.Add("Authorization", Token)
    return nil
}

func SetAuthToken(token string) {
    Token = token
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

    client := &http.Client{
        CheckRedirect: redirectPolicyFunc,
    }

    q := url.QueryEscape(strings.Join(terms, " "))
    fmt.Printf("Sending query to %s\n", IssuesURL + "?q=" + q)
    dest := IssuesURL + "?q=" + q
    req, err := http.NewRequest("GET", dest, nil)
    if err != nil {
        return nil, err
    }

    bearer := "Bearer " + Token
    fmt.Printf("auth: %s\n", bearer)
    //req.Header.Add("Authorization", bearer)
    req.Header.Add("Accept", "application/vnd.github+json")
    req.Header.Add("X-Github-Api-Version", "2022-11-28")
    resp, err := client.Do(req)
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
func CreateIssue(repo string, issue Issue) error {

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
    bearer := fmt.Sprintf("Bearer %s", Token)
    req.Header.Add("Authorization", bearer)

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
