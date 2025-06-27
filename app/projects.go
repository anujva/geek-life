package main

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/ajaxray/geek-life/jira"
	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
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
	jira                jira.Jira
	jiraConfig          util.JiraConfig
}

// NewProjectPane initializes
func NewProjectPane(repo repository.ProjectRepository) *ProjectPane {
	jiraConfig := util.GetJiraConfig()

	pane := ProjectPane{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		list:       tview.NewList().ShowSecondaryText(false),
		newProject: makeLightTextInput("+[New Project]"),
		repo:       repo,
		jiraConfig: jiraConfig,
	}

	if jiraConfig.IsConfigured() {
		pane.jira = jira.NewJiraClient(
			jiraConfig.URL,
			jiraConfig.Username,
			jiraConfig.APIToken,
			jiraConfig.APIToken,
			jiraConfig.ProjectKey,
		)
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
	}

	switch event.Key() {
	case tcell.KeyCtrlJ:
		// Get the project that is currently selected
		selectedIndex := pane.list.GetCurrentItem()
		projectindex := selectedIndex - pane.projectListStarting
		if projectindex >= 0 && projectindex < len(pane.projects) && pane.jira != nil {
			project := pane.projects[projectindex]
			if project.Jira == "" {
				p, err := pane.jira.CreateEpic(project.Title, project.Title)
				if err != nil {
					statusBar.showForSeconds("[red]Failed to create JIRA epic: "+err.Error(), 5)
					return nil
				}
				project.Jira = p
				_ = pane.repo.Update(&project)
				statusBar.showForSeconds("[lime]Created JIRA epic: "+p, 5)
			}
			pane.loadListItems(true)
			pane.list.SetCurrentItem(selectedIndex)
		}
		return nil
	case tcell.KeyCtrlI:
		// Import epics from JIRA
		pane.importEpicsFromJira()
		return nil
	}

	return event
}

func (pane *ProjectPane) activateProject(idx int) {
	pane.activeProject = &pane.projects[idx]
	taskPane.LoadProjectTasks(*pane.activeProject)

	removeThirdCol()
	projectDetailPane.SetProject(pane.activeProject)
	contents.AddItem(projectDetailPane, 25, 0, false)
	app.SetFocus(taskPane)
}

// RemoveActivateProject deletes the currently active project
func (pane *ProjectPane) RemoveActivateProject() {
	if pane.activeProject != nil && pane.repo.Delete(pane.activeProject) == nil {

		for i := range taskPane.tasks {
			_ = taskRepo.Delete(&taskPane.tasks[i])
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

// importEpicsFromJira imports all epics from JIRA as projects
func (pane *ProjectPane) importEpicsFromJira() {
	if pane.jira == nil {
		statusBar.showForSeconds(
			"[red]JIRA not configured. Set JIRA_URL, JIRA_USERNAME, JIRA_API_TOKEN, and JIRA_PROJECT_KEY environment variables.",
			8,
		)
		return
	}

	epics, err := pane.jira.ListEpics()
	if err != nil {
		statusBar.showForSeconds("[red]Failed to fetch epics from JIRA: "+err.Error(), 5)
		return
	}

	imported := 0
	for _, epic := range epics {
		// Check if project already exists with this JIRA ID
		existingProject := pane.findProjectByJiraID(epic.Key)
		if existingProject != nil {
			continue
		}

		// Create new project from epic
		project, err := pane.repo.Create(epic.Fields.Summary, epic.Key)
		if err != nil {
			continue
		}

		// Import tasks for this epic
		pane.importTasksForEpic(project, epic.Key)

		imported++
	}

	if imported > 0 {
		statusBar.showForSeconds(fmt.Sprintf("[lime]Imported %d epics from JIRA", imported), 5)
		pane.loadListItems(true)
	} else {
		statusBar.showForSeconds("[yellow]No new epics to import", 3)
	}
}

// importTasksForEpic imports tasks for a specific epic
func (pane *ProjectPane) importTasksForEpic(project model.Project, epicKey string) {
	tasks, err := pane.jira.ListTasksForEpic(epicKey)
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
			Title:     task.Fields.Summary,
			Details:   getTaskDescription(task),
			Completed: isTaskCompleted(task),
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

// getTaskDescription extracts description from JIRA task
func getTaskDescription(task jira.JiraIssue) string {
	if task.Fields.Description != nil {
		if desc, ok := task.Fields.Description.(string); ok {
			return desc
		}
	}
	return ""
}

// isTaskCompleted checks if JIRA task is completed
func isTaskCompleted(task jira.JiraIssue) bool {
	return task.Fields.Status.StatusCategory.Key == "done"
}
