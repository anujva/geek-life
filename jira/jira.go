package jira

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajaxray/geek-life/api"
)

var file *os.File

func init() {
	var err error
	file, err = os.Create("output.txt")
	if err != nil {
		panic(err)
	}
}

type Jira interface {
	CreateEpic(title, description string) (string, error)
	UpdateEpic(title, description string, epicID string) (string, error)
	CreateTask(title, description string, epicID string) (string, error)
	UpdateTask(title, description string, completed bool, taskID string) error
	ListEpics() ([]JiraIssue, error)
	DescribeEpic(epicID string) (*JiraIssue, error)
	DescribeTask(taskID string) (*JiraIssue, error)
}

func NewJiraClient(url, username, password, token, projectKey string) Jira {
	j := jira{
		username:   username,
		password:   password,
		projectKey: projectKey,
	}
	j.client = *api.NewClient(url, username, password, token)
	j.config = make(map[string]string)
	j.UpdateConfig()
	return &j
}

type jira struct {
	username   string
	password   string
	client     api.Client
	projectKey string
	config     map[string]string
}

func (j *jira) UpdateConfig() {
	b, err := j.client.MakeRequest(
		"GET",
		"/rest/api/2/field",
		nil,
	)
	if err != nil {
		fmt.Println("error making request", err)
		os.Exit(12)
	}

	fmt.Fprintf(file, "%s", string(b))
	v := make([]Field, 0)
	json.Unmarshal(b, &v)

	for _, field := range v {
		if field.Name == "Epic Name" {
			j.config["epicName"] = field.ID
		}
		if field.Name == "Parent" {
			j.config["epicLink"] = field.ID
		}
	}
}

func (j *jira) CreateEpic(title, description string) (string, error) {
	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": description,
			"issuetype": map[string]string{
				"name": "Epic",
			},
			j.config["epicName"]: title,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := "/rest/api/2/issue"
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		return "", err
	}
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		fmt.Fprintf(file, "error unmarshalling; %+v\n", err)
		return "", err
	}
	return epic.ID, nil
}

func (j *jira) UpdateEpic(title, description string, epicID string) (string, error) {
	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": description,
			"issuetype": map[string]string{
				"name": "Epic",
			},
			j.config["epicName"]: title,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("/rest/api/2/issue/%s", epicID)
	b, err := j.client.MakeRequest("PUT", url, payloadBytes)
	if err != nil {
		return "", err
	}
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		fmt.Fprintf(file, "error unmarshalling: %+v\n", err)
		return "", err
	}
	return epic.ID, nil
}

func (j *jira) CreateTask(title, description string, epicID string) (string, error) {
	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": description,
			"issuetype": map[string]string{
				"name": "Task",
			},
			j.config["epicLink"]: map[string]string{
				"key": epicID,
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := "/rest/api/2/issue"
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		return "", err
	}
	fmt.Fprintf(file, "Task: %s\n", string(b))
	task := &JiraIssue{}
	err = json.Unmarshal(b, task)
	if err != nil {
		fmt.Println("error unmarshalling", err)
		return "", err
	}
	return task.Key, nil
}

func getIDFromCompleted(completed bool) string {
	if completed {
		return "41"
	}
	return "11"
}

func (j *jira) UpdateTask(
	title,
	description string,
	completed bool,
	taskID string,
) error {
	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": description,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "Bytes: %s\n", payloadBytes)
	url := fmt.Sprintf("/rest/api/2/issue/%s", taskID)
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "B: %s\n", b)
	// Construct the request payload
	payload = map[string]interface{}{
		"transition": map[string]interface{}{
			"id": getIDFromCompleted(completed),
		},
	}
	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(file, "Error while marshing payload: %+v\n", err)
		return err
	}
	fmt.Fprintf(file, "Bytes: %s\n", payloadBytes)
	url = fmt.Sprintf("/rest/api/2/issue/%s/transitions", taskID)
	_, err = j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		fmt.Fprintf(file, "Error while calling transitions: %+v\n", err)
		return err
	}
	return nil
}

func (j *jira) ListEpics() ([]JiraIssue, error) {
	url := fmt.Sprintf("/rest/api/2/search?jql=project=%s+AND+issuetype=Epic", j.projectKey)
	b, err := j.client.MakeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	epics := JiraIssueResult{}
	err = json.Unmarshal(b, &epics)
	if err != nil {
		return nil, err
	}
	return epics.Issues, nil
}

func (j *jira) DescribeEpic(epicID string) (*JiraIssue, error) {
	url := fmt.Sprintf("/rest/api/2/issue/%s", epicID)
	b, err := j.client.MakeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	jiraissue := &JiraIssue{}
	err = json.Unmarshal(b, jiraissue)
	if err != nil {
		return nil, err
	}
	return jiraissue, nil
}

func (j *jira) DescribeTask(taskID string) (*JiraIssue, error) {
	url := fmt.Sprintf("/rest/api/2/issue/%s", taskID)
	b, err := j.client.MakeRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	jiraissue := &JiraIssue{}
	err = json.Unmarshal(b, jiraissue)
	if err != nil {
		return nil, err
	}
	return jiraissue, nil
}

/*
func main() {
	j := jira{
		username:   "anujva@gmail.com",
		password:   os.Getenv("JIRA_API_TOKEN"),
		projectKey: "SRE",
	}
	url := "http://localhost:8080"
	j.client = *api.NewClient(url, j.username, j.password, j.password)

	t, err := j.CreateTask("This is an EPIC", "Do something something", "SRE-1")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t)
}
*/
