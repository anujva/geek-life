package jira

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ajaxray/geek-life/api"
	"github.com/ajaxray/geek-life/util"
)

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
		"/rest/api/3/field",
		nil,
	)
	if err != nil {
		util.LogError("error making request: %v", err)
		return err
	}

	util.LogDebug("Field configuration response: %s", string(b))
	v := make([]Field, 0)
	err = json.Unmarshal(b, &v)
	if err != nil {
		util.LogError("error unmarshalling field configuration: %+v", err)
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
	// Ensure config is loaded to get the correct epic name field
	err := j.ensureConfigLoaded()
	if err != nil {
		util.LogWarning("failed to load config: %v", err)
		// Continue with basic payload if config loading fails
	}

	// Convert description to Atlassian Document Format (ADF)
	adfDescription := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": description,
					},
				},
			},
		},
	}

	// Build the basic payload
	fields := map[string]interface{}{
		"project": map[string]string{
			"key": j.projectKey,
		},
		"summary":     title,
		"description": adfDescription,
		"issuetype": map[string]string{
			"name": "Epic",
		},
	}

	// Add epic name field if we have it configured
	if epicNameField, exists := j.config["epicName"]; exists && epicNameField != "" {
		fields[epicNameField] = title
		util.LogInfo("Using configured epic name field: %s", epicNameField)
	} else {
		// Try common epic name fields as fallback
		util.LogInfo("No epic name field configured, trying common fields")
		// Don't add any custom field by default to avoid 400 errors
	}

	payload := map[string]interface{}{
		"fields": fields,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	util.LogDebug("Epic creation payload: %s", string(payloadBytes))

	url := "/rest/api/3/issue"
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		util.LogError("Epic creation failed: %v", err)
		return "", err
	}
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		util.LogError("error unmarshalling epic response: %+v", err)
		return "", err
	}
	return epic.Key, nil
}

func (j *jira) UpdateEpic(title, description string, epicID string) (string, error) {
	// Convert description to Atlassian Document Format (ADF)
	adfDescription := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": description,
					},
				},
			},
		},
	}

	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": adfDescription,
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
	url := fmt.Sprintf("/rest/api/3/issue/%s", epicID)
	b, err := j.client.MakeRequest("PUT", url, payloadBytes)
	if err != nil {
		return "", err
	}
	epic := &JiraIssue{}
	err = json.Unmarshal(b, epic)
	if err != nil {
		util.LogError("error unmarshalling epic response: %+v", err)
		return "", err
	}
	return epic.Key, nil
}

func (j *jira) CreateTask(title, description string, epicID string) (string, error) {
	// Convert description to Atlassian Document Format (ADF)
	adfDescription := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": description,
					},
				},
			},
		},
	}

	// Try simple task creation first
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": adfDescription,
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

	url := "/rest/api/3/issue"
	b, err := j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		return "", err
	}
	util.LogDebug("Task creation response: %s", string(b))
	task := &JiraIssue{}
	err = json.Unmarshal(b, task)
	if err != nil {
		util.LogError("error unmarshalling task response: %v", err)
		return "", err
	}
	return task.Key, nil
}

func (j *jira) getTransitionID(taskID string, completed bool) (string, error) {
	// Get available transitions for this task
	url := fmt.Sprintf("/rest/api/3/issue/%s/transitions", taskID)
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

	util.LogDebug("Available transitions for task %s:", taskID)
	for _, transition := range transitions.Transitions {
		util.LogDebug("  ID: %s, Name: %s, To: %s, Category: %s",
			transition.ID, transition.Name, transition.To.Name, transition.To.StatusCategory.Key)
	}

	// Look for appropriate transition based on completion status
	targetCategory := "new"
	if completed {
		targetCategory = "done"
	}

	for _, transition := range transitions.Transitions {
		if transition.To.StatusCategory.Key == targetCategory {
			util.LogInfo("Selected transition ID %s for completed=%v", transition.ID, completed)
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
				util.LogInfo(
					"Selected transition ID %s by name match for completed=%v",
					transition.ID,
					completed,
				)
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
	// Convert description to Atlassian Document Format (ADF)
	adfDescription := map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": description,
					},
				},
			},
		},
	}

	// Construct the request payload
	payload := map[string]interface{}{
		"fields": map[string]interface{}{
			"project": map[string]string{
				"key": j.projectKey,
			},
			"summary":     title,
			"description": adfDescription,
		},
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	util.LogDebug("Task update payload: %s", payloadBytes)
	url := fmt.Sprintf("/rest/api/3/issue/%s", taskID)
	b, err := j.client.MakeRequest("PUT", url, payloadBytes)
	if err != nil {
		return err
	}
	util.LogDebug("Task update response: %s", b)
	// Get the correct transition ID for this task
	transitionID, err := j.getTransitionID(taskID, completed)
	if err != nil {
		util.LogError("Error getting transition ID: %v", err)
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
		util.LogError("Error while marshaling payload: %+v", err)
		return err
	}
	util.LogDebug("Task update payload: %s", payloadBytes)
	url = fmt.Sprintf("/rest/api/3/issue/%s/transitions", taskID)
	_, err = j.client.MakeRequest("POST", url, payloadBytes)
	if err != nil {
		util.LogError("Error while calling transitions: %+v", err)
		return err
	}
	return nil
}

func (j *jira) ListEpics() ([]JiraIssue, error) {
	jql := fmt.Sprintf("project=%s AND issuetype=Epic", j.projectKey)

	// Use the new JQL search API endpoint with POST method
	requestURL := "/rest/api/3/search/jql"

	// Prepare the request payload with JQL and required fields
	payload := map[string]interface{}{
		"jql": jql,
		"fields": []string{
			"key",
			"summary",
			"description",
			"creator",
			"assignee",
			"status",
			"created",
			"updated",
		},
		"maxResults": 100, // Reasonable limit to prevent large responses
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JQL request payload: %w", err)
	}

	util.LogInfo("Making JIRA API call (v3 JQL): %s with JQL: %s", requestURL, jql)
	b, err := j.client.MakeRequest("POST", requestURL, payloadBytes)
	if err != nil {
		return nil, fmt.Errorf("JIRA API v3 JQL search failed: %w", err)
	}

	epics := JiraIssueResult{}
	err = json.Unmarshal(b, &epics)
	if err != nil {
		return nil, err
	}
	return epics.Issues, nil
}

func (j *jira) ListGeekLifeEpics() ([]JiraIssue, error) {
	util.LogInfo("Starting epic import - calling ListEpics()")
	// Get all epics and then filter by user involvement (created OR has tasks assigned)
	allEpics, err := j.ListEpics()
	if err != nil {
		util.LogError("Failed to list epics: %v", err)
		return nil, fmt.Errorf("failed to list epics - this might be due to API changes. Try updating your JIRA instance or check API permissions: %w", err)
	}

	util.LogInfo("Got %d total epics, filtering for user involvement: %s", len(allEpics), j.username)

	// If no epics found, return empty list instead of failing
	if len(allEpics) == 0 {
		util.LogInfo("No epics found in project %s", j.projectKey)
		return []JiraIssue{}, nil
	}

	var userEpics []JiraIssue

	for _, epic := range allEpics {
		util.LogDebug("Checking epic: %s (%s)", epic.Fields.Summary, epic.Key)

		// Check if user created this epic - use multiple matching strategies
		userCreated := j.isUserMatch(epic.Fields.Creator.EmailAddress, epic.Fields.Creator.DisplayName, epic.Fields.Creator.AccountID)

		if userCreated {
			util.LogInfo("Epic %s created by user", epic.Key)
			userEpics = append(userEpics, epic)
			continue
		}

		// Check if user has tasks in this epic (with error resilience)
		userHasTasks, err := j.userHasTasksInEpic(epic.Key)
		if err != nil {
			util.LogWarning("Error checking tasks for epic %s (continuing): %v", epic.Key, err)
			continue
		}

		if userHasTasks {
			util.LogInfo("Epic %s has tasks assigned to user", epic.Key)
			userEpics = append(userEpics, epic)
		}
	}

	util.LogInfo("Found %d epics with user involvement", len(userEpics))

	return userEpics, nil
}

// isUserMatch checks if the given user details match the current user using multiple strategies
func (j *jira) isUserMatch(emailAddress, displayName, accountID string) bool {
	// Primary match: email address (most reliable)
	if emailAddress != "" && strings.EqualFold(emailAddress, j.username) {
		return true
	}

	// Secondary match: display name
	if displayName != "" && strings.EqualFold(displayName, j.username) {
		return true
	}

	// Tertiary match: account ID (if username happens to be an account ID)
	if accountID != "" && accountID == j.username {
		return true
	}

	return false
}

// userHasTasksInEpic checks if the current user has any tasks assigned in the given epic
func (j *jira) userHasTasksInEpic(epicID string) (bool, error) {
	// Get tasks for this epic and check if any are assigned to the current user
	tasks, err := j.ListTasksForEpic(epicID)
	if err != nil {
		return false, err
	}

	for _, task := range tasks {
		if j.isUserMatch(task.Fields.Assignee.EmailAddress, task.Fields.Assignee.DisplayName, task.Fields.Assignee.AccountID) {
			util.LogDebug("Found task %s assigned to user in epic %s", task.Key, epicID)
			return true, nil
		}
	}

	return false, nil
}

func (j *jira) filterEpicsByUser() ([]JiraIssue, error) {
	// Get all epics first
	allEpics, err := j.ListEpics()
	if err != nil {
		return nil, err
	}

	util.LogInfo("Got %d total epics, filtering for user: %s", len(allEpics), j.username)

	var userEpics []JiraIssue
	for _, epic := range allEpics {
		util.LogDebug("Epic: %s, Creator email: %s, Creator name: %s, Creator ID: %s",
			epic.Fields.Summary,
			epic.Fields.Creator.EmailAddress,
			epic.Fields.Creator.DisplayName,
			epic.Fields.Creator.AccountID)

		// Use the improved user matching function
		if j.isUserMatch(epic.Fields.Creator.EmailAddress, epic.Fields.Creator.DisplayName, epic.Fields.Creator.AccountID) {
			userEpics = append(userEpics, epic)
			util.LogInfo("Matched epic: %s", epic.Fields.Summary)
		}
	}

	util.LogInfo("Found %d user epics", len(userEpics))
	return userEpics, nil
}

func (j *jira) ListTasksForEpic(epicID string) ([]JiraIssue, error) {
	// Try multiple JQL queries for epic-task relationships
	queries := []string{
		fmt.Sprintf("project=%s AND parent=%s", j.projectKey, epicID),
		fmt.Sprintf("project=%s AND \"Epic Link\"=%s", j.projectKey, epicID),
		fmt.Sprintf("project=%s AND cf[10014]=%s", j.projectKey, epicID), // Common epic link field
		fmt.Sprintf(
			"project=%s AND cf[10011]=%s",
			j.projectKey,
			epicID,
		), // Another common epic link field
	}

	requestURL := "/rest/api/3/search/jql"

	for _, jql := range queries {
		// Prepare the request payload with JQL and required fields
		payload := map[string]interface{}{
			"jql": jql,
			"fields": []string{
				"key",
				"summary",
				"description",
				"assignee",
				"status",
				"created",
				"updated",
			},
			"maxResults": 100,
		}

		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			util.LogWarning("Failed to marshal JQL payload for query: %s", jql)
			continue
		}

		b, err := j.client.MakeRequest("POST", requestURL, payloadBytes)
		if err != nil {
			util.LogDebug("JQL query failed: %s, error: %v", jql, err)
			continue
		}

		tasks := JiraIssueResult{}
		err = json.Unmarshal(b, &tasks)
		if err != nil {
			util.LogDebug("Failed to unmarshal response for query: %s", jql)
			continue
		}

		if len(tasks.Issues) > 0 {
			util.LogInfo("Found %d tasks for epic %s using query: %s", len(tasks.Issues), epicID, jql)
			return tasks.Issues, nil
		}
	}

	util.LogInfo("No tasks found for epic %s", epicID)
	return []JiraIssue{}, nil
}

func (j *jira) DescribeEpic(epicID string) (*JiraIssue, error) {
	url := fmt.Sprintf("/rest/api/3/issue/%s", epicID)
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
	url := fmt.Sprintf("/rest/api/3/issue/%s", taskID)
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
