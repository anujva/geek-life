package model

// Project represent a collection of related tasks (tags of Habitica)
type Project struct {
	ID    int64  `storm:"id,increment" json:"id"`
	Title string `storm:"index"        json:"title"`
	UUID  string `storm:"unique"       json:"uuid,omitempty"`
	Jira  string `storm:"index"        json:"jira"`
}

func (p Project) GetTitle() string {
	if p.Jira != "" {
		return p.Title + " [Ticket]"
	}

	return p.Title + " [No Ticket]"
}
