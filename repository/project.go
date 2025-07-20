package repository

import (
	"time"
	"github.com/ajaxray/geek-life/model"
)

// ProjectRepository interface defines methods of project data accessor
type ProjectRepository interface {
	GetAll() ([]model.Project, error)
	GetAllSortedByJiraDate() ([]model.Project, error)
	GetByID(id int64) (model.Project, error)
	GetByTitle(title string) (model.Project, error)
	GetByUUID(UUID string) (model.Project, error)
	Create(title, UUID string) (model.Project, error)
	CreateWithJira(title, jiraID string) (model.Project, error)
	CreateWithJiraAndDate(title, jiraID string, createdDate *time.Time) (model.Project, error)
	Update(p *model.Project) error
	UpdateField(p *model.Project, field string, value interface{}) error
	Delete(p *model.Project) error
	SearchProjects(query string) ([]model.Project, error)
}
