// Package main provides the CLI entry point for ActionHero
package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/fatih/color"
)

const (
	formatJSON = "json"
	formatList = "list"
)

// dumpConfig displays the current configuration in a formatted way
func dumpConfig(cfg *config.Config, logger *util.Logger, format string) {
	// Validate format
	if format != formatList && format != formatJSON {
		logger.Errorf("  Invalid format '%s'. Use 'list' or 'json'", format)
		return
	}

	switch format {
	case formatJSON:
		dumpConfigJSON(cfg, logger)
	default:
		dumpConfigList(cfg, logger)
	}
}

// dumpConfigJSON displays the configuration as JSON
func dumpConfigJSON(cfg *config.Config, logger *util.Logger) {
	// Create a safe copy for JSON output (mask passwords)
	jsonCfg := struct {
		Process  config.ProcessConfig  `json:"process"`
		Logger   config.LoggerConfig   `json:"logger"`
		Database config.DatabaseConfig `json:"database"`
		Redis    config.RedisConfig    `json:"redis"`
		Session  config.SessionConfig  `json:"session"`
		Server   config.ServerConfig   `json:"server"`
		Tasks    config.TasksConfig    `json:"tasks"`
	}{
		Process:  cfg.Process,
		Logger:   cfg.Logger,
		Database: cfg.Database,
		Redis:    cfg.Redis,
		Session:  cfg.Session,
		Server:   cfg.Server,
		Tasks:    cfg.Tasks,
	}

	// Mask passwords
	if cfg.Database.Password != "" {
		jsonCfg.Database.Password = maskPassword(cfg.Database.Password)
	} else {
		jsonCfg.Database.Password = ""
	}
	if cfg.Redis.Password != "" {
		jsonCfg.Redis.Password = maskPassword(cfg.Redis.Password)
	} else {
		jsonCfg.Redis.Password = ""
	}

	jsonData, err := json.MarshalIndent(jsonCfg, "", "  ")
	if err != nil {
		logger.Errorf("Failed to marshal config to JSON: %v", err)
		return
	}

	// For JSON format, output directly to stdout without logger formatting
	fmt.Println(string(jsonData))
}

// dumpConfigList displays the configuration in a formatted list (original format)
func dumpConfigList(cfg *config.Config, logger *util.Logger) {
	// Create color printers
	sectionColor := color.New(color.FgCyan, color.Bold)
	keyColor := color.New(color.FgYellow)
	valueColor := color.New(color.FgWhite)

	// Helper function to print key-value pairs
	printKV := func(key, value string) {
		logger.Info(fmt.Sprintf("  %s: %s", keyColor.Sprint(key), valueColor.Sprint(value)))
	}

	// Helper function to print section header
	printSection := func(title string) {
		logger.Info("")
		logger.Info(sectionColor.Sprint("  " + strings.ToUpper(title)))
		logger.Info(sectionColor.Sprint("  " + strings.Repeat("─", len(title)+2)))
	}

	// Process
	printSection("Process")
	printKV("Name", cfg.Process.Name)

	// Logger
	printSection("Logger")
	printKV("Level", cfg.Logger.Level)
	printKV("Colorize", fmt.Sprintf("%v", cfg.Logger.Colorize))
	printKV("Timestamp", fmt.Sprintf("%v", cfg.Logger.Timestamp))

	// Database
	printSection("Database")
	printKV("Type", cfg.Database.Type)
	printKV("Host", cfg.Database.Host)
	printKV("Port", fmt.Sprintf("%d", cfg.Database.Port))
	printKV("User", cfg.Database.User)
	printKV("Password", maskPassword(cfg.Database.Password))
	printKV("Database", cfg.Database.Database)
	printKV("SSL Mode", cfg.Database.SSLMode)

	// Redis
	printSection("Redis")
	printKV("Host", cfg.Redis.Host)
	printKV("Port", fmt.Sprintf("%d", cfg.Redis.Port))
	printKV("Password", maskPassword(cfg.Redis.Password))
	printKV("DB", fmt.Sprintf("%d", cfg.Redis.DB))

	// Session
	printSection("Session")
	printKV("Cookie Name", cfg.Session.CookieName)
	printKV("TTL", fmt.Sprintf("%d seconds", cfg.Session.TTL))

	// Server
	printSection("Server - Web")
	printKV("Enabled", fmt.Sprintf("%v", cfg.Server.Web.Enabled))
	printKV("Host", cfg.Server.Web.Host)
	printKV("Port", fmt.Sprintf("%d", cfg.Server.Web.Port))
	printKV("API Route", cfg.Server.Web.APIRoute)
	printKV("Allowed Origins", cfg.Server.Web.AllowedOrigins)
	printKV("Allowed Methods", cfg.Server.Web.AllowedMethods)
	printKV("Allowed Headers", cfg.Server.Web.AllowedHeaders)
	printKV("Static Files Enabled", fmt.Sprintf("%v", cfg.Server.Web.StaticFilesEnabled))
	if cfg.Server.Web.StaticFilesEnabled {
		printKV("Static Files Route", cfg.Server.Web.StaticFilesRoute)
		printKV("Static Files Directory", cfg.Server.Web.StaticFilesDirectory)
	}

	// Tasks
	printSection("Tasks")
	printKV("Enabled", fmt.Sprintf("%v", cfg.Tasks.Enabled))
	if cfg.Tasks.Enabled {
		printKV("Task Processors", fmt.Sprintf("%d", cfg.Tasks.TaskProcessors))
		printKV("Queues", fmt.Sprintf("%v", cfg.Tasks.Queues))
		printKV("Timeout", fmt.Sprintf("%d ms", cfg.Tasks.Timeout))
		printKV("Stuck Worker Timeout", fmt.Sprintf("%d ms", cfg.Tasks.StuckWorkerTimeout))
		printKV("Retry Stuck Jobs", fmt.Sprintf("%v", cfg.Tasks.RetryStuckJobs))
	}

	logger.Info("")
}

// maskPassword masks sensitive password values
func maskPassword(password string) string {
	if password == "" {
		return "(empty)"
	}
	return strings.Repeat("*", len(password))
}
