package storm

import (
	"strings"
	"time"

	"github.com/asdine/storm/v3"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
)

type taskRepository struct {
	DB *storm.DB
}

// NewTaskRepository will create an object that represent the repository.Task interface
func NewTaskRepository(db *storm.DB) repository.TaskRepository {
	return &taskRepository{db}
}

func (t *taskRepository) GetAll() ([]model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) GetAllByProject(project model.Project) ([]model.Task, error) {
	var tasks []model.Task
	//err = db.Find("ProjetID", project.ID, &tasks, storm.Limit(10), storm.Skip(10), storm.Reverse())
	err := t.DB.Find("ProjectID", project.ID, &tasks)

	return tasks, err
}

func (t *taskRepository) GetAllByDate(date time.Time) ([]model.Task, error) {
	var tasks []model.Task

	if date.IsZero() {
		var allTasks []model.Task
		err := t.DB.AllByIndex("ProjectID", &allTasks)
		for _, t := range allTasks {
			if t.DueDate == 0 {
				tasks = append(tasks, t)
			}
		}

		return tasks, err
	} else {
		err := t.DB.Find("DueDate", getRoundedDueDate(date), &tasks)
		return tasks, err
	}
}

func (t *taskRepository) GetAllByDateRange(from, to time.Time) ([]model.Task, error) {
	var tasks []model.Task

	err := t.DB.Range("DueDate", getRoundedDueDate(from), getRoundedDueDate(to), &tasks)
	return tasks, err
}

func (t *taskRepository) GetByID(ID string) (model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) GetByUUID(UUID string) (model.Task, error) {
	panic("implement me")
}

func (t *taskRepository) GetByJiraID(jiraID string) (*model.Task, error) {
	var task model.Task
	err := t.DB.One("JiraID", jiraID, &task)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (t *taskRepository) Create(
	project model.Project,
	title, details, UUID string,
	dueDate int64,
) (model.Task, error) {
	task := model.Task{
		ProjectID: project.ID,
		Title:     title,
		Details:   details,
		UUID:      UUID,
		DueDate:   dueDate,
	}

	err := t.DB.Save(&task)
	return task, err
}

func (t *taskRepository) CreateTask(task *model.Task) error {
	return t.DB.Save(task)
}

func (t *taskRepository) Update(task *model.Task) error {
	return t.DB.Update(task)
}

func (t *taskRepository) UpdateField(task *model.Task, field string, value interface{}) error {
	return t.DB.UpdateField(task, field, value)
}

func (t *taskRepository) Delete(task *model.Task) error {
	return t.DB.DeleteStruct(task)
}

func (t *taskRepository) DeleteAllByProjectID(projectID int64) error {
	var tasks []model.Task
	err := t.DB.Find("ProjectID", projectID, &tasks)
	if err != nil {
		// If no tasks are found, that's fine - nothing to delete
		if err == storm.ErrNotFound {
			return nil
		}
		return err
	}

	for i := range tasks {
		if err := t.DB.DeleteStruct(&tasks[i]); err != nil {
			return err
		}
	}

	return nil
}

func (t *taskRepository) SearchTasks(query string) ([]model.Task, error) {
	var allTasks []model.Task
	err := t.DB.All(&allTasks)
	if err != nil {
		return nil, err
	}

	var matchingTasks []model.Task
	lowerQuery := strings.ToLower(query)
	
	for _, task := range allTasks {
		// Search in title, details, and JIRA ID
		if strings.Contains(strings.ToLower(task.Title), lowerQuery) ||
		   strings.Contains(strings.ToLower(task.Details), lowerQuery) ||
		   strings.Contains(strings.ToLower(task.JiraID), lowerQuery) {
			matchingTasks = append(matchingTasks, task)
		}
	}

	return matchingTasks, nil
}

func (t *taskRepository) SearchTasksInProject(projectID int64, query string) ([]model.Task, error) {
	var projectTasks []model.Task
	err := t.DB.Find("ProjectID", projectID, &projectTasks)
	if err != nil {
		if err == storm.ErrNotFound {
			return []model.Task{}, nil
		}
		return nil, err
	}

	var matchingTasks []model.Task
	lowerQuery := strings.ToLower(query)
	
	for _, task := range projectTasks {
		// Search in title, details, and JIRA ID
		if strings.Contains(strings.ToLower(task.Title), lowerQuery) ||
		   strings.Contains(strings.ToLower(task.Details), lowerQuery) ||
		   strings.Contains(strings.ToLower(task.JiraID), lowerQuery) {
			matchingTasks = append(matchingTasks, task)
		}
	}

	return matchingTasks, nil
}

func getRoundedDueDate(date time.Time) int64 {
	if date.IsZero() {
		return 0
	}

	return date.Unix()
}
