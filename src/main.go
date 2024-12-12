package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/andygrunwald/go-jira"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
)

type workflowConfig struct {
	URL                  string `env:"jira_url"`
	Username             string `env:"username"`
	AltIcons             bool   `env:"alt_icons"`
	JiraTogglIntegration bool   `env:"jira_toggl_integration"`
	CancelModifierKey    string `env:"cancel_modifier_key"`
	APIToken             string
}

const (
	repo               = "rwilgaard/alfred-jira-search"
	keychainAccount    = "alfred-jira-search"
	updateJobName      = "checkForUpdates"
	projectCacheName   = "projects.json"
	issuetypeCacheName = "issuetypes.json"
	statusCacheName    = "status.json"
	maxCacheAge        = 168 * time.Hour
)

var (
	wf             *aw.Workflow
	cfg            *workflowConfig
	projectCache   []Project
	issuetypeCache []Issuetype
	statusCache    []Status
)

func init() {
	wf = aw.New(
		update.GitHub(repo),
		aw.AddMagic(magicAuth{wf}),
	)
}

func refreshCache(api *jira.Client) error {
	wf.Configure(aw.TextErrors(true))
	log.Println("[main] fetching projects...")
	projects, err := getProjects(api)
	if err != nil {
		return err
	}
	if err := wf.Cache.StoreJSON(projectCacheName, projects); err != nil {
		return err
	}
	log.Println("[main] cached projects")

	log.Println("[main] fetching issuetypes...")
	issuetypes, err := getAllIssuetypes(api)
	if err != nil {
		return err
	}
	if err := wf.Cache.StoreJSON(issuetypeCacheName, issuetypes); err != nil {
		return err
	}
	log.Println("[main] cached issuetypes")

	log.Println("[main] fetching status...")
	status, err := getStatus(api)
	if err != nil {
		return err
	}
	if err := wf.Cache.StoreJSON(statusCacheName, status); err != nil {
		return err
	}
	log.Println("[main] cached status")

	return nil
}

func run() {
	if err := cli.Parse(wf.Args()); err != nil {
		wf.FatalError(err)
	}
	opts.Query = cli.Arg(0)

	if opts.Update {
		wf.Configure(aw.TextErrors(true))
		log.Println("Checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		log.Println("Running update check in background...")
		cmd := exec.Command(os.Args[0], "-update")
		if err := wf.RunInBackground(updateJobName, cmd); err != nil {
			log.Printf("Error starting update check: %s", err)
		}
	}

	cfg = &workflowConfig{}
	if err := wf.Config.To(cfg); err != nil {
		panic(err)
	}

	if opts.Auth {
		runAuth()
	}

	parsedQuery := parseQuery(opts.Query)

	if a := autocomplete(opts.Query); a != "" {
		if err := wf.Alfred.RunTrigger(a, fmt.Sprintf("%s;%s", opts.Query, strings.Join(parsedQuery.Projects, ","))); err != nil {
			wf.FatalError(err)
		}
		return
	}

	token, err := wf.Keychain.Get(keychainAccount)
	if err != nil {
		wf.NewItem("You're not logged in.").
			Subtitle("Press ⏎ to authenticate").
			Icon(aw.IconInfo).
			Var("action", "auth").
			Arg(" ").
			Valid(true)
		wf.SendFeedback()
		return
	}
	cfg.APIToken = token

	tp := jira.BasicAuthTransport{
		Username: cfg.Username,
		Password: cfg.APIToken,
	}

	api, err := jira.NewClient(tp.Client(), cfg.URL)
	if err != nil {
		wf.FatalError(err)
	}

	if wf.Cache.Expired(projectCacheName, maxCacheAge) || wf.Cache.Expired(issuetypeCacheName, maxCacheAge) || wf.Cache.Expired(statusCacheName, maxCacheAge) {
		if err := refreshCache(api); err != nil {
			wf.FatalError(err)
		}
	}

	if opts.GetProjects {
		runGetProjects()
		wf.SendFeedback()
		return
	}

	if opts.GetStatus {
		runGetStatus()
		wf.SendFeedback()
		return
	}

	if opts.GetIssuetypes {
		runGetAllIssuetypes()
		wf.SendFeedback()
		return
	}

	if opts.GetAssignees {
		runGetAssignees(api)
		wf.SendFeedback()
		return
	}

	if opts.JQL {
        wf.NewItem("Edit JQL").
            Var("action", "jql-edit").
            Icon(getIcon("edit")).
            Valid(true)
		runSearch(api, nil, opts.Query, 50)
		wf.SendFeedback()
		return
	}

	if opts.Create {
		issueKey, err := createIssue(api, opts.Query, opts.Issuetype, opts.Project)
		if err != nil {
			wf.FatalError(err)
		}

		av := aw.NewArgVars()
		av.Var("message", fmt.Sprintf("%s created!", issueKey))
		av.Arg(issueKey)
		if err := av.Send(); err != nil {
			panic(err)
		}

		return
	}

	if opts.MyIssues {
		runMyIssues(api)
		wf.SendFeedback()
		return
	}

	if parsedQuery.IssueKey == "" {
		time.Sleep(500 * time.Millisecond)
	}

	runSearch(api, parsedQuery, "", 15)

	if wf.IsEmpty() {
		wf.NewItem("No results found...").
			Subtitle("Try a different query?").
			Icon(aw.IconInfo)
	}
	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
