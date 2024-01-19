package model

// Project represent a collection of related tasks (tags of Habitica)
type Project struct {
	ID    int64  `storm:"id,increment" json:"id"`
	Title string `storm:"index" json:"title"`
	UUID  string `storm:"unique" json:"uuid,omitempty"`
	Jira  *Jira  `storm:"index" json:"jira"`
}

type Jira struct {
	EpicID *int64 `storm:"unique" json:"epic_id,omitempty"`
	TaskID *int64 `storm:"unique" json:"task_id,omitempty"`
}

func (p Project) GetTitle() string {
	if p.Jira != nil {
		return p.Title + " [Jira]"
	}

	return p.Title + " [No Jira project]"
}
