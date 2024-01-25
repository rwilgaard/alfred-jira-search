package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQuery(t *testing.T) {
    expected := &parsedQuery{
        Text: "test query ",
        Projects: []string{"testproject"},
        Issuetypes: []string{"issuetype", "issue type"},
        Status: []string{"open", "in progress"},
        Assignees: []string{"testuser"},
    }
    q := "test query @testproject #issuetype #issue_type ?open ?in_progress %testuser"
    pq := parseQuery(q)
    assert.Equal(t, expected, pq)
}
