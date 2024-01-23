package main

import (
    "fmt"
    "log"
    "strings"

    "github.com/andygrunwald/go-jira"
)

type Project struct {
    Key  string
    Name string
}

type Issuetype struct {
    ID   string `json:"id"`
    Name string `json:"name"`
}

type Status struct {
    ID   string
    Name string
}

func autocomplete(query string) string {
    for _, w := range strings.Split(query, " ") {
        switch w {
        case "@":
            return "project"
        case "#":
            return "issuetype"
        case "?":
            return "status"
        case "%":
            return "assignee"
        }
    }
    return ""
}

func testAuthentication(api *jira.Client) (statusCode int, err error) {
    _, resp, err := api.User.GetSelf()
    if err != nil {
        return resp.StatusCode, err
    }
    return resp.StatusCode, nil
}

func buildJQL(query *parsedQuery) (jql string) {
    defaultOrder := " ORDER BY created DESC"
    if query.IssueKey != "" {
        jql = fmt.Sprintf("key = '%s'", query.IssueKey)
        return jql
    }

    if strings.TrimSpace(query.Text) != "" {
        jql = fmt.Sprintf("text ~ '%s'", strings.TrimSpace(query.Text))
    }

    if len(query.Projects) > 0 {
        if jql != "" {
            jql += " AND "
        }
        jql += fmt.Sprintf("project in (%s)", "'"+strings.Join(query.Projects, "','")+"'")
    }

    if len(query.Issuetypes) > 0 {
        if jql != "" {
            jql += " AND "
        }
        jql += fmt.Sprintf("issuetype in (%s)", "'"+strings.Join(query.Issuetypes, "','")+"'")
    }

    if len(query.Status) > 0 {
        if jql != "" {
            jql += " AND "
        }
        jql += fmt.Sprintf("status in (%s)", "'"+strings.Join(query.Status, "','")+"'")
    }

    if len(query.Assignees) > 0 {
        if jql != "" {
            jql += " AND "
        }
        jql += fmt.Sprintf("assignee in (%s)", "'"+strings.Join(query.Assignees, "','")+"'")
    }

    return jql+defaultOrder
}

func getAssignableUsers(api *jira.Client, query string, projects string) ([]jira.User, error) {
    users := new([]jira.User)
    u := fmt.Sprintf("/rest/api/2/user/assignable/multiProjectSearch?projectKeys=%s&maxResults=25&username=%s", projects, query)
    req, _ := api.NewRequest("GET", u, nil)
    _, err := api.Do(req, users)
    if err != nil {
        return nil, err
    }
    return *users, nil
}

func getIssues(api *jira.Client, jql string, maxResults int) ([]jira.Issue, error) {
    opts := jira.SearchOptions{
        Fields: []string{
            "key",
            "summary",
            "issuetype",
            "project",
            "status",
            "assignee",
            "created",
            "updated",
            "customfield_15636",
        },
        Expand:     "renderedFields",
        MaxResults: maxResults,
    }
    log.Println("Running Search!")
    issues, _, err := api.Issue.Search(jql, &opts)
    if err != nil {
        return nil, err
    }

    return issues, nil
}

func getProjects(api *jira.Client) ([]Project, error) {
    var projects []Project

    opts := jira.GetQueryOptions{Fields: "key,name"}
    pl, _, err := api.Project.ListWithOptions(&opts)
    if err != nil {
        return nil, err
    }

    for _, p := range *pl {
        project := Project{Key: p.Key, Name: p.Name}
        projects = append(projects, project)
    }

    return projects, nil
}

func getStatus(api *jira.Client) ([]Status, error) {
    var status []Status

    sl, _, err := api.Status.GetAllStatuses()
    if err != nil {
        return nil, err
    }

    for _, s := range sl {
        st := Status{ID: s.ID, Name: s.Name}
        status = append(status, st)
    }

    return status, nil
}

func getAllIssuetypes(api *jira.Client) ([]Issuetype, error) {
    issuetypes := new([]Issuetype)

    req, _ := api.NewRequest("GET", "/rest/api/2/issuetype", nil)
    _, err := api.Do(req, issuetypes)
    if err != nil {
        return nil, err
    }

    return *issuetypes, nil
}

func createIssue(api *jira.Client, summary string, issuetype string, project string) (issueKey string, error error) {
    i := jira.Issue{
        Fields: &jira.IssueFields{
            Assignee: &jira.User{
                Name: cfg.Username,
            },
            Reporter: &jira.User{
                Name: cfg.Username,
            },
            Type: jira.IssueType{
                Name: issuetype,
            },
            Project: jira.Project{
                Key: project,
            },
            Summary: summary,
        },
    }

    issue, _, err := api.Issue.Create(&i)
    if err != nil {
        return "", err
    }

    return issue.Key, nil
}
