package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"os/user"
	"reflect"
	"syscall"

	"github.com/evantahler/go-actionhero/actions"
	"github.com/evantahler/go-actionhero/internal/api"
	"github.com/evantahler/go-actionhero/internal/config"
	"github.com/evantahler/go-actionhero/internal/servers"
	"github.com/evantahler/go-actionhero/internal/util"
	"github.com/fatih/color"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	Run: func(cmd *cobra.Command, _ []string) {
		// Disable timestamps for help command
		disableTimestampsForCommand()
		showWelcome()
		_ = cmd.Help()
	},
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the ActionHero server",
	Long:  `Start the ActionHero server and begin accepting connections.`,
	Run: func(_ *cobra.Command, _ []string) {
		startServer()
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display current configuration",
	Long:  `Display the current application configuration in a formatted way.`,
	PreRun: func(cmd *cobra.Command, _ []string) {
		// Disable timestamps for config command
		disableTimestampsForCommand()
		// Skip welcome message for JSON output
		format, _ := cmd.Flags().GetString("format")
		if format != "json" {
			showWelcome()
		}
	},
	Run: func(cmd *cobra.Command, _ []string) {
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

	// Register action commands
	registerActionCommands()
}

// registerActionCommands adds each action as a CLI command
func registerActionCommands() {
	// Get all auto-registered actions
	for _, action := range actions.GetAll() {
		addActionCommand(action)
	}
}

// addActionCommand creates a CLI command for an action
func addActionCommand(action api.Action) {
	actionName := api.GetActionName(action)
	actionDesc := api.GetActionDescription(action)

	cmd := &cobra.Command{
		Use:   actionName,
		Short: fmt.Sprintf("Run action: %s", actionName),
		Long: fmt.Sprintf("Run action: %s\n\n%s\n\nInputs should be passed as flags. The server will be initialized and started, and the action will be executed via a CLI connection.",
			actionName, actionDesc),
		Run: func(cmd *cobra.Command, args []string) {
			runActionViaCLI(cmd, action)
		},
	}

	// Add flags for action inputs
	inputs := api.GetActionInputs(action)
	if inputs != nil {
		inputType := reflect.TypeOf(inputs)
		if inputType.Kind() == reflect.Struct {
			for i := 0; i < inputType.NumField(); i++ {
				field := inputType.Field(i)
				jsonTag := field.Tag.Get("json")
				if jsonTag != "" && jsonTag != "-" {
					// Get the field name from JSON tag
					flagName := jsonTag

					// Check if required (simplified - could be enhanced with validate tag parsing)
					validateTag := field.Tag.Get("validate")
					isRequired := validateTag != "" && (validateTag == "required" ||
						len(validateTag) > 8 && validateTag[:8] == "required")

					description := fmt.Sprintf("%s parameter", flagName)

					// Add the flag based on type
					switch field.Type.Kind() {
					case reflect.String:
						if isRequired {
							cmd.Flags().String(flagName, "", description+" (required)")
							_ = cmd.MarkFlagRequired(flagName)
						} else {
							cmd.Flags().String(flagName, "", description)
						}
					case reflect.Int, reflect.Int64:
						if isRequired {
							cmd.Flags().Int(flagName, 0, description+" (required)")
							_ = cmd.MarkFlagRequired(flagName)
						} else {
							cmd.Flags().Int(flagName, 0, description)
						}
					case reflect.Bool:
						cmd.Flags().Bool(flagName, false, description)
					default:
						// For other types, use string and let action parse it
						cmd.Flags().String(flagName, "", description)
					}
				}
			}
		}
	}

	rootCmd.AddCommand(cmd)
}

// runActionViaCLI executes an action via CLI connection
func runActionViaCLI(cmd *cobra.Command, action api.Action) {
	// Create API instance
	apiInstance := api.New(cfg, logger)

	// Register all actions
	for _, action := range actions.GetAll() {
		if err := apiInstance.RegisterAction(action); err != nil {
			logger.Fatalf("Failed to register action: %v", err)
		}
	}

	// Initialize API (but don't start servers)
	if err := apiInstance.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize: %v", err)
	}

	// Get current user for connection ID
	currentUser, _ := user.Current()
	username := "cli"
	if currentUser != nil {
		username = currentUser.Username
	}
	connectionID := fmt.Sprintf("cli:%s", username)

	// Create CLI connection
	conn := api.NewConnection("cli", connectionID, connectionID, nil)

	// Collect parameters from flags
	params := make(map[string]interface{})
	cmd.Flags().Visit(func(flag *pflag.Flag) {
		// Skip global flags
		if flag.Name == "no-color" || flag.Name == "no-timestamp" || flag.Name == "quiet" {
			return
		}
		params[flag.Name] = flag.Value.String()
	})

	// Execute action
	actionName := api.GetActionName(action)
	result := conn.Act(context.Background(), apiInstance, actionName, params, "CLI", "")

	// Prepare output
	output := map[string]interface{}{
		"response": result.Response,
	}
	exitCode := 0

	if result.Error != nil {
		exitCode = 1
		if typedErr, ok := result.Error.(*util.TypedError); ok {
			output["error"] = map[string]interface{}{
				"message": typedErr.Message,
				"code":    typedErr.Code(),
				"type":    typedErr.Type,
			}
		} else {
			output["error"] = map[string]interface{}{
				"message": result.Error.Error(),
			}
		}
	}

	// Output JSON to stdout (or stderr if error)
	jsonOutput, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		logger.Fatalf("Failed to marshal output: %v", err)
	}

	if exitCode == 0 {
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Fprintln(os.Stderr, string(jsonOutput))
	}

	os.Exit(exitCode)
}

// loadConfigAndInitLogger loads configuration and initializes the logger
// This runs before any command execution
func loadConfigAndInitLogger(_ *cobra.Command, _ []string) error {
	var err error

	// Load configuration
	cfg, err = config.Load()
	if err != nil {
		_, _ = color.New(color.FgRed).Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
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

// startServer initializes and starts the ActionHero server
func startServer() {
	showWelcome()

	// Create API instance
	apiInstance := api.New(cfg, logger)

	// Register all actions
	for _, action := range actions.GetAll() {
		if err := apiInstance.RegisterAction(action); err != nil {
			logger.Fatalf("Failed to register action: %v", err)
		}
	}

	// Register web server
	webServer := servers.NewWebServer(apiInstance)
	apiInstance.RegisterServer(webServer)

	// Initialize API
	logger.Info("Initializing...")
	if err := apiInstance.Initialize(); err != nil {
		logger.Fatalf("Failed to initialize: %v", err)
	}

	// Start API
	logger.Info("Starting...")
	if err := apiInstance.Start(); err != nil {
		logger.Fatalf("Failed to start: %v", err)
	}

	logger.Info(color.GreenString("Server is running! Press Ctrl+C to stop."))

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Graceful shutdown
	logger.Info("Shutting down gracefully...")
	if err := apiInstance.Stop(); err != nil {
		logger.Errorf("Error during shutdown: %v", err)
		os.Exit(1)
	}

	logger.Info(color.GreenString("Server stopped successfully"))
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
