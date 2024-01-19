package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ajaxray/geek-life/api"
)

type Jira interface {
	CreateEpic(title, description string) (string, error)
	CreateTask(title string, epicID int64) (int64, error)
	ListEpics() ([]JiraIssue, error)
	DescribeEpic(epicID string) (*JiraIssue, error)
	DescribeTask(taskID string) (*JiraIssue, error)
}

type Epic struct {
	EpicID string `storm:unique json:"id"`
	Name   string `json:"name"`
	Desc   string `json:"desc"`
}

type Task struct {
	TaskID string `json:"id"`
	Name   string
	Desc   string
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
			// Add more fields here as needed
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("/rest/api/2/issue")
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		return "", err
	}
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		return "", err
	}
	return epic.ID, nil
}

func (j *jira) CreateTask(title string, epicID int64) (int64, error) {
	panic("not implemented") // TODO: Implement
}

func (j *jira) ListEpics() ([]JiraIssue, error) {
	url := fmt.Sprintf("/rest/api/2/search?jql=project=%s+AND+issuetype=Epic", j.projectKey)
	b, err := j.client.MakeRequest("GET", url)
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
	b, err := j.client.MakeRequest("GET", url)
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
	b, err := j.client.MakeRequest("GET", url)
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

func main() {
	j := jira{
		username:   "anujvarma@thumbtack.com",
		password:   os.Getenv("JIRA_API_TOKEN"),
		projectKey: "SRE",
	}
	url := "https://thumbtack.atlassian.net"
	j.client = *api.NewClient(url, j.username, j.password)

	t, err := j.DescribeTask("SRE-2")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t)
}
