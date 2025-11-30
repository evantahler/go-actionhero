package main

import (
	"io"
	"os"

	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	// Global flags
	noColor     bool
	noTimestamp bool
	quiet       bool

	// Config and logger (set after LoadConfig)
	cfg    *config.Config
	logger *util.Logger
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "actionhero",
	Short: "Go ActionHero - A transport-agnostic API framework",
	Long: `Go ActionHero is a transport-agnostic API framework for building
scalable APIs with support for HTTP, WebSocket, and CLI transports.`,
	PersistentPreRunE: loadConfigAndInitLogger,
	Run: func(cmd *cobra.Command, args []string) {
		// Disable timestamps for help command
		disableTimestampsForCommand()
		showWelcome()
		cmd.Help()
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ActionHero server",
	Long:  `Start the ActionHero server and begin accepting connections.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("  Starting ActionHero server...")
		// TODO: Start the server
		logger.Warn("  Server start not yet implemented")
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display current configuration",
	Long:  `Display the current application configuration in a formatted way.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		// Disable timestamps for config command
		disableTimestampsForCommand()
		// Skip welcome message for JSON output
		format, _ := cmd.Flags().GetString("format")
		if format != "json" {
			showWelcome()
		}
	},
	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		dumpConfig(cfg, logger, format)
	},
}

func init() {
	// Global flags (persistent across all commands)
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
	rootCmd.PersistentFlags().BoolVar(&noTimestamp, "no-timestamp", false, "Disable timestamps in output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Quiet mode (hide logging output)")

	// Config command flags
	configCmd.Flags().String("format", "list", "Output format: list or json")

	// Add subcommands
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(configCmd)
}

// loadConfigAndInitLogger loads configuration and initializes the logger
// This runs before any command execution
func loadConfigAndInitLogger(cmd *cobra.Command, args []string) error {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		color.New(color.FgRed).Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Override config with CLI flags
	if noColor {
		cfg.Logger.Colorize = false
	}
	if noTimestamp {
		cfg.Logger.Timestamp = false
	}

	// Initialize logger
	logger = util.NewLogger(cfg.Logger)

	// Configure color library based on config
	if !cfg.Logger.Colorize {
		color.NoColor = true
	}

	// Quiet mode: redirect logger output to discard
	if quiet {
		logger.SetOutput(io.Discard)
	}

	return nil
}

// disableTimestampsForCommand disables timestamps in the logger for display commands
func disableTimestampsForCommand() {
	if logger != nil && !noTimestamp {
		logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp: true,
			ForceColors:      cfg.Logger.Colorize,
		})
	}
}

// showWelcome displays the welcome message
func showWelcome() {
	headerLine := "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	titleLine := "  🚀 Go ActionHero"

	logger.Info(color.New(color.FgBlue, color.Bold).Sprint(headerLine))
	logger.Info(color.New(color.FgBlue, color.Bold).Sprint(titleLine))
	logger.Info(color.New(color.FgBlue, color.Bold).Sprint(headerLine))
	logger.Info(color.New(color.FgCyan).Sprintf("  Process: %s", cfg.Process.Name))
	logger.Info(color.New(color.FgCyan).Sprintf("  Logger Level: %s", cfg.Logger.Level))
	logger.Info(color.New(color.FgCyan).Sprintf("  Web Server: %s:%d", cfg.Server.Web.Host, cfg.Server.Web.Port))
	logger.Info(color.New(color.FgBlue, color.Bold).Sprint(headerLine))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
