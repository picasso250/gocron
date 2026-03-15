# gocron

A lightweight task scheduler written in Go, migrated from `cron.ps1`. It executes tasks defined in a JSON configuration file at regular intervals.

## Features

- **JSON Configuration**: Define tasks, commands, arguments, and schedules in `cron.json`.
- **Flexible Scheduling**: Supports `hourly` and `daily` (at a specific hour) tasks.
- **Logging**: Dual-output logging (stdout and a file) with custom log path support.
- **Cross-Platform**: Built for Windows but compatible with Go-supported OSs.
- **Python Support**: Automatically detects `python`, `python3`, or `py` on Windows.

## Usage

### Command Line Arguments

- `-config`: Path to the JSON configuration file (default: `cron.json`).
- `-log`: Path to the log file (default: `log/cron.log`).
- `-interval`: Duration between task checks (default: `1h`).

### Running

```powershell
.\gocron.exe -config ..\cron.json -log ..\log\cron.log
```

### Configuration Format (`cron.json`)

```json
[
    {
        "name": "Market Monitor",
        "command": "python",
        "args": ["finance/market_monitor.py"],
        "schedule": "hourly",
        "enabled": true
    },
    {
        "name": "Daily Report",
        "command": "python",
        "args": ["scripts/report.py"],
        "schedule": "daily",
        "hour": 8,
        "enabled": true
    }
]
```

## Deployment

This tool is used in the `g-claw` workspace and is typically started via `init.ps1`.

```powershell
Start-Process .\gocron\gocron.exe -ArgumentList "-config cron.json -log log/cron.log" -WindowStyle Hidden
```

## Building

```bash
go build -o gocron.exe main.go
```
