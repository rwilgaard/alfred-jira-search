package main

import (
	"github.com/andygrunwald/go-jira"
)

func getIssues(api *jira.Client, jql string) ([]jira.Issue, error) {
    defaultOrder := " ORDER BY created DESC"
    opts := jira.SearchOptions{
        Fields: []string{
            "key",
            "summary",
            "issuetype",
            "project",
            "status",
            "assignee",
            "created",
            "customfield_15636",
        },
        MaxResults: 15,
    }
    issues, _, err := api.Issue.Search(jql+defaultOrder, &opts)
    if err != nil {
        return nil, err
    }

    return issues, nil
}
