package util

import (
	"io"
	"log"
	"log/syslog"
	"os"
	"path/filepath"
	"runtime"
)

var (
	Logger      *log.Logger
	logWriter   io.Writer
	initialized bool
)

// InitLogger initializes the logging system
// On Linux systems, tries syslog first, then falls back to file logging
// On other systems, uses file logging directly
func InitLogger() error {
	if initialized {
		return nil
	}

	var err error

	// Try syslog first on Linux systems
	if runtime.GOOS == "linux" {
		logWriter, err = syslog.New(syslog.LOG_INFO|syslog.LOG_USER, "geek-life")
		if err == nil {
			Logger = log.New(logWriter, "", log.LstdFlags)
			initialized = true
			Logger.Println("Logging initialized with syslog")
			return nil
		}
		// If syslog fails, fall through to file logging
	}

	// Fallback to file logging
	logDir := getLogDirectory()
	CreateDirIfNotExist(logDir)

	logFile := filepath.Join(logDir, "application.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// Last resort: use stderr
		Logger = log.New(os.Stderr, "[geek-life] ", log.LstdFlags)
		initialized = true
		Logger.Println("Logging initialized with stderr fallback")
		return nil
	}

	Logger = log.New(file, "", log.LstdFlags)
	initialized = true
	Logger.Printf("Logging initialized with file: %s", logFile)
	return nil
}

// getLogDirectory returns the appropriate log directory based on the system
func getLogDirectory() string {
	switch runtime.GOOS {
	case "linux", "darwin":
		// Check if /var/log/messages is writable (for system-wide logging)
		if isWritable("/var/log") {
			return "/var/log"
		}
		// Fall back to user's local directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, ".local", "geek-life")
		}
		// Ultimate fallback
		return "/tmp"
	case "windows":
		homeDir, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(homeDir, "AppData", "Local", "geek-life")
		}
		return os.TempDir()
	default:
		return os.TempDir()
	}
}

// isWritable checks if a directory is writable
func isWritable(dir string) bool {
	testFile := filepath.Join(dir, ".geek-life-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}

// LogInfo logs an info message
func LogInfo(msg string, args ...interface{}) {
	if !initialized {
		InitLogger()
	}
	Logger.Printf("[INFO] "+msg, args...)
}

// LogError logs an error message
func LogError(msg string, args ...interface{}) {
	if !initialized {
		InitLogger()
	}
	Logger.Printf("[ERROR] "+msg, args...)
}

// LogWarning logs a warning message
func LogWarning(msg string, args ...interface{}) {
	if !initialized {
		InitLogger()
	}
	Logger.Printf("[WARNING] "+msg, args...)
}

// LogDebug logs a debug message
func LogDebug(msg string, args ...interface{}) {
	if !initialized {
		InitLogger()
	}
	Logger.Printf("[DEBUG] "+msg, args...)
}

// GetLogPath returns the current log file path (for file-based logging)
func GetLogPath() string {
	if runtime.GOOS == "linux" {
		// Check if we're using syslog
		if _, ok := logWriter.(*syslog.Writer); ok {
			return "/var/log/messages (via syslog)"
		}
	}

	logDir := getLogDirectory()
	return filepath.Join(logDir, "application.log")
}
