package storm

import (
	"sort"
	"strings"
	"time"

	"github.com/asdine/storm/v3"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
)

type projectRepository struct {
	DB *storm.DB
}

// NewProjectRepository will create an object that represent the repository.Project interface
func NewProjectRepository(db *storm.DB) repository.ProjectRepository {
	return &projectRepository{db}
}

func (repo *projectRepository) GetAll() ([]model.Project, error) {
	var projects []model.Project
	err := repo.DB.All(&projects)

	return projects, err
}

func (repo *projectRepository) GetAllSortedByJiraDate() ([]model.Project, error) {
	var projects []model.Project
	err := repo.DB.All(&projects)
	if err != nil {
		return projects, err
	}

	// Sort projects: JIRA projects by creation date (newest first), then non-JIRA projects
	sort.Slice(projects, func(i, j int) bool {
		projectI := projects[i]
		projectJ := projects[j]

		// Both have JIRA creation dates - sort by date (newest first)
		if projectI.JiraCreatedDate != nil && projectJ.JiraCreatedDate != nil {
			return projectI.JiraCreatedDate.After(*projectJ.JiraCreatedDate)
		}

		// Only projectI has JIRA date - it comes first
		if projectI.JiraCreatedDate != nil && projectJ.JiraCreatedDate == nil {
			return true
		}

		// Only projectJ has JIRA date - it comes first
		if projectI.JiraCreatedDate == nil && projectJ.JiraCreatedDate != nil {
			return false
		}

		// Neither has JIRA date - sort by title alphabetically
		return projectI.Title < projectJ.Title
	})

	return projects, nil
}

func (repo *projectRepository) GetByID(id int64) (model.Project, error) {
	return repo.getOneByField("ID", id)
}

func (repo *projectRepository) GetByTitle(title string) (model.Project, error) {
	return repo.getOneByField("Title", title)
}

func (repo *projectRepository) GetByUUID(UUID string) (model.Project, error) {
	return repo.getOneByField("CloudId", UUID)
}

func (repo *projectRepository) Create(title, UUID string) (model.Project, error) {
	project := model.Project{
		Title: title,
		UUID:  UUID,
	}

	err := repo.DB.Save(&project)
	return project, err
}

func (repo *projectRepository) CreateWithJira(title, jiraID string) (model.Project, error) {
	project := model.Project{
		Title: title,
		Jira:  jiraID,
	}

	err := repo.DB.Save(&project)
	return project, err
}

func (repo *projectRepository) CreateWithJiraAndDate(title, jiraID string, createdDate *time.Time) (model.Project, error) {
	project := model.Project{
		Title:           title,
		Jira:            jiraID,
		JiraCreatedDate: createdDate,
	}

	err := repo.DB.Save(&project)
	return project, err
}

func (repo *projectRepository) Update(project *model.Project) error {
	return repo.DB.Save(project)
}

func (repo *projectRepository) Delete(project *model.Project) error {
	return repo.DB.DeleteStruct(project)
}

func (repo *projectRepository) UpdateField(
	task *model.Project,
	field string,
	value interface{},
) error {
	return repo.DB.UpdateField(task, field, value)
}

func (repo *projectRepository) SearchProjects(query string) ([]model.Project, error) {
	var allProjects []model.Project
	err := repo.DB.All(&allProjects)
	if err != nil {
		return nil, err
	}

	var matchingProjects []model.Project
	lowerQuery := strings.ToLower(query)

	for _, project := range allProjects {
		// Search in title and JIRA ID
		if strings.Contains(strings.ToLower(project.Title), lowerQuery) ||
			strings.Contains(strings.ToLower(project.Jira), lowerQuery) {
			matchingProjects = append(matchingProjects, project)
		}
	}

	return matchingProjects, nil
}

func (repo *projectRepository) getOneByField(
	fieldName string,
	val interface{},
) (model.Project, error) {
	var project model.Project
	err := repo.DB.One(fieldName, val, &project)

	return project, err
}
