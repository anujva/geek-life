package repository

import (
	"time"

	"github.com/ajaxray/geek-life/model"
)

// TaskRepository interface defines methods of task data accessor
type TaskRepository interface {
	GetAll() ([]model.Task, error)
	GetAllByProject(project model.Project) ([]model.Task, error)
	GetAllByDate(date time.Time) ([]model.Task, error)
	GetAllByDateRange(from, to time.Time) ([]model.Task, error)
	GetByID(ID string) (model.Task, error)
	GetByUUID(UUID string) (model.Task, error)
	GetByJiraID(jiraID string) (*model.Task, error)
	Create(project model.Project, title, details, UUID string, dueDate int64) (model.Task, error)
	CreateTask(task *model.Task) error
	Update(t *model.Task) error
	UpdateField(t *model.Task, field string, value interface{}) error
	Delete(t *model.Task) error
	DeleteAllByProjectID(projectID int64) error
	SearchTasks(query string) ([]model.Task, error)
	SearchTasksInProject(projectID int64, query string) ([]model.Task, error)
}
