package ticketmanager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/ajaxray/geek-life/util"
)

type LinearTicketManager struct {
	apiKey  string
	teamKey string
	teamID  string
	client  *http.Client
	baseURL string
}

type LinearConfig struct {
	APIKey  string
	TeamKey string
}

func NewLinearTicketManager(config LinearConfig) *LinearTicketManager {
	return &LinearTicketManager{
		apiKey:  config.APIKey,
		teamKey: config.TeamKey,
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: "https://api.linear.app/graphql",
	}
}

func (l *LinearTicketManager) makeRequest(
	query string,
	variables map[string]interface{},
) ([]byte, error) {
	payload := map[string]interface{}{
		"query": query,
	}
	if variables != nil {
		payload["variables"] = variables
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", l.baseURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", l.apiKey)

	resp, err := l.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	responseBody := buf.Bytes()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"linear API request failed with status %d: %s",
			resp.StatusCode,
			string(responseBody),
		)
	}

	// Check for GraphQL errors
	var errorCheck struct {
		Errors []struct {
			Message string        `json:"message"`
			Path    []interface{} `json:"path"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(responseBody, &errorCheck); err == nil && len(errorCheck.Errors) > 0 {
		errorMsg := errorCheck.Errors[0].Message
		if len(errorCheck.Errors[0].Path) > 0 {
			errorMsg = fmt.Sprintf("%s (path: %v)", errorMsg, errorCheck.Errors[0].Path)
		}
		return nil, fmt.Errorf("GraphQL error: %s", errorMsg)
	}

	return responseBody, nil
}

func (l *LinearTicketManager) getTeamID() (string, error) {
	if l.teamID != "" {
		return l.teamID, nil
	}

	query := `
		query GetTeams {
			teams {
				nodes {
					id
					key
					name
				}
			}
		}
	`

	resp, err := l.makeRequest(query, nil)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			Teams struct {
				Nodes []struct {
					ID   string `json:"id"`
					Key  string `json:"key"`
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"teams"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	for _, team := range result.Data.Teams.Nodes {
		if team.Key == l.teamKey {
			l.teamID = team.ID
			return team.ID, nil
		}
	}

	return "", fmt.Errorf("team with key %s not found", l.teamKey)
}

func (l *LinearTicketManager) CreateEpic(title, description string) (string, error) {
	teamID, err := l.getTeamID()
	if err != nil {
		return "", err
	}

	query := `
		mutation CreateProject($input: ProjectCreateInput!) {
			projectCreate(input: $input) {
				success
				project {
					id
					name
					description
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"name":        title,
			"description": description,
			"teamIds":     []string{teamID},
		},
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			ProjectCreate struct {
				Success bool `json:"success"`
				Project struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"project"`
			} `json:"projectCreate"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if !result.Data.ProjectCreate.Success {
		return "", fmt.Errorf("failed to create project in Linear")
	}

	return result.Data.ProjectCreate.Project.ID, nil
}

func (l *LinearTicketManager) UpdateEpic(title, description string, epicID string) (string, error) {
	query := `
		mutation UpdateProject($id: String!, $input: ProjectUpdateInput!) {
			projectUpdate(id: $id, input: $input) {
				success
				project {
					id
					name
					description
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": epicID,
		"input": map[string]interface{}{
			"name":        title,
			"description": description,
		},
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			ProjectUpdate struct {
				Success bool `json:"success"`
				Project struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
				} `json:"project"`
			} `json:"projectUpdate"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if !result.Data.ProjectUpdate.Success {
		return "", fmt.Errorf("failed to update project in Linear")
	}

	return result.Data.ProjectUpdate.Project.ID, nil
}

func (l *LinearTicketManager) ListEpics() ([]Epic, error) {
	teamID, err := l.getTeamID()
	if err != nil {
		return nil, err
	}

	query := `
		query GetProjects($filter: ProjectFilter!) {
			projects(filter: $filter) {
				nodes {
					id
					name
					description
					state
					creator {
						id
						email
						displayName
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"teams": map[string]interface{}{
				"some": map[string]interface{}{
					"id": map[string]interface{}{
						"eq": teamID,
					},
				},
			},
		},
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Projects struct {
				Nodes []struct {
					ID          string `json:"id"`
					Name        string `json:"name"`
					Description string `json:"description"`
					State       string `json:"state"`
					Creator     struct {
						ID          string `json:"id"`
						Email       string `json:"email"`
						DisplayName string `json:"displayName"`
					} `json:"creator"`
				} `json:"nodes"`
			} `json:"projects"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	epics := make([]Epic, len(result.Data.Projects.Nodes))
	for i, project := range result.Data.Projects.Nodes {
		epics[i] = Epic{
			ID:          project.ID,
			Key:         project.ID, // Projects don't have identifiers like issues do
			Title:       project.Name,
			Description: project.Description,
			Status:      project.State,
			Creator: User{
				ID:          project.Creator.ID,
				Email:       project.Creator.Email,
				DisplayName: project.Creator.DisplayName,
			},
		}
	}

	return epics, nil
}

func (l *LinearTicketManager) ListUserEpics() ([]Epic, error) {
	query := `
		query GetUserProjects {
			viewer {
				createdProjects {
					nodes {
						id
						name
						description
						state
						creator {
							id
							email
							displayName
						}
					}
				}
			}
		}
	`

	resp, err := l.makeRequest(query, nil)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Viewer struct {
				CreatedProjects struct {
					Nodes []struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						Description string `json:"description"`
						State       string `json:"state"`
						Creator     struct {
							ID          string `json:"id"`
							Email       string `json:"email"`
							DisplayName string `json:"displayName"`
						} `json:"creator"`
					} `json:"nodes"`
				} `json:"createdProjects"`
			} `json:"viewer"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	epics := make([]Epic, len(result.Data.Viewer.CreatedProjects.Nodes))
	for i, project := range result.Data.Viewer.CreatedProjects.Nodes {
		epics[i] = Epic{
			ID:          project.ID,
			Key:         project.ID,
			Title:       project.Name,
			Description: project.Description,
			Status:      project.State,
			Creator: User{
				ID:          project.Creator.ID,
				Email:       project.Creator.Email,
				DisplayName: project.Creator.DisplayName,
			},
		}
	}

	return epics, nil
}

func (l *LinearTicketManager) DescribeEpic(epicID string) (*Epic, error) {
	query := `
		query GetProject($id: String!) {
			project(id: $id) {
				id
				name
				description
				state
				creator {
					id
					email
					displayName
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": epicID,
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Project struct {
				ID          string `json:"id"`
				Name        string `json:"name"`
				Description string `json:"description"`
				State       string `json:"state"`
				Creator     struct {
					ID          string `json:"id"`
					Email       string `json:"email"`
					DisplayName string `json:"displayName"`
				} `json:"creator"`
			} `json:"project"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	project := result.Data.Project
	return &Epic{
		ID:          project.ID,
		Key:         project.ID,
		Title:       project.Name,
		Description: project.Description,
		Status:      project.State,
		Creator: User{
			ID:          project.Creator.ID,
			Email:       project.Creator.Email,
			DisplayName: project.Creator.DisplayName,
		},
	}, nil
}

func (l *LinearTicketManager) CreateTask(title, description string, epicID string) (string, error) {
	teamID, err := l.getTeamID()
	if err != nil {
		return "", err
	}

	query := `
		mutation CreateIssue($input: IssueCreateInput!) {
			issueCreate(input: $input) {
				success
				issue {
					id
					identifier
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       title,
			"description": description,
			"teamId":      teamID,
			"projectId":   epicID, // Assign issue to project
		},
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return "", err
	}

	var result struct {
		Data struct {
			IssueCreate struct {
				Success bool `json:"success"`
				Issue   struct {
					ID         string `json:"id"`
					Identifier string `json:"identifier"`
					Title      string `json:"title"`
				} `json:"issue"`
			} `json:"issueCreate"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return "", err
	}

	if !result.Data.IssueCreate.Success {
		return "", fmt.Errorf("failed to create task in Linear")
	}

	return result.Data.IssueCreate.Issue.Identifier, nil
}

func (l *LinearTicketManager) UpdateTask(
	title, description string,
	completed bool,
	taskID string,
) error {
	query := `
		mutation UpdateIssue($id: String!, $input: IssueUpdateInput!) {
			issueUpdate(id: $id, input: $input) {
				success
			}
		}
	`

	input := map[string]interface{}{
		"title":       title,
		"description": description,
	}

	if completed {
		// Get completed state ID - this would need to be configured per team
		// For now, we'll assume there's a "Done" state
		stateQuery := `
			query GetStates($teamId: String!) {
				team(id: $teamId) {
					states {
						nodes {
							id
							name
						}
					}
				}
			}
		`

		teamID, err := l.getTeamID()
		if err != nil {
			return err
		}

		stateResp, err := l.makeRequest(stateQuery, map[string]interface{}{"teamId": teamID})
		if err != nil {
			return err
		}

		var stateResult struct {
			Data struct {
				Team struct {
					States struct {
						Nodes []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"nodes"`
					} `json:"states"`
				} `json:"team"`
			} `json:"data"`
		}

		err = json.Unmarshal(stateResp, &stateResult)
		if err != nil {
			return err
		}

		// Find Done state
		for _, state := range stateResult.Data.Team.States.Nodes {
			if state.Name == "Done" || state.Name == "Completed" {
				input["stateId"] = state.ID
				break
			}
		}
	}

	variables := map[string]interface{}{
		"id":    taskID,
		"input": input,
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return err
	}

	var result struct {
		Data struct {
			IssueUpdate struct {
				Success bool `json:"success"`
			} `json:"issueUpdate"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return err
	}

	if !result.Data.IssueUpdate.Success {
		return fmt.Errorf("failed to update task in Linear")
	}

	return nil
}

func (l *LinearTicketManager) ListTasksForEpic(epicID string) ([]Task, error) {
	query := `
		query GetProjectIssues($filter: IssueFilter!) {
			issues(filter: $filter) {
				nodes {
					id
					identifier
					title
					description
					state {
						name
					}
					project {
						id
					}
					creator {
						id
						email
						displayName
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"project": map[string]interface{}{
				"id": map[string]interface{}{
					"eq": epicID,
				},
			},
		},
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Issues struct {
				Nodes []struct {
					ID          string `json:"id"`
					Identifier  string `json:"identifier"`
					Title       string `json:"title"`
					Description string `json:"description"`
					State       struct {
						Name string `json:"name"`
					} `json:"state"`
					Project struct {
						ID string `json:"id"`
					} `json:"project"`
					Creator struct {
						ID          string `json:"id"`
						Email       string `json:"email"`
						DisplayName string `json:"displayName"`
					} `json:"creator"`
				} `json:"nodes"`
			} `json:"issues"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	tasks := make([]Task, len(result.Data.Issues.Nodes))
	for i, issue := range result.Data.Issues.Nodes {
		tasks[i] = Task{
			ID:          issue.ID,
			Key:         issue.Identifier,
			Title:       issue.Title,
			Description: issue.Description,
			Status:      issue.State.Name,
			Completed:   issue.State.Name == "Done" || issue.State.Name == "Completed",
			EpicID:      issue.Project.ID,
			Creator: User{
				ID:          issue.Creator.ID,
				Email:       issue.Creator.Email,
				DisplayName: issue.Creator.DisplayName,
			},
		}
	}

	return tasks, nil
}

func (l *LinearTicketManager) DescribeTask(taskID string) (*Task, error) {
	query := `
		query GetIssue($id: String!) {
			issue(id: $id) {
				id
				identifier
				title
				description
				state {
					name
				}
				project {
					id
				}
				creator {
					id
					email
					displayName
				}
			}
		}
	`

	variables := map[string]interface{}{
		"id": taskID,
	}

	resp, err := l.makeRequest(query, variables)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Issue struct {
				ID          string `json:"id"`
				Identifier  string `json:"identifier"`
				Title       string `json:"title"`
				Description string `json:"description"`
				State       struct {
					Name string `json:"name"`
				} `json:"state"`
				Project struct {
					ID string `json:"id"`
				} `json:"project"`
				Creator struct {
					ID          string `json:"id"`
					Email       string `json:"email"`
					DisplayName string `json:"displayName"`
				} `json:"creator"`
			} `json:"issue"`
		} `json:"data"`
	}

	err = json.Unmarshal(resp, &result)
	if err != nil {
		return nil, err
	}

	issue := result.Data.Issue
	return &Task{
		ID:          issue.ID,
		Key:         issue.Identifier,
		Title:       issue.Title,
		Description: issue.Description,
		Status:      issue.State.Name,
		Completed:   issue.State.Name == "Done" || issue.State.Name == "Completed",
		EpicID:      issue.Project.ID,
		Creator: User{
			ID:          issue.Creator.ID,
			Email:       issue.Creator.Email,
			DisplayName: issue.Creator.DisplayName,
		},
	}, nil
}

func GetLinearConfig() LinearConfig {
	return LinearConfig{
		APIKey:  util.GetEnvStr("LINEAR_API_KEY", ""),
		TeamKey: util.GetEnvStr("LINEAR_TEAM_KEY", ""),
	}
}

func (c LinearConfig) IsConfigured() bool {
	return c.APIKey != "" && c.TeamKey != ""
}

// parseIssueNumber extracts the numeric part from a Linear issue identifier
// e.g., "ENG-123" -> 123
func parseIssueNumber(identifier string) int {
	re := regexp.MustCompile(`(\d+)$`)
	matches := re.FindStringSubmatch(identifier)
	if len(matches) > 1 {
		if num, err := strconv.Atoi(matches[1]); err == nil {
			return num
		}
	}
	return 0
}
