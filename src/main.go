package main

import (
	"github.com/andygrunwald/go-jira"
	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
)

type workflowConfig struct {
    URL      string `env:"jira_url"`
    Username string `env:"username"`
    APIToken string
}

const (
    repo            = "rwilgaard/alfred-jira-search"
    updateJobName   = "checkForUpdates"
    keychainAccount = "alfred-jira-search"
)

var (
    wf          *aw.Workflow
    cfg         *workflowConfig
    // cacheName   = "projects.json"
    // maxCacheAge = 24 * time.Hour
    // spaceCache  []
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
