// ghicli provides a command line interface to create, read, modify, and delete GitHub issues
package main

import (
	"bytes"
	"flag"
	"fmt"
	"ghicli/github"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
)

func main() {

    repoPtr := flag.String("repo", "golang/go", "repo for which issues are queried (default: golang/go)")
    actionPtr := flag.String("action", "list", "action to perform (create, list, fetch, modify, delete) (default: list)")
    urlPtr := flag.String("url", "https://api.github.com/", "url to query (default: https://api.github.com/)")
    userPtr := flag.String("user", "ur mom", "github username (default: ur mom)")
    flag.Parse()

    if !strings.Contains(*repoPtr, "/") {
        log.Fatal("Invalid repo owner/name provided")
    }

    if *userPtr == "ur mom" {
        log.Fatal("ur mom can't code!")
    }

    github.Username = *userPtr

    switch *actionPtr {
    case "list":
        url := []byte(*urlPtr)
        issuesPath := []byte("search/issues")
        url = append(url, issuesPath...)
        fmt.Printf("querying URL %s\n", url)
        github.ListIssues(*repoPtr, flag.Args())

    case "create":
        issue := createIssueUsingEditor()
        if err := github.CreateIssue(*repoPtr, issue); err != nil {
            log.Fatal(err)
        }

    case "fetch":
        log.Fatal("not currently supported")

    case "modify":
        log.Fatal("not currently supported")

    case "delete":
        log.Fatal("not currently supported")

    default:
        log.Fatal("Invalid action provided")

    }
}

func createIssueUsingEditor() github.IssueRequest {

    filepath := os.TempDir() + "/issue.txt"
    f, err := os.Create(filepath)
    if err != nil {
        log.Fatal(err)
    }

    f.WriteString("Title: \n")
    f.WriteString("Body: \n")

    f.Close()

    fmt.Println("Enter the binary for the editor you would like to use (nvim, nano, etc.):")
    var editor string
    fmt.Scanln(&editor)

    cmd := exec.Command(editor, filepath)
    cmd.Stdin = os.Stdin
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    if err := cmd.Start(); err != nil {
        log.Fatal(err)
    }

    if err = cmd.Wait(); err != nil {
        log.Fatal(err)
    }

    log.Printf("%s successfully edited\n", filepath)

    f, err = os.Open(filepath)
    defer f.Close()
    if err != nil {
        log.Fatal(nil)
    }

    content, err := io.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }

    prefix := []byte("Title: ")
    after, found := bytes.CutPrefix(content, prefix)
    if !found {
        fmt.Fprintf(os.Stderr, "%s tag not found in file %s", prefix, filepath)
        os.Exit(1)
    }

    bodyTag := []byte("Body: ")
    if !bytes.Contains(after, bodyTag) {
        fmt.Fprintf(os.Stderr, "%s tag not found in file %s", bodyTag, filepath)
        os.Exit(1)
    }

    tokens := bytes.SplitAfter(after, bodyTag)

    tokens[0] = tokens[0][:len(tokens[0]) - 7]
    tokens[1] = tokens[1][:len(tokens[1]) - 1]

    var issue github.IssueRequest;
    issue.Title = string(tokens[0])
    fmt.Printf("Found title: %s\n", issue.Title)
    issue.Body = string(tokens[1])
    fmt.Printf("Found body: %s\n", issue.Body)

    return issue
}
