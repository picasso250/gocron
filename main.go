package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Task struct {
	Name     string   `json:"name"`
	Command  string   `json:"command"`
	Args     []string `json:"args"`
	Schedule string   `json:"schedule"` // "hourly" or "daily"
	Hour     *int     `json:"hour,omitempty"`
	Enabled  bool     `json:"enabled"`
}

const (
	configFile = "cron.json"
	logDir     = "log"
	logFile    = "log/cron.log"
	interval   = 1 * time.Hour
)

func main() {
	// Setup logging
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Error creating log directory: %v\n", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(os.Stdout, f)
	logger := log.New(multiWriter, "", 0)

	logger.Printf("[%s] [INFO] gocron Service Started. Interval: %v", time.Now().Format("2006-01-02 15:04:05"), interval)

	for {
		runTasks(logger)
		logger.Printf("[%s] [INFO] Sleeping for %v. Next run at %s", 
			time.Now().Format("2006-01-02 15:04:05"), 
			interval, 
			time.Now().Add(interval).Format("15:04:05"))
		time.Sleep(interval)
	}
}

func runTasks(logger *log.Logger) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		logger.Printf("[%s] [ERROR] Error reading config file: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		logger.Printf("[%s] [ERROR] Error parsing config file: %v", time.Now().Format("2006-01-02 15:04:05"), err)
		return
	}

	now := time.Now()
	currentHour := now.Hour()
	logger.Printf("[%s] [INFO] Running scheduled tasks...", now.Format("2006-01-02 15:04:05"))

	for _, task := range tasks {
		if !task.Enabled {
			logger.Printf("[%s] [SKIP] %s (disabled)", time.Now().Format("2006-01-02 15:04:05"), task.Name)
			continue
		}

		isHourly := task.Schedule == "" || task.Schedule == "hourly"
		isDailyMatch := task.Schedule == "daily" && task.Hour != nil && *task.Hour == currentHour

		if isHourly || isDailyMatch {
			logger.Printf("[%s] [EXEC] %s (%s %v)", time.Now().Format("2006-01-02 15:04:05"), task.Name, task.Command, task.Args)
			
			// Handle python execution on windows if needed
			cmdName := task.Command
			if runtime.GOOS == "windows" && cmdName == "python" {
				// Check if 'python' exists, if not try 'py' or 'python3'
				if _, err := exec.LookPath("python"); err != nil {
					if _, err := exec.LookPath("python3"); err == nil {
						cmdName = "python3"
					} else if _, err := exec.LookPath("py"); err == nil {
						cmdName = "py"
					}
				}
			}

			cmd := exec.Command(cmdName, task.Args...)
			// Run in the parent directory of gocron to match script behavior
			cmd.Dir = ".." 
			
			output, err := cmd.CombinedOutput()
			if err != nil {
				logger.Printf("[%s] [FAIL] %s failed: %v", time.Now().Format("2006-01-02 15:04:05"), task.Name, err)
				if len(output) > 0 {
					logger.Printf("Output: %s", string(output))
				}
			}
		} else {
			hourStr := "N/A"
			if task.Hour != nil {
				hourStr = fmt.Sprintf("%d", *task.Hour)
			}
			logger.Printf("[%s] [SKIP] %s (Scheduled for %s at hour %s)", time.Now().Format("2006-01-02 15:04:05"), task.Name, task.Schedule, hourStr)
		}
	}
}
