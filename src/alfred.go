package main

import (
	"fmt"
	"regexp"

	"github.com/andygrunwald/go-jira"
	"github.com/ncruces/zenity"
)

func runAuth() {
    _, pwd, err := zenity.Password(
        zenity.Title(fmt.Sprintf("Enter API Token for %s", cfg.Username)),
    )
    if err != nil {
        wf.FatalError(err)
    }
    if err := wf.Keychain.Set(keychainAccount, pwd); err != nil {
        wf.FatalError(err)
    }
}

func runSearch(api *jira.Client) {
    var jql string
    issueKeyRegex := regexp.MustCompile("^[a-zA-Z]+-[0-9]+$")

    if issueKeyRegex.MatchString(opts.Query) {
        jql = fmt.Sprintf("key = '%s'", opts.Query)
    } else {
        jql = fmt.Sprintf("text ~ '%s'", opts.Query)
    }

    issues, err := getIssues(api, jql) 
    if err != nil {
        wf.FatalError(err)
    }

    for _, issue := range issues {
        wf.NewItem(issue.Key).
            Arg("open").
            Subtitle(fmt.Sprintf("%s  •  %s", issue.Fields.Status.Name, issue.Fields.Summary)).
            Var("issuekey", issue.Key).
            Var("item_url", fmt.Sprintf("%s/browse/%s", cfg.URL, issue.Key)).
            Valid(true)
    }
}
