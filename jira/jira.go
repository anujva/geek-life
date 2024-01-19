package jira

import (
	"encoding/json"
	"fmt"

	"github.com/ajaxray/geek-life/api"
)

type Jira interface {
	CreateEpic(title, description string) (string, error)
	CreateTask(title, description string, epicID string) (string, error)
	ListEpics() ([]JiraIssue, error)
	DescribeEpic(epicID string) (*JiraIssue, error)
	DescribeTask(taskID string) (*JiraIssue, error)
}

type jira struct {
	username   string
	password   string
	client     api.Client
	projectKey string
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
			"customfield_10104": title,
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
	fmt.Println(string(b))
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		fmt.Println("error unmarshalling", err)
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
			"customfield_10108": epicID,
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
	fmt.Println(string(b))
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		fmt.Println("error unmarshalling", err)
		return "", err
	}
	return epic.ID, nil
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
