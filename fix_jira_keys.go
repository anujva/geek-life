package main

import (
	"fmt"
	"log"
	"regexp"

	"github.com/ajaxray/geek-life/jira"
	"github.com/ajaxray/geek-life/repository/storm"
	"github.com/ajaxray/geek-life/util"
)

func main() {
	fmt.Println("Starting JIRA key fix script...")

	// Get JIRA configuration
	jiraConfig := util.GetJiraConfig()
	if !jiraConfig.IsConfigured() {
		log.Fatal("JIRA not configured. Set JIRA_URL, JIRA_USERNAME, JIRA_API_TOKEN, and JIRA_PROJECT_KEY environment variables.")
	}

	// Connect to database
	db := util.ConnectStorm("")
	defer db.Close()

	// Initialize repositories
	projectRepo := storm.NewProjectRepository(db)
	
	// Initialize JIRA client
	jiraClient := jira.NewJiraClient(
		jiraConfig.URL,
		jiraConfig.Username,
		jiraConfig.APIToken,
		jiraConfig.APIToken,
		jiraConfig.ProjectKey,
	)

	// Get all projects
	projects, err := projectRepo.GetAll()
	if err != nil {
		log.Fatalf("Failed to get projects: %v", err)
	}

	fmt.Printf("Found %d projects to check\n", len(projects))

	// Regex to check if Jira field contains only digits (indicating it's an internal ID)
	numericRegex := regexp.MustCompile(`^\d+$`)
	
	fixed := 0
	errors := 0

	for _, project := range projects {
		if project.Jira == "" {
			fmt.Printf("Project '%s' has no JIRA ID, skipping\n", project.Title)
			continue
		}

		// Check if this looks like a numeric ID (internal JIRA ID) rather than a key (PROJ-123)
		if !numericRegex.MatchString(project.Jira) {
			fmt.Printf("Project '%s' already has a proper JIRA key: %s\n", project.Title, project.Jira)
			continue
		}

		fmt.Printf("Project '%s' has numeric JIRA ID: %s, attempting to fix...\n", project.Title, project.Jira)

		// Try to get the epic details using the numeric ID
		epic, err := jiraClient.DescribeEpic(project.Jira)
		if err != nil {
			fmt.Printf("ERROR: Could not fetch epic details for ID %s: %v\n", project.Jira, err)
			errors++
			continue
		}

		if epic.Key == "" {
			fmt.Printf("ERROR: Epic %s has no key\n", project.Jira)
			errors++
			continue
		}

		// Update the project with the correct key
		project.Jira = epic.Key
		err = projectRepo.Update(&project)
		if err != nil {
			fmt.Printf("ERROR: Could not update project '%s': %v\n", project.Title, err)
			errors++
			continue
		}

		fmt.Printf("SUCCESS: Updated project '%s' from ID %s to key %s\n", project.Title, epic.ID, epic.Key)
		fixed++
	}

	fmt.Printf("\nFix completed!\n")
	fmt.Printf("Fixed: %d projects\n", fixed)
	fmt.Printf("Errors: %d projects\n", errors)
	
	if fixed > 0 {
		fmt.Println("\nYour projects should now have correct JIRA keys and Ctrl+B should work properly!")
	}
}