package main

import (
    "log"
    "time"
    "os"
    "os/exec"

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
    updateJobName   = "checkForUpdates"
    keychainAccount = "alfred-jira-search"
)

var (
    wf           *aw.Workflow
    cfg          *workflowConfig
    cacheName    = "projects.json"
    maxCacheAge  = 24 * time.Hour
    projectCache []Project
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

    if opts.Projects {
        runProjects()
        if len(opts.Query) > 0 {
            wf.Filter(opts.Query)
        }
        wf.SendFeedback()
        return
    }

    if a := autocomplete(opts.Query); a != "" {
        if err := wf.Cache.StoreJSON("prev_query", opts.Query); err != nil {
            wf.FatalError(err)
        }
        if err := wf.Alfred.RunTrigger(a, ""); err != nil {
            wf.FatalError(err)
        }
        return
    }

    token, err := wf.Keychain.Get(keychainAccount)
    if err != nil {
        wf.NewItem("You're not logged in.").
            Subtitle("Press ⏎ to authenticate").
            Icon(aw.IconInfo).
            Arg("auth").
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

    if opts.Cache {
        wf.Configure(aw.TextErrors(true))
        log.Println("[main] fetching projects...")
        projects, err := getProjects(api)
        if err != nil {
            wf.FatalError(err)
        }
        if err := wf.Cache.StoreJSON(cacheName, projects); err != nil {
            wf.FatalError(err)
        }
        log.Println("[main] cached projects")
        return
    }

    if wf.Cache.Expired(cacheName, maxCacheAge) {
        wf.Rerun(0.3)
        if !wf.IsRunning("cache") {
            log.Println("[main] fetching projects...")
            cmd := exec.Command(os.Args[0], "-cache")
            if err := wf.RunInBackground("cache", cmd); err != nil {
                wf.FatalError(err)
            } else {
                log.Printf("cache job already running.")
            }
        }
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
