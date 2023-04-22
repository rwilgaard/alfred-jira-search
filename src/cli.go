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
    Update   bool
    Cache    bool
    Auth     bool
}

func init() {
    cli.BoolVar(&opts.Update, "update", false, "check for updates")
    cli.BoolVar(&opts.Cache, "cache", false, "cache spaces")
    cli.BoolVar(&opts.Auth, "auth", false, "authenticate")
}
