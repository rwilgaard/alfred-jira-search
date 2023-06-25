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
    Update     bool
    Cache      bool
    Projects   bool
    Issuetypes bool
    Auth       bool
}

func init() {
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.Cache, "cache", false, "cache spaces")
    cli.BoolVar(&opts.Projects, "projects", false, "search projects")
    cli.BoolVar(&opts.Issuetypes, "issuetypes", false, "search issuetypes")
    cli.BoolVar(&opts.Auth, "auth", false, "authenticate")
}
