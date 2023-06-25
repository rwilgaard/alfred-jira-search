package main

import "flag"

var (
    opts = &options{}
    cli  = flag.NewFlagSet("alfred-jira-search", flag.ContinueOnError)
)

type options struct {
    // Arguments
    Query string

    // Commands
    Update        bool
    Cache         bool
    GetProjects   bool
    GetIssuetypes bool
    Auth          bool
    Create        bool
    Project       string
    Issuetype     string
}

func init() {
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.Cache, "cache", false, "cache spaces")
    cli.BoolVar(&opts.GetProjects, "get-projects", false, "search projects")
    cli.BoolVar(&opts.GetIssuetypes, "get-issuetypes", false, "search issuetypes")
    cli.BoolVar(&opts.Auth, "auth", false, "authenticate")
    cli.BoolVar(&opts.Create, "create", false, "create jira issue")
    cli.StringVar(&opts.Project, "project", "", "project")
    cli.StringVar(&opts.Issuetype, "issuetype", "", "issuetype")
}
