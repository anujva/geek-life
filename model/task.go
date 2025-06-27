package model

// Task represent a task - the building block of the TaskManager app
type Task struct {
	ID        int64  `storm:"id,increment" json:"ID"`
	ProjectID int64  `storm:"index"        json:"ProjectID"`
	UUID      string `storm:"unique"       json:"UUID,omitempty"`
	Title     string `                     json:"text"`
	Details   string `                     json:"notes"`
	Completed bool   `storm:"index"        json:"Completed"`
	DueDate   int64  `storm:"index"        json:"DueDate,omitempty"`
	JiraID    string `storm:"unique"       json:"jira,omitempty"`
}
