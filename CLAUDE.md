# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Geek-life is a CLI-based task manager/to-do list application built in Go, designed specifically for developers and terminal users. It's a terminal UI (TUI) application built with tview for interactive widgets and uses BoltDB (via Storm) for local data storage.

## Build and Development Commands

### Building the Application
- `go build -o geek-life ./app` - Build for current platform
- `./build.sh` - Cross-compile for multiple platforms (Darwin, Linux, Windows)
- The build script creates optimized binaries in the `builds/` directory with UPX compression

### Running the Application
- `go run ./app` - Run directly from source
- `./geek-life` - Run compiled binary
- `./geek-life --db-file=/path/to/custom.db` - Specify custom database file location
- `./geek-life migrate` - Run database migration

### Testing
No formal test suite is present in the codebase. Manual testing should be done via running the application.

## Architecture Overview

### Core Components
- **app/cli.go** - Main entry point and application setup
- **model/** - Data models (Project, Task) with Storm ORM tags
- **repository/** - Data access layer with interfaces and Storm implementation
- **jira/** - JIRA integration functionality
- **api/** - API layer (minimal implementation)
- **util/** - Utility functions including database connection

### Key Architecture Patterns
- Repository pattern for data access with interface-based design
- Storm ORM for BoltDB integration with struct tags for indexing
- TUI built with tview framework using panes and keyboard shortcuts
- Flexible layout system with three main panes: Projects, Tasks, Details

### Data Storage
- Uses BoltDB via Storm ORM for local file-based storage
- Default location: user's home directory
- Configurable via `DB_FILE` environment variable or `--db-file` flag
- Models use Storm tags for indexing and relationships

### UI Structure
The application has a three-pane layout:
1. **Projects Pane** (left) - List of projects with JIRA integration indicators
2. **Tasks Pane** (center) - Tasks for selected project or dynamic lists (Today, Tomorrow, etc.)
3. **Detail Pane** (right) - Task/project details with markdown editor

### Key Dependencies
- `github.com/rivo/tview` - Terminal UI framework
- `github.com/asdine/storm/v3` - BoltDB ORM
- `github.com/pgavlin/femto` - Markdown editor with syntax highlighting
- `github.com/gdamore/tcell/v2` - Terminal cell manipulation
- `github.com/atotto/clipboard` - Clipboard operations

### Ticket Management Integration
The application supports both JIRA and Linear.app integration for importing epics as projects and syncing task status:

#### Configuration
Set the `TICKET_PROVIDER` environment variable to choose your provider:
- `TICKET_PROVIDER=jira` - Use JIRA (default)
- `TICKET_PROVIDER=linear` - Use Linear.app

**JIRA Configuration:**
Set these environment variables to enable JIRA integration:
- `JIRA_URL` - Your JIRA instance URL
- `JIRA_USERNAME` - Your JIRA username/email  
- `JIRA_API_TOKEN` - Your JIRA API token
- `JIRA_PROJECT_KEY` - The JIRA project key to work with

**Linear Configuration:**
Set these environment variables to enable Linear integration:
- `LINEAR_API_KEY` - Your Linear API key (get from https://linear.app/settings/api)
- `LINEAR_TEAM_KEY` - Your Linear team key (e.g., "ENG", "DESIGN")

#### Features
- **Import User Epics**: Press `Ctrl+I` in the Projects pane to import epics/projects created by the current user as projects
- **Create Epic**: Press `Ctrl+J` on a selected project to create a corresponding epic/project in your ticket system
- **Task Sync**: When marking tasks as complete/incomplete, changes are automatically synced to your ticket system
- **Epic-Task Mapping**: 
  - **JIRA**: Projects → Epics, Tasks → Issues under epics
  - **Linear**: Projects → Projects, Tasks → Issues within projects
- **User Filtering**: Only imports epics/projects where the creator matches the current user
- **Browser Integration**: Press `Ctrl+B` to open tickets directly in your browser

### Development Notes
- Go 1.19+ required
- No external runtime dependencies (single binary)
- Cross-platform support (Darwin, Linux, Windows)
- Keyboard-driven interface with vim-like navigation (h/j/k/l)
- Markdown support throughout the application
- JIRA integration is optional and only activates when properly configured