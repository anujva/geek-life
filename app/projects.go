package main

import (
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
	"github.com/ajaxray/geek-life/ticketmanager"
	"github.com/ajaxray/geek-life/util"
)

// ProjectPane Displays projects and dynamic lists
type ProjectPane struct {
	*tview.Flex
	projects            []model.Project
	list                *tview.List
	newProject          *tview.InputField
	repo                repository.ProjectRepository
	activeProject       *model.Project
	projectListStarting int // The index in list where project names starts
	ticketManager       ticketmanager.TicketManager
	providerType        ticketmanager.ProviderType
	lastGKeyTime        int64 // Timestamp for tracking double 'g' press
}

// NewProjectPane initializes
func NewProjectPane(repo repository.ProjectRepository) *ProjectPane {
	pane := ProjectPane{
		Flex:         tview.NewFlex().SetDirection(tview.FlexRow),
		list:         tview.NewList().ShowSecondaryText(false),
		newProject:   makeLightTextInput("+[New Project]"),
		repo:         repo,
		providerType: ticketmanager.GetProviderType(),
	}

	if ticketmanager.IsAnyProviderConfigured() {
		if tm, err := ticketmanager.NewTicketManager(); err == nil {
			pane.ticketManager = tm
		}
	}

	pane.newProject.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			pane.addNewProject()
		case tcell.KeyEsc:
			app.SetFocus(projectPane)
		}
	})

	pane.AddItem(pane.list, 0, 1, true).
		AddItem(pane.newProject, 1, 0, false)

	pane.SetBorder(true).SetTitle("[::u]P[::-]rojects")
	pane.loadListItems(false)

	return &pane
}

func (pane *ProjectPane) addNewProject() {
	name := pane.newProject.GetText()
	if len(name) < 3 {
		statusBar.showForSeconds("[red::]Project name should be at least 3 character", 5)
		return
	}

	project, err := pane.repo.Create(name, "")
	if err != nil {
		statusBar.showForSeconds("[red::]Failed to create Project:"+err.Error(), 5)
	} else {
		statusBar.showForSeconds(fmt.Sprintf("[yellow::]Project %s created. Press n to start adding new tasks.", name), 10)
		pane.projects = append(pane.projects, project)
		pane.addProjectToList(len(pane.projects)-1, true)
		pane.newProject.SetText("")
	}
}

func (pane *ProjectPane) addDynamicLists() {
	pane.addSection("Dynamic Lists")
	pane.list.AddItem("- Today", "", 0, func() { taskPane.LoadDynamicList("today") })
	pane.list.AddItem("- Tomorrow", "", 0, func() { taskPane.LoadDynamicList("tomorrow") })
	pane.list.AddItem("- Upcoming", "", 0, func() { taskPane.LoadDynamicList("upcoming") })
	pane.list.AddItem("- Unscheduled", "", 0, func() { taskPane.LoadDynamicList("unscheduled") })
}

func (pane *ProjectPane) addProjectList() {
	pane.addSection("Projects")
	pane.projectListStarting = pane.list.GetItemCount()

	var err error
	pane.projects, err = pane.repo.GetAll()
	if err != nil {
		statusBar.showForSeconds("Could not load Projects: "+err.Error(), 5)
		return
	}

	for i := range pane.projects {
		pane.addProjectToList(i, false)
	}

	pane.list.SetCurrentItem(2) // Keep "Today" selected on start
}

func (pane *ProjectPane) addProjectToList(i int, selectItem bool) {
	// To avoid overriding of loop variables - https://www.calhoun.io/gotchas-and-common-mistakes-with-closures-in-go/
	pane.list.AddItem("- "+pane.projects[i].GetTitle(), "", 0, func(idx int) func() {
		return func() { pane.activateProject(idx) }
	}(i))

	if selectItem {
		pane.list.SetCurrentItem(-1)
		pane.activateProject(i)
	}
}

func (pane *ProjectPane) addSection(name string) {
	pane.list.AddItem("[::d]"+name, "", 0, nil)
	pane.list.AddItem("[::d]"+strings.Repeat(string(tcell.RuneHLine), 25), "", 0, nil)
}

func (pane *ProjectPane) handleShortcuts(event *tcell.EventKey) *tcell.EventKey {
	// Handle Shift+G (uppercase G) BEFORE the lowercase conversion
	if event.Rune() == 'G' {
		// Go to bottom
		itemCount := pane.list.GetItemCount()
		if itemCount > 0 {
			pane.list.SetCurrentItem(itemCount - 1)
		}
		return nil
	}

	switch unicode.ToLower(event.Rune()) {
	case 'j':
		pane.list.SetCurrentItem(pane.list.GetCurrentItem() + 1)
		return nil
	case 'k':
		pane.list.SetCurrentItem(pane.list.GetCurrentItem() - 1)
		return nil
	case 'n':
		app.SetFocus(pane.newProject)
		return nil
	case 'g':
		now := time.Now().UnixMilli()
		if now-pane.lastGKeyTime < 500 { // 500ms window for double press
			// Double 'g' pressed - go to top
			pane.list.SetCurrentItem(0)
			pane.lastGKeyTime = 0 // Reset
		} else {
			// Single 'g' pressed - store timestamp
			pane.lastGKeyTime = now
		}
		return nil
	}

	switch event.Key() {
	case tcell.KeyCtrlB:
		// Open epic in browser
		selectedIndex := pane.list.GetCurrentItem()
		projectindex := selectedIndex - pane.projectListStarting
		if projectindex >= 0 && projectindex < len(pane.projects) {
			project := pane.projects[projectindex]
			if project.Jira != "" && pane.ticketManager != nil {
				var ticketURL string
				var providerName string

				switch pane.providerType {
				case ticketmanager.ProviderJira:
					jiraConfig := util.GetJiraConfig()
					ticketURL = fmt.Sprintf("%s/browse/%s", jiraConfig.URL, project.Jira)
					providerName = "JIRA"
				case ticketmanager.ProviderLinear:
					ticketURL = fmt.Sprintf("https://linear.app/team/issue/%s", project.Jira)
					providerName = "Linear"
				}

				err := util.OpenInBrowser(ticketURL)
				if err != nil {
					statusBar.showForSeconds("[red]Failed to open browser: "+err.Error(), 5)
				} else {
					statusBar.showForSeconds(fmt.Sprintf("[lime]Opened %s epic in browser", providerName), 3)
				}
			} else if project.Jira == "" {
				statusBar.showForSeconds("[yellow]Project has no ticket associated", 3)
			} else {
				statusBar.showForSeconds("[yellow]Ticket manager not configured", 3)
			}
		}
		return nil
	case tcell.KeyCtrlJ:
		// Get the project that is currently selected
		selectedIndex := pane.list.GetCurrentItem()
		projectindex := selectedIndex - pane.projectListStarting
		if projectindex >= 0 && projectindex < len(pane.projects) && pane.ticketManager != nil {
			project := pane.projects[projectindex]
			if project.Jira == "" {
				ticketID, err := pane.ticketManager.CreateEpic(project.Title, project.Title)
				if err != nil {
					statusBar.showForSeconds("[red]Failed to create epic: "+err.Error(), 5)
					return nil
				}
				project.Jira = ticketID
				_ = pane.repo.Update(&project)

				providerName := string(pane.providerType)
				statusBar.showForSeconds(
					fmt.Sprintf("[lime]Created %s epic: %s", providerName, ticketID),
					5,
				)
			}
			pane.loadListItems(true)
			pane.list.SetCurrentItem(selectedIndex)
		}
		return nil
	case tcell.KeyCtrlI:
		// Import epics from ticket manager
		pane.importEpicsFromTicketManager()
		return nil
	case tcell.KeyCtrlR:
		// Clean up duplicate projects and re-link existing ones
		pane.cleanupAndRelinkProjects()
		return nil
	case tcell.KeyCtrlT:
		// Force refresh tasks for current project from ticket system
		pane.forceRefreshTasks()
		return nil
	case tcell.KeyCtrlF:
		// Fix orphaned tasks - relink tasks to current project
		pane.fixOrphanedTasks()
		return nil
	}

	return event
}

func (pane *ProjectPane) activateProject(idx int) {
	pane.activeProject = &pane.projects[idx]

	// If this project has a ticket ID but no tasks, try to import them
	if pane.activeProject.Jira != "" && pane.ticketManager != nil {
		existingTasks, err := taskRepo.GetAllByProject(*pane.activeProject)
		if err == nil && len(existingTasks) == 0 {
			// No tasks exist for this project, try to import from ticket manager
			providerName := string(pane.providerType)
			statusBar.showForSeconds(
				fmt.Sprintf("[yellow]Loading tasks from %s...", providerName),
				2,
			)
			pane.importTasksForEpic(*pane.activeProject, pane.activeProject.Jira)
		}
	}

	taskPane.LoadProjectTasks(*pane.activeProject)

	removeThirdCol()
	projectDetailPane.SetProject(pane.activeProject)
	contents.AddItem(projectDetailPane, 25, 0, false)
	app.SetFocus(taskPane)
}

// RemoveActivateProject deletes the currently active project
func (pane *ProjectPane) RemoveActivateProject() {
	if pane.activeProject != nil {
		// Delete all tasks associated with this project first
		err := taskRepo.DeleteAllByProjectID(pane.activeProject.ID)
		if err != nil {
			statusBar.showForSeconds("[red]Failed to delete project tasks: "+err.Error(), 5)
			return
		}

		// Delete the project itself
		err = pane.repo.Delete(pane.activeProject)
		if err != nil {
			statusBar.showForSeconds("[red]Failed to delete project: "+err.Error(), 5)
			return
		}

		taskPane.ClearList()
		statusBar.showForSeconds("[lime]Removed Project: "+pane.activeProject.Title, 5)
		removeThirdCol()
		pane.loadListItems(true)
	}
}

func (pane *ProjectPane) loadListItems(focus bool) {
	pane.list.Clear()
	pane.addDynamicLists()
	pane.list.AddItem("", "", 0, nil)
	pane.addProjectList()

	if focus {
		app.SetFocus(pane)
	}
}

// GetActiveProject provides pointer to currently active project
func (pane *ProjectPane) GetActiveProject() *model.Project {
	return pane.activeProject
}

// importEpicsFromTicketManager imports all epics from ticket manager as projects
func (pane *ProjectPane) importEpicsFromTicketManager() {
	if pane.ticketManager == nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(
			fmt.Sprintf(
				"[red]%s not configured. Set required environment variables.",
				providerName,
			),
			8,
		)
		return
	}

	// Reload projects to ensure we have the latest data
	var err error
	pane.projects, err = pane.repo.GetAll()
	if err != nil {
		statusBar.showForSeconds("[red]Failed to load projects: "+err.Error(), 5)
		return
	}

	epics, err := pane.ticketManager.ListUserEpics()
	if err != nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(
			fmt.Sprintf("[red]Failed to fetch epics from %s: %s", providerName, err.Error()),
			5,
		)
		return
	}

	imported := 0
	updated := 0
	for _, epic := range epics {
		// Check if project already exists with this ticket ID
		existingProject := pane.findProjectByJiraID(epic.Key)
		if existingProject != nil {
			continue
		}

		// Check if there's a project with the same title but no ticket ID
		existingProjectByTitle := pane.findProjectByTitle(epic.Title)
		if existingProjectByTitle != nil && existingProjectByTitle.Jira == "" {
			// Update existing project with ticket ID
			existingProjectByTitle.Jira = epic.Key
			err := pane.repo.Update(existingProjectByTitle)
			if err == nil {
				// Update the in-memory project list
				for i := range pane.projects {
					if pane.projects[i].ID == existingProjectByTitle.ID {
						pane.projects[i].Jira = epic.Key
						break
					}
				}
				// Import tasks for this epic using the updated project
				pane.importTasksForEpic(*existingProjectByTitle, epic.Key)
				updated++
			}
			continue
		}

		// Create new project from epic with ticket ID
		project, err := pane.repo.CreateWithJira(epic.Title, epic.Key)
		if err != nil {
			continue
		}

		// Add to in-memory projects list
		pane.projects = append(pane.projects, project)

		// Import tasks for this epic
		pane.importTasksForEpic(project, epic.Key)

		imported++
	}

	if imported > 0 || updated > 0 {
		providerName := string(pane.providerType)
		message := ""
		if imported > 0 && updated > 0 {
			message = fmt.Sprintf(
				"[lime]Imported %d new epics and updated %d existing projects with %s IDs",
				imported,
				updated,
				providerName,
			)
		} else if imported > 0 {
			message = fmt.Sprintf("[lime]Imported %d user-created epics from %s", imported, providerName)
		} else {
			message = fmt.Sprintf("[lime]Updated %d existing projects with %s IDs", updated, providerName)
		}
		statusBar.showForSeconds(message, 5)
		pane.loadListItems(true)
	} else {
		statusBar.showForSeconds("[yellow]No new user-created epics to import", 3)
	}
}

// importTasksForEpic imports tasks for a specific epic
func (pane *ProjectPane) importTasksForEpic(project model.Project, epicKey string) {
	if pane.ticketManager == nil {
		return
	}

	tasks, err := pane.ticketManager.ListTasksForEpic(epicKey)
	if err != nil {
		return
	}

	for _, task := range tasks {
		// Check if task already exists
		existing, err := taskRepo.GetByJiraID(task.Key)
		if err == nil && existing != nil {
			continue
		}

		// Create task
		newTask := model.Task{
			ProjectID: project.ID,
			Title:     task.Title,
			Details:   task.Description,
			Completed: task.Completed,
			JiraID:    task.Key,
		}

		_ = taskRepo.CreateTask(&newTask)
	}
}

// findProjectByJiraID finds a project by its JIRA ID
func (pane *ProjectPane) findProjectByJiraID(jiraID string) *model.Project {
	for i := range pane.projects {
		if pane.projects[i].Jira == jiraID {
			return &pane.projects[i]
		}
	}
	return nil
}

// findProjectByTitle finds a project by its title
func (pane *ProjectPane) findProjectByTitle(title string) *model.Project {
	for i := range pane.projects {
		if pane.projects[i].Title == title {
			return &pane.projects[i]
		}
	}
	return nil
}

// cleanupAndRelinkProjects removes duplicate projects and links existing projects to ticket manager
func (pane *ProjectPane) cleanupAndRelinkProjects() {
	if pane.ticketManager == nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(
			fmt.Sprintf(
				"[red]%s not configured. Set required environment variables.",
				providerName,
			),
			8,
		)
		return
	}

	// Reload projects to ensure we have the latest data
	var err error
	pane.projects, err = pane.repo.GetAll()
	if err != nil {
		statusBar.showForSeconds("[red]Failed to load projects: "+err.Error(), 5)
		return
	}

	epics, err := pane.ticketManager.ListUserEpics()
	if err != nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(
			fmt.Sprintf("[red]Failed to fetch epics from %s: %s", providerName, err.Error()),
			5,
		)
		return
	}

	linkedCount := 0
	removedCount := 0

	for _, epic := range epics {
		// Find all projects with this epic title
		var projectsWithTitle []*model.Project
		var projectWithJira *model.Project

		for i := range pane.projects {
			if pane.projects[i].Title == epic.Title {
				if pane.projects[i].Jira == epic.Key {
					projectWithJira = &pane.projects[i]
				} else if pane.projects[i].Jira == "" {
					projectsWithTitle = append(projectsWithTitle, &pane.projects[i])
				}
			}
		}

		// If we have a project with JIRA ID and projects without, merge them
		if projectWithJira != nil && len(projectsWithTitle) > 0 {
			for _, oldProject := range projectsWithTitle {
				// Move tasks from old project to the JIRA-linked project
				tasks, err := taskRepo.GetAllByProject(*oldProject)
				if err == nil {
					for _, task := range tasks {
						task.ProjectID = projectWithJira.ID
						_ = taskRepo.Update(&task)
					}
				}
				// Remove the old project
				_ = pane.repo.Delete(oldProject)
				removedCount++
			}
		} else if len(projectsWithTitle) == 1 && projectWithJira == nil {
			// Link the existing project to JIRA
			projectsWithTitle[0].Jira = epic.Key
			err := pane.repo.Update(projectsWithTitle[0])
			if err == nil {
				// Import tasks for this epic
				pane.importTasksForEpic(*projectsWithTitle[0], epic.Key)
				linkedCount++
			}
		}
	}

	providerName := string(pane.providerType)
	statusBar.showForSeconds(
		fmt.Sprintf(
			"[lime]Linked %d projects to %s, removed %d duplicates",
			linkedCount,
			providerName,
			removedCount,
		),
		5,
	)
	pane.loadListItems(true)
}

// forceRefreshTasks forces a refresh of tasks for the currently selected project
func (pane *ProjectPane) forceRefreshTasks() {
	if pane.ticketManager == nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(fmt.Sprintf("[red]%s not configured", providerName), 3)
		return
	}

	selectedIndex := pane.list.GetCurrentItem()
	projectindex := selectedIndex - pane.projectListStarting
	if projectindex >= 0 && projectindex < len(pane.projects) {
		project := pane.projects[projectindex]
		if project.Jira == "" {
			statusBar.showForSeconds("[yellow]Project has no ticket associated", 3)
			return
		}

		providerName := string(pane.providerType)
		statusBar.showForSeconds(
			fmt.Sprintf("[yellow]Refreshing tasks from %s...", providerName),
			2,
		)
		pane.importTasksForEpic(project, project.Jira)

		// If this is the active project, reload its tasks
		if pane.activeProject != nil && pane.activeProject.ID == project.ID {
			taskPane.LoadProjectTasks(*pane.activeProject)
		}

		statusBar.showForSeconds(fmt.Sprintf("[lime]Tasks refreshed from %s", providerName), 3)
	} else {
		statusBar.showForSeconds("[yellow]Select a project first", 3)
	}
}

// fixOrphanedTasks fixes tasks that exist with ticket IDs but wrong ProjectIDs
func (pane *ProjectPane) fixOrphanedTasks() {
	if pane.ticketManager == nil {
		providerName := string(pane.providerType)
		statusBar.showForSeconds(fmt.Sprintf("[red]%s not configured", providerName), 3)
		return
	}

	selectedIndex := pane.list.GetCurrentItem()
	projectindex := selectedIndex - pane.projectListStarting
	if projectindex >= 0 && projectindex < len(pane.projects) {
		project := pane.projects[projectindex]
		if project.Jira == "" {
			statusBar.showForSeconds("[yellow]Project has no ticket associated", 3)
			return
		}

		statusBar.showForSeconds("[yellow]Finding and fixing orphaned tasks...", 2)

		// Get all tasks for this epic from ticket manager
		tasks, err := pane.ticketManager.ListTasksForEpic(project.Jira)
		if err != nil {
			providerName := string(pane.providerType)
			statusBar.showForSeconds(
				fmt.Sprintf("[red]Error getting tasks from %s: %s", providerName, err.Error()),
				5,
			)
			return
		}

		fixed := 0
		for _, task := range tasks {
			// Check if task exists with wrong ProjectID
			existing, err := taskRepo.GetByJiraID(task.Key)
			if err == nil && existing != nil && existing.ProjectID != project.ID {
				// Update the ProjectID
				existing.ProjectID = project.ID
				err = taskRepo.Update(existing)
				if err == nil {
					fixed++
				}
			}
		}

		if fixed > 0 {
			statusBar.showForSeconds(fmt.Sprintf("[lime]Fixed %d orphaned tasks", fixed), 5)
			// Reload tasks to show the fixed ones
			if pane.activeProject != nil && pane.activeProject.ID == project.ID {
				taskPane.LoadProjectTasks(*pane.activeProject)
			}
		} else {
			statusBar.showForSeconds("[yellow]No orphaned tasks found", 3)
		}
	} else {
		statusBar.showForSeconds("[yellow]Select a project first", 3)
	}
}
