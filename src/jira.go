package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/andygrunwald/go-jira"
	aw "github.com/deanishe/awgo"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

type Project struct {
    Key  string
    Name string
}

type Issuetype struct {
    Id   string `json:"id"`
    Name string `json:"name"`
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

func parseQuery(query string) (string, []string) {
    if wf.Cache.Exists(cacheName) {
        if err := wf.Cache.LoadJSON(cacheName, &projectCache); err != nil {
            wf.FatalError(err)
        }
    }

    var jql string
    var text string
    var projects []string

    issueKeyRegex := regexp.MustCompile("^[a-zA-Z]+-[0-9]+$")
    projectRegex := regexp.MustCompile(`^@\w+`)

    if issueKeyRegex.MatchString(query) {
        jql = fmt.Sprintf("key = '%s'", query)
        return jql, nil
    }

    for _, w := range strings.Split(query, " ") {
        switch {
        case projectRegex.MatchString(w):
            projectKey := w[1:]
            if projectExists(projectKey, projectCache) {
                projects = append(projects, projectKey)
            } else {
                title := fmt.Sprintf("%s project not found...", strings.ToUpper(projectKey))
                s := fuzzy.Find(projectKey, projectCacheToSlice(projectCache))
                sub := fmt.Sprintf("Did you mean %s?", strings.Join(s, ", "))
                wf.NewItem(title).Subtitle(sub).Icon(aw.IconInfo)
            }
        default:
            text = text + fmt.Sprintf("%s ", w)
        }
    }

    if strings.TrimSpace(text) != "" {
        jql = "text ~ '%s'"
        jql = fmt.Sprintf(jql, strings.TrimSpace(text))
    }

    if len(projects) > 0 {
        if jql != "" {
            jql += " AND "
        }
        jql = jql + fmt.Sprintf("project in (%s)", strings.Join(projects, ","))
    }

    return jql, projects
}

func projectExists(key string, projects []Project) bool {
    for _, s := range projects {
        if strings.EqualFold(s.Key, key) {
            return true
        }
    }
    return false
}

func projectCacheToSlice(projects []Project) []string {
    var projectList []string
    for _, p := range projects {
        projectList = append(projectList, p.Key)
    }
    return projectList
}

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

func getIssuetypes(api *jira.Client) (*[]Issuetype, error) {
    issuetypes := new([]Issuetype)

    req, _ := api.NewRequest("GET", "/rest/api/2/issuetype", nil)
    _, err := api.Do(req, issuetypes)
    if err != nil {
        return nil, err
    }

    return issuetypes, nil
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
