package ticketmanager

type TicketManager interface {
	// Epic/Project management
	CreateEpic(title, description string) (string, error)
	UpdateEpic(title, description string, epicID string) (string, error)
	ListEpics() ([]Epic, error)
	ListUserEpics() ([]Epic, error)
	DescribeEpic(epicID string) (*Epic, error)

	// Task management
	CreateTask(title, description string, epicID string) (string, error)
	UpdateTask(title, description string, completed bool, taskID string) error
	ListTasksForEpic(epicID string) ([]Task, error)
	DescribeTask(taskID string) (*Task, error)
}

type Epic struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Creator     User   `json:"creator"`
}

type Task struct {
	ID          string `json:"id"`
	Key         string `json:"key"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Completed   bool   `json:"completed"`
	EpicID      string `json:"epicId"`
	Creator     User   `json:"creator"`
}

type User struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}
