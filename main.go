package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
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

func main() {
	configPath := flag.String("config", "cron.json", "Path to the cron configuration file")
	logPath := flag.String("log", "log/cron.log", "Path to the log file")
	interval := flag.Duration("interval", 1*time.Hour, "Interval between task runs")
	flag.Parse()

	// Ensure log directory exists
	logDir := filepath.Dir(*logPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Error creating log directory: %v\n", err)
		os.Exit(1)
	}

	f, err := os.OpenFile(*logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	multiWriter := io.MultiWriter(os.Stdout, f)
	logger := log.New(multiWriter, "", 0)

	logger.Printf("[%s] [INFO] gocron Service Started. Interval: %v", time.Now().Format("2006-01-02 15:04:05"), *interval)
	logger.Printf("[%s] [INFO] Config: %s, Log: %s", time.Now().Format("2006-01-02 15:04:05"), *configPath, *logPath)

	for {
		runTasks(logger, *configPath)
		logger.Printf("[%s] [INFO] Sleeping for %v. Next run at %s", 
			time.Now().Format("2006-01-02 15:04:05"), 
			*interval, 
			time.Now().Add(*interval).Format("15:04:05"))
		time.Sleep(*interval)
	}
}

func runTasks(logger *log.Logger, configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		logger.Printf("[%s] [ERROR] Error reading config file (%s): %v", time.Now().Format("2006-01-02 15:04:05"), configPath, err)
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

	// Get the directory of the config file to use as the base for relative paths if needed
	// But as per original behavior, we run from the current working directory of gocron
	// We'll stick to running from the directory where gocron is executed (usually the root)
	
	for _, task := range tasks {
		if !task.Enabled {
			logger.Printf("[%s] [SKIP] %s (disabled)", time.Now().Format("2006-01-02 15:04:05"), task.Name)
			continue
		}

		isHourly := task.Schedule == "" || task.Schedule == "hourly"
		isDailyMatch := task.Schedule == "daily" && task.Hour != nil && *task.Hour == currentHour

		if isHourly || isDailyMatch {
			logger.Printf("[%s] [EXEC] %s (%s %v)", time.Now().Format("2006-01-02 15:04:05"), task.Name, task.Command, task.Args)
			
			cmdName := task.Command
			if runtime.GOOS == "windows" && (cmdName == "python" || cmdName == "python3") {
				if _, err := exec.LookPath(cmdName); err != nil {
					if _, err := exec.LookPath("py"); err == nil {
						cmdName = "py"
					}
				}
			}

			cmd := exec.Command(cmdName, task.Args...)
			// Run from the same directory as gocron (root)
			
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
