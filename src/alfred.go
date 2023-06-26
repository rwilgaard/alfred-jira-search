package main

import (
	"fmt"

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
    jql, _ := parseQuery(opts.Query)
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
            Icon(getIcon(issue.Fields.Type.Name)).
            Valid(true)
    }
}

func runGetProjects() {
    if wf.Cache.Exists(cacheName) {
        if err := wf.Cache.LoadJSON(cacheName, &projectCache); err != nil {
            wf.FatalError(err)
        }
    }

    var prevQuery string
    if err := wf.Cache.LoadJSON("prev_query", &prevQuery); err != nil {
        wf.FatalError(err)
    }

    wf.NewItem("Cancel").
        Arg("cancel").
        Valid(true)

    for _, s := range projectCache {
        wf.NewItem(s.Key).
            Match(fmt.Sprintf("%s %s", s.Key, s.Name)).
            Subtitle(s.Name).
            Arg(prevQuery + s.Key + " ").
            Var("project", s.Key).
            Valid(true)
    }
}

func runGetIssuetypes() {
    if wf.Cache.Exists(issuetypesCacheName) {
        if err := wf.Cache.LoadJSON(issuetypesCacheName, &issuetypeCache); err != nil {
            wf.FatalError(err)
        }
    }

    var prevQuery string
    if err := wf.Cache.LoadJSON("prev_query", &prevQuery); err != nil {
        wf.FatalError(err)
    }

    wf.NewItem("Cancel").
        Arg("cancel").
        Valid(true)

    for _, i := range issuetypeCache {
        wf.NewItem(i.Name).
            UID(i.Name).
            Arg(prevQuery + i.Name + " ").
            Var("issuetype", i.Name).
            Valid(true)
    }
}
