package jira

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

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
	ListGeekLifeEpics() ([]JiraIssue, error)
	ListTasksForEpic(epicID string) ([]JiraIssue, error)
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
	// Don't call UpdateConfig() here to avoid auth errors on initialization
	// It will be called lazily when needed
	return &j
}

type jira struct {
	username     string
	password     string
	client       api.Client
	projectKey   string
	config       map[string]string
	configLoaded bool
}

func (j *jira) ensureConfigLoaded() error {
	if j.configLoaded {
		return nil
	}
	return j.UpdateConfig()
}

func (j *jira) UpdateConfig() error {
	b, err := j.client.MakeRequest(
		"GET",
		"/rest/api/2/field",
		nil,
	)
	if err != nil {
		fmt.Fprintf(file, "error making request: %v\n", err)
		return err
	}

	fmt.Fprintf(file, "%s", string(b))
	v := make([]Field, 0)
	err = json.Unmarshal(b, &v)
	if err != nil {
		fmt.Fprintf(file, "error unmarshalling: %+v\n", err)
		return err
	}

	for _, field := range v {
		if field.Name == "Epic Name" {
			j.config["epicName"] = field.ID
		}
		if field.Name == "Parent" {
			j.config["epicLink"] = field.ID
		}
	}

	j.configLoaded = true
	return nil
}

func (j *jira) CreateEpic(title, description string) (string, error) {
	// Try simple epic creation first, fall back to complex config if needed
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
			// Try common epic name fields
			"customfield_10011": title, // Common epic name field
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
	// Try simple task creation first
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
			// Try common epic link fields
			"parent": map[string]string{
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

func (j *jira) getTransitionID(taskID string, completed bool) (string, error) {
	// Get available transitions for this task
	url := fmt.Sprintf("/rest/api/2/issue/%s/transitions", taskID)
	b, err := j.client.MakeRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	
	var transitions struct {
		Transitions []struct {
			ID   string `json:"id"`
			Name string `json:"name"`
			To   struct {
				Name           string `json:"name"`
				StatusCategory struct {
					Key string `json:"key"`
				} `json:"statusCategory"`
			} `json:"to"`
		} `json:"transitions"`
	}
	
	err = json.Unmarshal(b, &transitions)
	if err != nil {
		return "", err
	}
	
	fmt.Fprintf(file, "Available transitions for task %s:\n", taskID)
	for _, transition := range transitions.Transitions {
		fmt.Fprintf(file, "  ID: %s, Name: %s, To: %s, Category: %s\n", 
			transition.ID, transition.Name, transition.To.Name, transition.To.StatusCategory.Key)
	}
	
	// Look for appropriate transition based on completion status
	targetCategory := "new"
	if completed {
		targetCategory = "done"
	}
	
	for _, transition := range transitions.Transitions {
		if transition.To.StatusCategory.Key == targetCategory {
			fmt.Fprintf(file, "Selected transition ID %s for completed=%v\n", transition.ID, completed)
			return transition.ID, nil
		}
	}
	
	// Fallback to common transition names
	targetNames := []string{"To Do", "Open", "New"}
	if completed {
		targetNames = []string{"Done", "Closed", "Complete", "Resolved"}
	}
	
	for _, targetName := range targetNames {
		for _, transition := range transitions.Transitions {
			if strings.Contains(strings.ToLower(transition.To.Name), strings.ToLower(targetName)) {
				fmt.Fprintf(file, "Selected transition ID %s by name match for completed=%v\n", transition.ID, completed)
				return transition.ID, nil
			}
		}
	}
	
	return "", fmt.Errorf("no suitable transition found for completed=%v", completed)
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
	b, err := j.client.MakeRequest("PUT", url, payloadBytes)
	if err != nil {
		return err
	}
	fmt.Fprintf(file, "B: %s\n", b)
	// Get the correct transition ID for this task
	transitionID, err := j.getTransitionID(taskID, completed)
	if err != nil {
		fmt.Fprintf(file, "Error getting transition ID: %v\n", err)
		return err
	}

	// Construct the request payload
	payload = map[string]interface{}{
		"transition": map[string]interface{}{
			"id": transitionID,
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
	jql := fmt.Sprintf("project=%s AND issuetype=Epic", j.projectKey)
	encodedJQL := url.QueryEscape(jql)
	requestURL := fmt.Sprintf("/rest/api/2/search?jql=%s", encodedJQL)
	b, err := j.client.MakeRequest("GET", requestURL, nil)
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

func (j *jira) ListGeekLifeEpics() ([]JiraIssue, error) {
	// Use JQL to filter epics created by the current user
	// Try multiple formats for username matching
	queries := []string{
		fmt.Sprintf("project=%s AND issuetype=Epic AND creator=\"%s\"", j.projectKey, j.username),
		fmt.Sprintf("project=%s AND issuetype=Epic AND creator=%s", j.projectKey, j.username),
		fmt.Sprintf("project=%s AND issuetype=Epic AND creator=currentUser()", j.projectKey),
	}

	fmt.Fprintf(
		file,
		"Trying to find epics for user: %s in project: %s\n",
		j.username,
		j.projectKey,
	)

	for i, jql := range queries {
		// Properly URL encode the JQL query
		encodedJQL := url.QueryEscape(jql)
		requestURL := fmt.Sprintf("/rest/api/2/search?jql=%s", encodedJQL)
		fmt.Fprintf(file, "Attempt %d - JQL: %s\n", i+1, jql)
		fmt.Fprintf(file, "Attempt %d - Encoded URL: %s\n", i+1, requestURL)

		b, err := j.client.MakeRequest("GET", requestURL, nil)
		if err != nil {
			fmt.Fprintf(file, "Error with query %d: %v\n", i+1, err)
			continue
		}

		epics := JiraIssueResult{}
		err = json.Unmarshal(b, &epics)
		if err != nil {
			fmt.Fprintf(file, "Error unmarshaling response for query %d: %v\n", i+1, err)
			continue
		}

		fmt.Fprintf(file, "Query %d returned %d epics\n", i+1, len(epics.Issues))
		if len(epics.Issues) > 0 {
			// Log the first epic's creator info for debugging
			if len(epics.Issues) > 0 {
				fmt.Fprintf(file, "First epic creator: %s (email: %s)\n",
					epics.Issues[0].Fields.Creator.DisplayName,
					epics.Issues[0].Fields.Creator.EmailAddress)
			}
			return epics.Issues, nil
		}
	}

	// If no queries returned results, fall back to getting all epics and filter manually
	fmt.Fprintf(file, "No filtered results, falling back to manual filtering\n")
	return j.filterEpicsByUser()
}

func (j *jira) filterEpicsByUser() ([]JiraIssue, error) {
	// Get all epics first
	allEpics, err := j.ListEpics()
	if err != nil {
		return nil, err
	}

	fmt.Fprintf(file, "Got %d total epics, filtering for user: %s\n", len(allEpics), j.username)

	var userEpics []JiraIssue
	for _, epic := range allEpics {
		fmt.Fprintf(file, "Epic: %s, Creator email: %s, Creator name: %s\n",
			epic.Fields.Summary,
			epic.Fields.Creator.EmailAddress,
			epic.Fields.Creator.DisplayName)

		// Try multiple matching criteria
		if epic.Fields.Creator.EmailAddress == j.username ||
			epic.Fields.Creator.DisplayName == j.username ||
			strings.ToLower(epic.Fields.Creator.EmailAddress) == strings.ToLower(j.username) {
			userEpics = append(userEpics, epic)
			fmt.Fprintf(file, "Matched epic: %s\n", epic.Fields.Summary)
		}
	}

	fmt.Fprintf(file, "Found %d user epics\n", len(userEpics))
	return userEpics, nil
}

func (j *jira) ListTasksForEpic(epicID string) ([]JiraIssue, error) {
	// Try multiple JQL queries for epic-task relationships
	queries := []string{
		fmt.Sprintf("project=%s AND parent=%s", j.projectKey, epicID),
		fmt.Sprintf("project=%s AND \"Epic Link\"=%s", j.projectKey, epicID),
		fmt.Sprintf("project=%s AND cf[10014]=%s", j.projectKey, epicID), // Common epic link field
		fmt.Sprintf("project=%s AND cf[10011]=%s", j.projectKey, epicID), // Another common epic link field
	}
	
	for _, jql := range queries {
		encodedJQL := url.QueryEscape(jql)
		requestURL := fmt.Sprintf("/rest/api/2/search?jql=%s", encodedJQL)
		
		b, err := j.client.MakeRequest("GET", requestURL, nil)
		if err != nil {
			continue
		}
		
		tasks := JiraIssueResult{}
		err = json.Unmarshal(b, &tasks)
		if err != nil {
			continue
		}
		
		if len(tasks.Issues) > 0 {
			return tasks.Issues, nil
		}
	}
	
	return []JiraIssue{}, nil
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
