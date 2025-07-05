package util

import (
	"os"
	"strconv"
	"github.com/subosito/gotenv"
)

func init() {
	gotenv.Load()
}

// GetEnvInt finds an ENV variable and converts to int, otherwise return default value
func GetEnvInt(key string, defaultVal int) int {
	var err error
	intVal := defaultVal

	if v, ok := os.LookupEnv(key); ok {
		if intVal, err = strconv.Atoi(v); err != nil {
			LogError("Failed to convert env var %s to int: %v", key, err)
			return defaultVal
		}
	}

	return intVal
}

// GetEnvStr finds an ENV variable, otherwise return default value
func GetEnvStr(key, defaultVal string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}

	return defaultVal
}

// JiraConfig holds JIRA configuration
type JiraConfig struct {
	URL        string
	Username   string
	APIToken   string
	ProjectKey string
}

// GetJiraConfig returns JIRA configuration from environment variables
func GetJiraConfig() JiraConfig {
	config := JiraConfig{
		URL:        GetEnvStr("JIRA_URL", ""),
		Username:   GetEnvStr("JIRA_USERNAME", ""),
		APIToken:   GetEnvStr("JIRA_API_TOKEN", ""),
		ProjectKey: GetEnvStr("JIRA_PROJECT_KEY", ""),
	}


	return config
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return "***"
	}
	return token[:4] + "***" + token[len(token)-4:]
}

// IsJiraConfigured checks if JIRA is properly configured
func (c JiraConfig) IsConfigured() bool {
	return c.URL != "" && c.Username != "" && c.APIToken != "" && c.ProjectKey != ""
}
