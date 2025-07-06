package ticketmanager

import (
	"github.com/ajaxray/geek-life/jira"
	"github.com/ajaxray/geek-life/util"
)

type JiraTicketManager struct {
	client jira.Jira
	config util.JiraConfig
}

func NewJiraTicketManager(config util.JiraConfig) *JiraTicketManager {
	client := jira.NewJiraClient(
		config.URL,
		config.Username,
		config.APIToken,
		config.APIToken,
		config.ProjectKey,
	)

	return &JiraTicketManager{
		client: client,
		config: config,
	}
}

func (j *JiraTicketManager) CreateEpic(title, description string) (string, error) {
	return j.client.CreateEpic(title, description)
}

func (j *JiraTicketManager) UpdateEpic(title, description string, epicID string) (string, error) {
	return j.client.UpdateEpic(title, description, epicID)
}

func (j *JiraTicketManager) ListEpics() ([]Epic, error) {
	jiraEpics, err := j.client.ListEpics()
	if err != nil {
		return nil, err
	}

	epics := make([]Epic, len(jiraEpics))
	for i, je := range jiraEpics {
		epics[i] = Epic{
			ID:          je.ID,
			Key:         je.Key,
			Title:       je.Fields.Summary,
			Description: getDescriptionString(je.Fields.Description),
			Status:      je.Fields.Status.Name,
			Creator: User{
				ID:          je.Fields.Creator.AccountID,
				Email:       je.Fields.Creator.EmailAddress,
				DisplayName: je.Fields.Creator.DisplayName,
			},
		}
	}

	return epics, nil
}

func (j *JiraTicketManager) ListUserEpics() ([]Epic, error) {
	jiraEpics, err := j.client.ListGeekLifeEpics()
	if err != nil {
		return nil, err
	}

	epics := make([]Epic, len(jiraEpics))
	for i, je := range jiraEpics {
		epics[i] = Epic{
			ID:          je.ID,
			Key:         je.Key,
			Title:       je.Fields.Summary,
			Description: getDescriptionString(je.Fields.Description),
			Status:      je.Fields.Status.Name,
			Creator: User{
				ID:          je.Fields.Creator.AccountID,
				Email:       je.Fields.Creator.EmailAddress,
				DisplayName: je.Fields.Creator.DisplayName,
			},
		}
	}

	return epics, nil
}

func (j *JiraTicketManager) DescribeEpic(epicID string) (*Epic, error) {
	jiraEpic, err := j.client.DescribeEpic(epicID)
	if err != nil {
		return nil, err
	}

	return &Epic{
		ID:          jiraEpic.ID,
		Key:         jiraEpic.Key,
		Title:       jiraEpic.Fields.Summary,
		Description: getDescriptionString(jiraEpic.Fields.Description),
		Status:      jiraEpic.Fields.Status.Name,
		Creator: User{
			ID:          jiraEpic.Fields.Creator.AccountID,
			Email:       jiraEpic.Fields.Creator.EmailAddress,
			DisplayName: jiraEpic.Fields.Creator.DisplayName,
		},
	}, nil
}

func (j *JiraTicketManager) CreateTask(title, description string, epicID string) (string, error) {
	return j.client.CreateTask(title, description, epicID)
}

func (j *JiraTicketManager) UpdateTask(
	title, description string,
	completed bool,
	taskID string,
) error {
	return j.client.UpdateTask(title, description, completed, taskID)
}

func (j *JiraTicketManager) ListTasksForEpic(epicID string) ([]Task, error) {
	jiraTasks, err := j.client.ListTasksForEpic(epicID)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, len(jiraTasks))
	for i, jt := range jiraTasks {
		tasks[i] = Task{
			ID:          jt.ID,
			Key:         jt.Key,
			Title:       jt.Fields.Summary,
			Description: getDescriptionString(jt.Fields.Description),
			Status:      jt.Fields.Status.Name,
			Completed:   jt.Fields.Status.StatusCategory.Key == "done",
			EpicID:      epicID,
			Creator: User{
				ID:          jt.Fields.Creator.AccountID,
				Email:       jt.Fields.Creator.EmailAddress,
				DisplayName: jt.Fields.Creator.DisplayName,
			},
		}
	}

	return tasks, nil
}

func (j *JiraTicketManager) DescribeTask(taskID string) (*Task, error) {
	jiraTask, err := j.client.DescribeTask(taskID)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:          jiraTask.ID,
		Key:         jiraTask.Key,
		Title:       jiraTask.Fields.Summary,
		Description: getDescriptionString(jiraTask.Fields.Description),
		Status:      jiraTask.Fields.Status.Name,
		Completed:   jiraTask.Fields.Status.StatusCategory.Key == "done",
		Creator: User{
			ID:          jiraTask.Fields.Creator.AccountID,
			Email:       jiraTask.Fields.Creator.EmailAddress,
			DisplayName: jiraTask.Fields.Creator.DisplayName,
		},
	}, nil
}

func getDescriptionString(desc interface{}) string {
	if desc == nil {
		return ""
	}
	if str, ok := desc.(string); ok {
		return str
	}
	return ""
}
