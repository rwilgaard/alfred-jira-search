package main

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/andygrunwald/go-jira"
    aw "github.com/deanishe/awgo"
    "github.com/ncruces/zenity"
)

type parsedQuery struct {
    IssueKey   string
    Text       string
    Projects   []string
    Issuetypes []string
    Status     []string
    Assignees  []string
}

type magicAuth struct {
    wf *aw.Workflow
}

func (a magicAuth) Keyword() string     { return "clearauth" }
func (a magicAuth) Description() string { return "Clear credentials" }
func (a magicAuth) RunText() string     { return "Credentials cleared!" }
func (a magicAuth) Run() error          { return clearAuth() }

func runAuth() {
    _, pwd, err := zenity.Password(
        zenity.Title(fmt.Sprintf("Enter API Token for %s", cfg.Username)),
    )
    if err != nil {
        wf.FatalError(err)
    }

    tp := jira.BasicAuthTransport{
        Username: cfg.Username,
        Password: pwd,
    }
    api, err := jira.NewClient(tp.Client(), cfg.URL)
    if err != nil {
        wf.FatalError(err)
    }

    sc, err := testAuthentication(api)
    if err != nil {
        zerr := zenity.Error(
            fmt.Sprintf("Error authenticating: HTTP %d", sc),
            zenity.ErrorIcon,
        )
        if zerr != nil {
            wf.FatalError(err)
        }
        wf.FatalError(err)
    }

    if err := wf.Keychain.Set(keychainAccount, pwd); err != nil {
        wf.FatalError(err)
    }
}

func parseQuery(query string) *parsedQuery {
    q := new(parsedQuery)
    issueKeyRegex := regexp.MustCompile("^[a-zA-Z]+-[0-9]+$")
    projectRegex := regexp.MustCompile(`^@\w+`)
    issuetypeRegex := regexp.MustCompile(`^#\w+`)
    statusRegex := regexp.MustCompile(`^\?\w+`)
    assigneeRegex := regexp.MustCompile(`^%\w+`)

    if opts.Project != "" {
        q.Projects = append(q.Projects, opts.Project)
    }

    for _, w := range strings.Split(query, " ") {
        switch {
        case issueKeyRegex.MatchString(w):
            q.IssueKey = w
        case projectRegex.MatchString(w):
            s := strings.ReplaceAll(w[1:], "_", " ")
            q.Projects = append(q.Projects, s)
        case issuetypeRegex.MatchString(w):
            s := strings.ReplaceAll(w[1:], "_", " ")
            q.Issuetypes = append(q.Issuetypes, s)
        case statusRegex.MatchString(w):
            s := strings.ReplaceAll(w[1:], "_", " ")
            q.Status = append(q.Status, s)
        case assigneeRegex.MatchString(w):
            s := strings.ReplaceAll(w[1:], "_", " ")
            q.Assignees = append(q.Assignees, s)
        default:
            q.Text = q.Text + w + " "
        }
    }

    return q
}

func runSearch(api *jira.Client, query *parsedQuery) {
    jql := buildJQL(query)
    issues, err := getIssues(api, jql, 15)
    if err != nil {
        wf.FatalError(err)
    }

    for _, issue := range issues {
        i := wf.NewItem(issue.Key).
            Subtitle(fmt.Sprintf("%s  •  %s", issue.Fields.Status.Name, issue.Fields.Summary)).
            Arg(issue.Key).
            Var("action", "open").
            Var("issuekey", issue.Key).
            Var("item_url", fmt.Sprintf("%s/browse/%s", cfg.URL, issue.Key)).
            Icon(getIcon(issue.Fields.Type.Name)).
            Valid(true)

        assignee := "Unassigned"
        if issue.Fields.Assignee != nil {
            assignee = string(issue.Fields.Assignee.DisplayName)
        }

        i.NewModifier(aw.ModOpt).
            Subtitle(fmt.Sprintf("Assignee: %s  •  Created: %s  •  Updated: %s", assignee, issue.RenderedFields.Created, issue.RenderedFields.Updated))

        i.NewModifier(aw.ModCtrl).
            Subtitle("Open search results in Jira").
            Var("action", "jql").
            Var("jql", jql)

        if cfg.JiraTogglIntegration {
            i.NewModifier(aw.ModCmd).
                Subtitle("Start Toggl").
                Var("action", "toggl").
                Var("issue_key", issue.Key)
        }
    }
}

func runMyIssues(api *jira.Client) {
    jql := fmt.Sprintf("assignee = '%s' AND resolution = Unresolved ", cfg.Username)
    issues, err := getIssues(api, jql, 9999)
    if err != nil {
        wf.FatalError(err)
    }

    for _, issue := range issues {
        i := wf.NewItem(issue.Key).
            Subtitle(fmt.Sprintf("%s  •  %s", issue.Fields.Status.Name, issue.Fields.Summary)).
            Match(fmt.Sprintf("%s %s", issue.Key, issue.Fields.Summary)).
            Arg(issue.Key).
            Var("action", "open").
            Var("issuekey", issue.Key).
            Var("item_url", fmt.Sprintf("%s/browse/%s", cfg.URL, issue.Key)).
            Icon(getIcon(issue.Fields.Type.Name)).
            Valid(true)

        i.NewModifier(aw.ModOpt).
            Subtitle(fmt.Sprintf("Created: %s  •  Updated: %s", issue.RenderedFields.Created, issue.RenderedFields.Updated))

        if cfg.JiraTogglIntegration {
            i.NewModifier(aw.ModCmd).
                Subtitle("Start Toggl").
                Var("action", "toggl").
                Var("issue_key", issue.Key)
        }
    }
}

func runGetProjects() {
    if wf.Cache.Exists(projectCacheName) {
        if err := wf.Cache.LoadJSON(projectCacheName, &projectCache); err != nil {
            wf.FatalError(err)
        }
    }

    prevQuery, _ := wf.Config.Env.Lookup("prev_query")

    for _, s := range projectCache {
        i := wf.NewItem(s.Key).
            UID(s.Key).
            Match(fmt.Sprintf("%s %s", s.Key, s.Name)).
            Subtitle(s.Name).
            Arg(prevQuery+s.Key+" ").
            Var("project", strings.ReplaceAll(s.Key, " ", "_")).
            Var("project_raw", s.Key).
            Valid(true)
        i.NewModifier(aw.ModCmd).
            Subtitle("Cancel").
            Arg("cancel")
    }
}

func runGetStatus() {
    if wf.Cache.Exists(statusCacheName) {
        if err := wf.Cache.LoadJSON(statusCacheName, &statusCache); err != nil {
            wf.FatalError(err)
        }
    }

    prevQuery, _ := wf.Config.Env.Lookup("prev_query")

    for _, s := range statusCache {
        status := strings.ReplaceAll(s.Name, " ", "_")
        i := wf.NewItem(s.Name).
            UID(s.ID).
            Match(s.Name).
            Arg(prevQuery+status+" ").
            Var("status", status).
            Var("status_raw", s.Name).
            Valid(true)
        i.NewModifier(aw.ModCmd).
            Subtitle("Cancel").
            Arg("cancel")
    }
}

func runGetAllIssuetypes() {
    if wf.Cache.Exists(issuetypeCacheName) {
        if err := wf.Cache.LoadJSON(issuetypeCacheName, &issuetypeCache); err != nil {
            wf.FatalError(err)
        }
    }

    prevQuery, _ := wf.Config.Env.Lookup("prev_query")

    for _, i := range issuetypeCache {
        name := strings.ReplaceAll(i.Name, " ", "_")
        i := wf.NewItem(i.Name).
            UID(i.Name).
            Arg(prevQuery+name+" ").
            Var("issuetype", name).
            Var("issuetype_raw", i.Name).
            Valid(true)
        i.NewModifier(aw.ModCmd).
            Subtitle("Cancel").
            Arg("cancel")
    }
}

func runGetAssignees(api *jira.Client) {
    prevQuery, _ := wf.Config.Env.Lookup("prev_query")
    var users []jira.User
    var err error

    query := strings.TrimSpace(opts.Query)
    if opts.Projects != "" {
        users, err = getAssignableUsers(api, query, opts.Projects)
    } else {
        users, _, err = api.User.Find("username", jira.WithMaxResults(25), jira.WithUsername(query))
    }

    if err != nil {
        wf.FatalError(err)
    }

    for _, u := range users {
        i := wf.NewItem(u.Name).
            Subtitle(u.DisplayName).
            Arg(prevQuery+u.Name+" ").
            Var("user", u.Name).
            Valid(true)
        i.NewModifier(aw.ModCmd).
            Subtitle("Cancel").
            Arg("cancel")
    }
}

func clearAuth() error {
    if err := wf.Keychain.Delete(keychainAccount); err != nil {
        return err
    }
    return nil
}
