package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ajaxray/geek-life/repository/storm"
	"github.com/ajaxray/geek-life/ticketmanager"
	"github.com/ajaxray/geek-life/util"
)

// BackfillProjectDates backfills JIRA creation dates for existing projects
func main() {
	fmt.Println("🔄 Backfilling JIRA creation dates for existing projects...")

	// Initialize database connection
	db := util.ConnectStorm("")
	defer db.Close()

	// Initialize repository
	projectRepo := storm.NewProjectRepository(db)

	// Check if ticket manager is configured
	if !ticketmanager.IsAnyProviderConfigured() {
		log.Fatal("❌ No ticket provider configured. Please set JIRA or Linear environment variables.")
	}

	// Initialize ticket manager
	ticketManager, err := ticketmanager.NewTicketManager()
	if err != nil {
		log.Fatalf("❌ Failed to initialize ticket manager: %v", err)
	}

	// Get all projects
	projects, err := projectRepo.GetAll()
	if err != nil {
		log.Fatalf("❌ Failed to get projects: %v", err)
	}

	fmt.Printf("📋 Found %d projects to check\n", len(projects))

	updatedCount := 0
	errorCount := 0

	for _, project := range projects {
		// Skip projects without JIRA IDs
		if project.Jira == "" {
			continue
		}

		// Skip projects that already have creation dates
		if project.JiraCreatedDate != nil {
			continue
		}

		fmt.Printf("🔍 Processing project: %s (JIRA: %s)\n", project.Title, project.Jira)

		// Fetch epic details to get creation date
		epicDetails, err := ticketManager.DescribeEpic(project.Jira)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to fetch epic details: %v\n", err)
			errorCount++
			continue
		}

		// Parse creation date
		createdDate, err := parseJiraDate(epicDetails.CreatedDate)
		if err != nil {
			fmt.Printf("   ⚠️  Failed to parse creation date '%s': %v\n", epicDetails.CreatedDate, err)
			errorCount++
			continue
		}

		// Update project with creation date
		project.JiraCreatedDate = &createdDate
		err = projectRepo.Update(&project)
		if err != nil {
			fmt.Printf("   ❌ Failed to update project: %v\n", err)
			errorCount++
			continue
		}

		fmt.Printf("   ✅ Updated creation date: %s\n", createdDate.Format("2006-01-02 15:04:05"))
		updatedCount++
	}

	fmt.Printf("\n🎉 Backfill complete!\n")
	fmt.Printf("   ✅ Updated: %d projects\n", updatedCount)
	if errorCount > 0 {
		fmt.Printf("   ⚠️  Errors: %d projects\n", errorCount)
	}

	if updatedCount > 0 {
		fmt.Println("\n💡 Tip: Restart the application to see projects sorted by creation date")
	}
}

// parseJiraDate parses JIRA's ISO 8601 date format
func parseJiraDate(dateStr string) (time.Time, error) {
	// JIRA typically returns dates in RFC3339 format like "2023-10-15T14:30:00.000+0000"
	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05.000-0700",
		"2006-01-02T15:04:05.000+0000",
		"2006-01-02T15:04:05-0700",
		"2006-01-02T15:04:05+0000",
	}
	
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}