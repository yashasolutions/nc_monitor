package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

// Config holds all configuration values
type Config struct {
	WatchDir       string
	RemoteDir      string
	NextcloudURL   string
	Username       string
	Password       string
	LogFile        string
	PidFile        string
	SyncCooldown   time.Duration
	SyncInterval   string
	IgnorePatterns []string
	Verbose        bool
}

var (
	cfg *Config
)

func init() {
	cfg = initConfig()
}

func initConfig() *Config {
	// Set up Viper
	viper.SetConfigName("nextcloud_monitor")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/nextcloud_monitor")
	viper.AddConfigPath("/etc/nextcloud_monitor")

	// Set environment variable prefix
	viper.SetEnvPrefix("NEXTCLOUD")
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("watch_dir", "~/org/roam")
	viper.SetDefault("remote_dir", "/org/roam")
	viper.SetDefault("url", "https://nextcloud.example.com")
	viper.SetDefault("user", "nextcloud_user")
	viper.SetDefault("password", "nextcloud_pass")
	viper.SetDefault("log_file", "/tmp/nextcloud_monitor.log")
	viper.SetDefault("pid_file", "/tmp/nextcloud_monitor.pid")
	viper.SetDefault("sync_cooldown", "10s")
	viper.SetDefault("sync_interval", "*/5 * * * *")
	viper.SetDefault("verbose", false)
	viper.SetDefault("ignore_patterns", []string{
		"*.tmp", "*.temp", "*.log", "*~", ".DS_Store",
		"Thumbs.db", ".git/*", "*.swp", "*.lock", ".nextcloud_sync_*",
	})

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Error reading config file: %v", err)
		}
	}

	// Parse sync cooldown duration
	cooldownStr := viper.GetString("sync_cooldown")
	cooldown, err := time.ParseDuration(cooldownStr)
	if err != nil {
		log.Printf("Invalid sync_cooldown format '%s', using default 10s", cooldownStr)
		cooldown = 10 * time.Second
	}

	return &Config{
		WatchDir:       viper.GetString("watch_dir"),
		RemoteDir:      viper.GetString("remote_dir"),
		NextcloudURL:   viper.GetString("url"),
		Username:       viper.GetString("user"),
		Password:       viper.GetString("password"),
		LogFile:        viper.GetString("log_file"),
		PidFile:        viper.GetString("pid_file"),
		SyncCooldown:   cooldown,
		SyncInterval:   viper.GetString("sync_interval"),
		Verbose:        viper.GetBool("verbose"),
		IgnorePatterns: viper.GetStringSlice("ignore_patterns"),
	}
}

type Monitor struct {
	watcher    *fsnotify.Watcher
	lastSync   time.Time
	logger     *log.Logger
	cron       *cron.Cron
}

func main() {
	// Setup logging
	logFile, err := os.OpenFile(cfg.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	var logWriter io.Writer = logFile
	if cfg.Verbose {
		logWriter = io.MultiWriter(logFile, os.Stdout)
	}

	logger := log.New(logWriter, "", log.LstdFlags|log.Lshortfile)

	// Check for existing instance
	if err := checkInstance(); err != nil {
		logger.Fatal(err)
	}
	defer os.Remove(cfg.PidFile)

	// Create monitor
	monitor := &Monitor{
		logger: logger,
	}

	// Initialize watcher
	if err := monitor.initWatcher(); err != nil {
		logger.Fatal(err)
	}
	defer monitor.watcher.Close()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		monitor.cleanup()
		os.Exit(0)
	}()

	// Initial sync
	if err := monitor.sync(); err != nil {
		logger.Println("Initial sync failed:", err)
	}

	// Start periodic sync
	monitor.startPeriodicSync()

	// Start watching for changes
	monitor.watch()
}

func (m *Monitor) initWatcher() error {
	if cfg.Verbose {
		m.logger.Println("Initializing file watcher...")
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %v", err)
	}
	m.watcher = watcher

	// Expand home directory path
	dir := os.ExpandEnv(cfg.WatchDir)
	if cfg.Verbose {
		m.logger.Printf("Watching directory: %s", dir)
	}

	// Walk through directory tree and add watches
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && !shouldIgnore(path) {
			return watcher.Add(path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %v", err)
	}

	return nil
}

func (m *Monitor) startPeriodicSync() {
	if cfg.Verbose {
		m.logger.Printf("Setting up periodic sync with interval: %s", cfg.SyncInterval)
	}
	m.cron = cron.New()
	_, err := m.cron.AddFunc(cfg.SyncInterval, func() {
		if err := m.sync(); err != nil {
			m.logger.Println("Periodic sync failed:", err)
		}
	})
	if err != nil {
		m.logger.Println("Failed to start periodic sync:", err)
		return
	}
	m.cron.Start()
}

func (m *Monitor) watch() {
	if cfg.Verbose {
		m.logger.Println("Starting file watcher...")
	}
	for {
		select {
		case event, ok := <-m.watcher.Events:
			if !ok {
				return
			}

			if shouldIgnore(event.Name) {
				m.logger.Printf("Ignoring event: %s", event)
				continue
			}

			// Check cooldown
			if time.Since(m.lastSync) < cfg.SyncCooldown {
				m.logger.Println("Sync cooldown active, skipping...")
				continue
			}

			m.logger.Printf("Detected change: %s", event)

			if err := m.sync(); err != nil {
				m.logger.Println("Sync failed:", err)
			}

		case err, ok := <-m.watcher.Errors:
			if !ok {
				return
			}
			m.logger.Println("Watcher error:", err)
		}
	}
}

func (m *Monitor) sync() error {
	if cfg.Verbose {
		m.logger.Println("Starting Nextcloud sync...")
		m.logger.Printf("Sync parameters - RemoteDir: %s, WatchDir: %s, URL: %s", 
			cfg.RemoteDir, cfg.WatchDir, cfg.NextcloudURL)
	}

	// Get password from pass command
	password := os.ExpandEnv(cfg.Password)

	// Build sync command
	cmd := exec.Command("nextcloudcmd",
		"--path", cfg.RemoteDir,
		os.ExpandEnv(cfg.WatchDir),
		fmt.Sprintf("https://%s:%s@%s", cfg.Username, password, cfg.NextcloudURL),
	)

	// Log command if verbose
	if cfg.Verbose {
		m.logger.Printf("Sync command: %v", cmd.Args)
	}

	// Run command
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("sync failed: %v", err)
	}

	m.lastSync = time.Now()
	m.logger.Println("Sync completed successfully")
	notify("Nextcloud Sync Done")
	return nil
}

func (m *Monitor) cleanup() {
	if m.cron != nil {
		m.cron.Stop()
	}
	m.logger.Println("Stopping monitor...")
}

func checkInstance() error {
	pidData, err := os.ReadFile(cfg.PidFile)
	if err == nil {
		pid := strings.TrimSpace(string(pidData))
		if _, err := os.Stat(fmt.Sprintf("/proc/%s", pid)); err == nil {
			return fmt.Errorf("another instance is already running (PID: %s)", pid)
		}
	}

	return os.WriteFile(cfg.PidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func shouldIgnore(path string) bool {
	for _, pattern := range cfg.IgnorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}

func notify(message string) {
	if os.Getenv("DISPLAY") != "" {
		cmd := exec.Command("dunstify", message)
		_ = cmd.Run()
	}
}
