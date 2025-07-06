package ticketmanager

import (
	"fmt"
	"strings"

	"github.com/ajaxray/geek-life/util"
)

type ProviderType string

const (
	ProviderJira   ProviderType = "jira"
	ProviderLinear ProviderType = "linear"
)

func NewTicketManager() (TicketManager, error) {
	provider := strings.ToLower(util.GetEnvStr("TICKET_PROVIDER", "jira"))

	switch ProviderType(provider) {
	case ProviderJira:
		jiraConfig := util.GetJiraConfig()
		if !jiraConfig.IsConfigured() {
			return nil, fmt.Errorf(
				"JIRA is not configured. Please set JIRA_URL, JIRA_USERNAME, JIRA_API_TOKEN, and JIRA_PROJECT_KEY environment variables",
			)
		}
		return NewJiraTicketManager(jiraConfig), nil

	case ProviderLinear:
		linearConfig := GetLinearConfig()
		if !linearConfig.IsConfigured() {
			return nil, fmt.Errorf(
				"Linear is not configured. Please set LINEAR_API_KEY and LINEAR_TEAM_KEY environment variables",
			)
		}
		return NewLinearTicketManager(linearConfig), nil

	default:
		return nil, fmt.Errorf(
			"unknown ticket provider: %s. Supported providers are: jira, linear",
			provider,
		)
	}
}

func GetProviderType() ProviderType {
	provider := strings.ToLower(util.GetEnvStr("TICKET_PROVIDER", "jira"))
	return ProviderType(provider)
}

func IsAnyProviderConfigured() bool {
	jiraConfig := util.GetJiraConfig()
	linearConfig := GetLinearConfig()

	return jiraConfig.IsConfigured() || linearConfig.IsConfigured()
}
