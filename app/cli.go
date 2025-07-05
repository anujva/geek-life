package main

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/asdine/storm/v3"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	flag "github.com/spf13/pflag"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
	repo "github.com/ajaxray/geek-life/repository/storm"
	"github.com/ajaxray/geek-life/util"
)

var (
	app              *tview.Application
	layout, contents *tview.Flex

	statusBar         *StatusBar
	projectPane       *ProjectPane
	taskPane          *TaskPane
	taskDetailPane    *TaskDetailPane
	projectDetailPane *ProjectDetailPane

	db          *storm.DB
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository

	// Flag variables
	dbFile string
)

func init() {
	flag.StringVarP(&dbFile, "db-file", "d", "", "Specify DB file path manually.")
}

func main() {
	app = tview.NewApplication()
	flag.Parse()

	// Initialize logging system
	if err := util.InitLogger(); err != nil {
		fmt.Printf("Warning: Failed to initialize logger: %v\n", err)
	}

	db = util.ConnectStorm(dbFile)
	defer func() {
		if err := db.Close(); err != nil {
			util.LogIfError(err, "Error in closing storm Db")
		}
	}()

	if flag.NArg() > 0 && flag.Arg(0) == "migrate" {
		migrate(db)
		fmt.Println("Database migrated successfully!")
	} else {
		projectRepo = repo.NewProjectRepository(db)
		taskRepo = repo.NewTaskRepository(db)

		layout = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(makeTitleBar(), 2, 1, false).
			AddItem(prepareContentPages(), 0, 2, true).
			AddItem(prepareStatusBar(app), 1, 1, false)

		setKeyboardShortcuts()

		if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}

}

func migrate(database *storm.DB) {
	util.FatalIfError(database.ReIndex(&model.Project{}), "Error in migrating Projects")
	util.FatalIfError(database.ReIndex(&model.Task{}), "Error in migrating Tasks")

	fmt.Println("Migration completed. Start geek-life normally.")
	os.Exit(0)
}

func setKeyboardShortcuts() *tview.Application {
	return app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ignoreKeyEvt() {
			return event
		}

		// Global shortcuts
		switch event.Rune() {
		case '/':
			ShowSearchModal()
			return nil
		}
		
		switch unicode.ToLower(event.Rune()) {
		case 'p':
			app.SetFocus(projectPane)
			contents.RemoveItem(taskDetailPane)
			return nil
		case 'q':
			app.Stop()
			return nil
		case 't':
			app.SetFocus(taskPane)
			contents.RemoveItem(taskDetailPane)
			return nil
		}

		// Handle based on current focus. Handlers may modify event
		switch {
		case projectPane.HasFocus():
			event = projectPane.handleShortcuts(event)
		case taskPane.HasFocus():
			event = taskPane.handleShortcuts(event)
			if event != nil && projectDetailPane.isShowing() {
				event = projectDetailPane.handleShortcuts(event)
			}
		case taskDetailPane.HasFocus():
			event = taskDetailPane.handleShortcuts(event)
		}

		return event
	})
}

func prepareContentPages() *tview.Flex {
	projectPane = NewProjectPane(projectRepo)
	taskPane = NewTaskPane(projectRepo, taskRepo)
	projectDetailPane = NewProjectDetailPane()
	taskDetailPane = NewTaskDetailPane(taskRepo)

	contents = tview.NewFlex().
		AddItem(projectPane, 0, 1, true).
		AddItem(taskPane, 0, 4, false)

	return contents

}

func makeTitleBar() *tview.Flex {
	titleText := tview.NewTextView().
		SetText("[lime::b]Geek-life [::-]- Task Manager for geeks!").
		SetDynamicColors(true)
	versionInfo := tview.NewTextView().
		SetText("[::d]Version: 0.1.2").
		SetTextAlign(tview.AlignRight).
		SetDynamicColors(true)

	return tview.NewFlex().
		AddItem(titleText, 0, 2, false).
		AddItem(versionInfo, 0, 1, false)
}

func AskYesNo(text string, f func()) {

	activePane := app.GetFocus()
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(
			func(
				_ int,
				buttonLabel string,
			) {
				if buttonLabel == "Yes" {
					f()
				}
				app.SetRoot(layout, true).EnableMouse(true)
				app.SetFocus(activePane)
			})

	pages := tview.NewPages().
		AddPage("background", layout, true, true).
		AddPage("modal", modal, true, true)
	_ = app.SetRoot(pages, true).EnableMouse(true)
}

func ShowSearchModal() {
	activePane := app.GetFocus()
	
	// Create search input field
	searchInput := tview.NewInputField().
		SetLabel("Search: ").
		SetPlaceholder("Enter search query...").
		SetFieldWidth(40)
	
	// Create results list
	resultsList := tview.NewList().ShowSecondaryText(false)
	resultsList.SetBorder(true).SetTitle("Search Results")
	
	// Track current search results
	var currentResults []SearchResult
	
	// Search function
	performSearch := func(query string) {
		if len(strings.TrimSpace(query)) < 2 {
			resultsList.Clear()
			currentResults = nil
			resultsList.AddItem("Type at least 2 characters to search", "", 0, nil)
			return
		}
		
		// Clear previous results
		resultsList.Clear()
		currentResults = nil
		
		// Search tasks
		tasks, err := taskRepo.SearchTasks(query)
		if err == nil {
			for _, task := range tasks {
				// Get project name for context
				project, err := projectRepo.GetByID(task.ProjectID)
				projectName := "Unknown Project"
				if err == nil {
					projectName = project.Title
				}
				
				result := SearchResult{
					Type:        "task",
					Title:       fmt.Sprintf("%s: %s", projectName, task.Title),
					TaskID:      task.ID,
					ProjectID:   task.ProjectID,
				}
				currentResults = append(currentResults, result)
			}
		}
		
		// Search projects
		projects, err := projectRepo.SearchProjects(query)
		if err == nil {
			for _, project := range projects {
				result := SearchResult{
					Type:      "project",
					Title:     fmt.Sprintf("[Project] %s", project.Title),
					ProjectID: project.ID,
				}
				currentResults = append(currentResults, result)
			}
		}
		
		// Display results
		if len(currentResults) == 0 {
			resultsList.AddItem("No results found", "", 0, nil)
		} else {
			for i, result := range currentResults {
				resultsList.AddItem(result.Title, "", rune('1'+i), func(idx int) func() {
					return func() { selectSearchResult(idx, currentResults, activePane) }
				}(i))
			}
		}
	}
	
	// Set up input field behavior
	searchInput.SetChangedFunc(performSearch)
	searchInput.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			// Move focus to results list
			app.SetFocus(resultsList)
		case tcell.KeyEsc:
			// Close modal
			app.SetRoot(layout, true).EnableMouse(true)
			app.SetFocus(activePane)
		case tcell.KeyTab:
			// Move focus to results list
			app.SetFocus(resultsList)
		}
	})
	
	// Set up results list behavior
	resultsList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			// Close modal
			app.SetRoot(layout, true).EnableMouse(true)
			app.SetFocus(activePane)
			return nil
		case tcell.KeyTab:
			// Move focus back to search input
			app.SetFocus(searchInput)
			return nil
		}
		return event
	})
	
	// Create modal layout
	modalFlex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(searchInput, 3, 0, true).
		AddItem(resultsList, 0, 1, false)
	
	modalFlex.SetBorder(true).SetTitle("Search Tasks and Projects (ESC to close, Tab to switch)")
	
	// Show modal
	pages := tview.NewPages().
		AddPage("background", layout, true, true).
		AddPage("search", modalFlex, true, true)
	_ = app.SetRoot(pages, true).EnableMouse(true)
	app.SetFocus(searchInput)
}

type SearchResult struct {
	Type      string // "task" or "project"
	Title     string
	TaskID    int64
	ProjectID int64
}

func selectSearchResult(index int, results []SearchResult, activePane tview.Primitive) {
	if index >= len(results) {
		return
	}
	
	result := results[index]
	
	// Close modal first
	app.SetRoot(layout, true).EnableMouse(true)
	
	if result.Type == "task" {
		// Load the project and then activate the task
		project, err := projectRepo.GetByID(result.ProjectID)
		if err == nil {
			// Set active project
			projectPane.activeProject = &project
			
			// Load project tasks
			taskPane.LoadProjectTasks(project)
			
			// Find and activate the specific task
			for i, task := range taskPane.tasks {
				if task.ID == result.TaskID {
					taskPane.ActivateTask(i)
					app.SetFocus(taskDetailPane)
					break
				}
			}
		}
	} else if result.Type == "project" {
		// Find and activate the project
		for i, project := range projectPane.projects {
			if project.ID == result.ProjectID {
				projectPane.activateProject(i)
				app.SetFocus(taskPane)
				break
			}
		}
	}
}
