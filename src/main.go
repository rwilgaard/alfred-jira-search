package main

import (
    "fmt"
    "log"
    "os"
    "os/exec"
    "time"

    "github.com/andygrunwald/go-jira"
    aw "github.com/deanishe/awgo"
    "github.com/deanishe/awgo/update"
)

type workflowConfig struct {
    URL      string `env:"jira_url"`
    Username string `env:"username"`
    AltIcons bool   `env:"alt_icons"`
    APIToken string
}

const (
    repo            = "rwilgaard/alfred-jira-search"
    keychainAccount = "alfred-jira-search"
    // updateJobName   = "checkForUpdates"
)

var (
    wf                 *aw.Workflow
    cfg                *workflowConfig
    projectCacheName   = "projects.json"
    issuetypeCacheName = "issuetypes.json"
    maxCacheAge        = 24 * time.Hour
    projectCache       []Project
    issuetypeCache     []Issuetype
)

func init() {
    wf = aw.New(
        update.GitHub(repo),
    )
}

func run() {
    if err := cli.Parse(wf.Args()); err != nil {
        wf.FatalError(err)
    }
    opts.Query = cli.Arg(0)

    cfg = &workflowConfig{}
    if err := wf.Config.To(cfg); err != nil {
        panic(err)
    }

    if opts.Auth {
        runAuth()
    }

    if opts.GetProjects {
        runGetProjects()
        if len(opts.Query) > 0 {
            wf.Filter(opts.Query)
        }
        wf.SendFeedback()
        return
    }

    if a := autocomplete(opts.Query); a != "" {
        if err := wf.Alfred.RunTrigger(a, opts.Query); err != nil {
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

    if opts.GetIssuetypes {
        if opts.Project != "" {
            runGetProjectIssuetypes(api, opts.Project)
        } else {
            runGetAllIssuetypes(api)
            if opts.Query != "" {
                wf.Filter(opts.Query)
            }
        }
        wf.SendFeedback()
        return
    }

    if opts.Cache {
        wf.Configure(aw.TextErrors(true))
        log.Println("[main] fetching projects...")
        projects, err := getProjects(api)
        if err != nil {
            wf.FatalError(err)
        }
        if err := wf.Cache.StoreJSON(projectCacheName, projects); err != nil {
            wf.FatalError(err)
        }
        log.Println("[main] cached projects")

        log.Println("[main] fetching issuetypes...")
        issuetypes, err := getAllIssuetypes(api)
        if err != nil {
            wf.FatalError(err)
        }
        if err := wf.Cache.StoreJSON(issuetypeCacheName, issuetypes); err != nil {
            wf.FatalError(err)
        }
        log.Println("[main] cached issuetypes")
        return
    }

    if wf.Cache.Expired(projectCacheName, maxCacheAge) {
        wf.Rerun(0.3)
        if !wf.IsRunning("cache") {
            log.Println("[main] refreshing cache...")
            cmd := exec.Command(os.Args[0], "-cache")
            if err := wf.RunInBackground("cache", cmd); err != nil {
                wf.FatalError(err)
            } else {
                log.Printf("cache job already running.")
            }
        }
    }

    if opts.Create {
        issueKey, err := createIssue(api, opts.Query, opts.Issuetype, opts.Project)
        if err != nil {
            wf.FatalError(err)
        }

        av := aw.NewArgVars()
        av.Var("message", fmt.Sprintf("%s created!", issueKey))
        if err := av.Send(); err != nil {
            panic(err)
        }

        return
    }

    runSearch(api)

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
