# Nextcloud Monitor

A lightweight Go application that monitors a local directory for file changes and automatically synchronizes them with Nextcloud using the `nextcloudcmd` client.

## Overview

This tool provides real-time file synchronization between a local directory and your Nextcloud instance. It watches for file system changes and triggers syncs automatically, while also performing periodic syncs to ensure consistency.

## Features

- **Real-time Monitoring**: Uses `fsnotify` to watch for file changes in real-time
- **Automatic Synchronization**: Triggers sync operations when files are modified
- **Periodic Sync**: Configurable cron-based periodic synchronization (default: every 5 minutes)
- **Cooldown Protection**: Prevents excessive sync operations with a 10-second cooldown
- **Instance Management**: Prevents multiple instances from running simultaneously using PID files
- **Comprehensive Logging**: Detailed logging with configurable log file location
- **Desktop Notifications**: Optional desktop notifications via `dunstify`
- **Smart Ignore Patterns**: Automatically ignores temporary files, system files, and common artifacts
- **Verbose Mode**: Optional detailed logging for debugging

## Prerequisites

- Go 1.24.3 or later
- `nextcloudcmd` client installed and accessible in PATH
- Access to a Nextcloud instance
- `dunstify` (optional, for desktop notifications)

### Installing nextcloudcmd

**Ubuntu/Debian:**
```bash
sudo apt install nextcloud-desktop-cmd
```

**Arch Linux:**
```bash
sudo pacman -S nextcloud-client
```

**macOS:**
```bash
brew install nextcloud
```

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd nc_monitor
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the application:
```bash
go build -o nextcloud_monitor nextcloud_monitor.go
```

## Configuration

The application is configured entirely through environment variables. Create a `.env` file or set these variables in your shell:

### Required Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXTCLOUD_WATCH_DIR` | `~/org/roam` | Local directory to monitor |
| `NEXTCLOUD_REMOTE_DIR` | `/org/roam` | Remote Nextcloud directory path |
| `NEXTCLOUD_URL` | `https://nextcloud.example.com` | Your Nextcloud server URL |
| `NEXTCLOUD_USER` | `nextcloud_user` | Nextcloud username |
| `NEXTCLOUD_PASSWORD` | `nexcloud_pass` | Nextcloud password or app token |

### Optional Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXTCLOUD_LOG_FILE` | `/tmp/nextcloud_monitor.log` | Log file path |
| `NEXTCLOUD_PID_FILE` | `/tmp/nextcloud_monitor.pid` | PID file path |
| `NEXTCLOUD_SYNC_INTERVAL` | `*/5 * * * *` | Cron expression for periodic sync |
| `NEXTCLOUD_VERBOSE` | `false` | Enable verbose logging |

### Example .env file

```bash
NEXTCLOUD_URL=https://your-nextcloud.example.com
NEXTCLOUD_USER=your-username
NEXTCLOUD_PASSWORD=your-app-password
NEXTCLOUD_WATCH_DIR=/home/user/Documents/sync
NEXTCLOUD_REMOTE_DIR=/Documents/sync
NEXTCLOUD_VERBOSE=true
```

## Usage

1. Set up your environment variables (create a `.env` file or export them)

2. Run the monitor:
```bash
./nextcloud_monitor
```

The application will:
- Perform an initial sync
- Start monitoring the specified directory for changes
- Sync changes automatically when detected
- Run periodic syncs based on the configured interval
- Log all activities to the specified log file

3. To stop the monitor, use `Ctrl+C` or send a SIGTERM signal

## Security Considerations

- **Use App Passwords**: Instead of your main Nextcloud password, create an app-specific password in your Nextcloud settings
- **Environment Variables**: Store sensitive credentials in environment variables or a `.env` file (which is gitignored)
- **File Permissions**: Ensure proper permissions on log and PID files
- **Password Managers**: The password field supports environment variable expansion for integration with password managers

## Ignored Files

The monitor automatically ignores common temporary and system files:
- `*.tmp`, `*.temp`, `*.log`
- Backup files (`*~`)
- System files (`.DS_Store`, `Thumbs.db`)
- Git directories (`.git/*`)
- Editor files (`*.swp`, `*.lock`)
- Nextcloud sync artifacts (`.nextcloud_sync_*`)

## Logging

All activities are logged to the specified log file with timestamps and source file information. Enable verbose mode for detailed debugging information.

## Troubleshooting

### Common Issues

1. **"nextcloudcmd not found"**: Install the Nextcloud desktop client
2. **Permission denied**: Check file permissions on watch directory and log files
3. **Sync failures**: Verify Nextcloud URL, credentials, and network connectivity
4. **Multiple instances**: Check for existing PID file and running processes

### Debug Mode

Enable verbose logging to see detailed information:
```bash
export NEXTCLOUD_VERBOSE=true
./nextcloud_monitor
```

### Log Analysis

Check the log file for detailed error messages:
```bash
tail -f /tmp/nextcloud_monitor.log
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Development

### Project Structure

- `nextcloud_monitor.go` - Main application code
- `go.mod` - Go module definition and dependencies
- `.gitignore` - Git ignore patterns
- `README.md` - This documentation

### Dependencies

- `github.com/fsnotify/fsnotify` - File system event notifications
- `github.com/robfig/cron/v3` - Cron-based job scheduling

### Building

```bash
go build -o nextcloud_monitor nextcloud_monitor.go
```

### Testing

Run the application in verbose mode with a test directory to verify functionality.

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## Support

For issues and questions:
1. Check the troubleshooting section above
2. Review the log files for error details
3. Open an issue on the project repository
