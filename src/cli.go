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
	GetProjects   bool
	GetIssuetypes bool
	GetStatus     bool
	GetAssignees  bool
	MyIssues      bool
	Auth          bool
	Create        bool
	JQL           bool
	Project       string
	Projects      string
	Issuetype     string
}

func init() {
	cli.BoolVar(&opts.Update, "update", false, "check for updates")
	cli.BoolVar(&opts.GetProjects, "get-projects", false, "search projects")
	cli.BoolVar(&opts.GetIssuetypes, "get-issuetypes", false, "search issuetypes")
	cli.BoolVar(&opts.GetStatus, "get-status", false, "search status")
	cli.BoolVar(&opts.GetAssignees, "get-assignees", false, "search assignees")
	cli.BoolVar(&opts.MyIssues, "myissues", false, "search my assigned issues")
	cli.BoolVar(&opts.Auth, "auth", false, "authenticate")
	cli.BoolVar(&opts.Create, "create", false, "create jira issue")
	cli.BoolVar(&opts.JQL, "jql", false, "jql search")
	cli.StringVar(&opts.Project, "project", "", "project")
	cli.StringVar(&opts.Projects, "projects", "", "projects")
	cli.StringVar(&opts.Issuetype, "issuetype", "", "issuetype")
}
