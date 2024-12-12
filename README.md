# Jira Search

A workflow to search Jira issues.

## Installation

- [Download the latest release](https://github.com/rwilgaard/alfred-jira-search/releases)
- Open the downloaded file in Finder.
- If running on macOS Catalina or later, you _**MUST**_ add Alfred to the list of security exceptions for running unsigned software. See [this guide](https://github.com/deanishe/awgo/wiki/Catalina) for instructions on how to do this.

## Keywords

- `js` is used for searching issues.
- `ja` is used for searching issues assigned to yourself.
- `jq` is used for searching issues using JQL.
- `jc` is used for creating an issue with yourself as assignee.

## Actions

The following actions can be used on a highlighted repository:
- `⏎` will open the issue in Jira.
- `⌥` will show you additional info about the issue.
- `⌘` + `⏎` will start a time entry in Toggl if you have enabled the [Jira Toggl Integration](#jira-toggl-integration).
- `⌃` + `⏎` will open all search results in Jira.

## Jira Toggl Integration

In the **User Configurtion** you can enable an integration to the [Jira Toggl Integration](https://github.com/atoftegaard-git/alfred-jira-toggl) workflow if you have it installed.
This will allow you to start tracking time on an issue directly from this workflow.
