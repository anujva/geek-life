package model

import "github.com/ajaxray/geek-life/jira"

// Project represent a collection of related tasks (tags of Habitica)
type Project struct {
	ID    int64           `storm:"id,increment" json:"id"`
	Title string          `storm:"index" json:"title"`
	UUID  string          `storm:"unique" json:"uuid,omitempty"`
	Jira  *jira.JiraIssue `storm:"index" json:"jira"`
}

func (p Project) GetTitle() string {
	if p.Jira != nil {
		return p.Title + " [Jira]"
	}

	return p.Title + " [No Jira project]"
}
