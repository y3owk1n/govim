package cli

import (
	"fmt"
	"os"

	"github.com/y3owk1n/govim/internal/config"
	"go.uber.org/zap"
)

// Command represents a CLI command
type Command struct {
	Name        string
	Description string
	Handler     func(args []string) error
}

// CLI manages command-line interface
type CLI struct {
	commands map[string]*Command
	logger   *zap.Logger
	config   *config.Config
}

// NewCLI creates a new CLI
func NewCLI(cfg *config.Config, logger *zap.Logger) *CLI {
	cli := &CLI{
		commands: make(map[string]*Command),
		logger:   logger,
		config:   cfg,
	}

	cli.registerCommands()
	return cli
}

// registerCommands registers all available commands
func (c *CLI) registerCommands() {
	c.Register(&Command{
		Name:        "hint",
		Description: "Activate hint mode",
		Handler:     c.hintCommand,
	})

	c.Register(&Command{
		Name:        "scroll",
		Description: "Activate scroll mode",
		Handler:     c.scrollCommand,
	})

	c.Register(&Command{
		Name:        "exit",
		Description: "Exit current mode",
		Handler:     c.exitCommand,
	})

	c.Register(&Command{
		Name:        "status",
		Description: "Show current status",
		Handler:     c.statusCommand,
	})

	c.Register(&Command{
		Name:        "reload-config",
		Description: "Reload configuration",
		Handler:     c.reloadConfigCommand,
	})

	c.Register(&Command{
		Name:        "list-elements",
		Description: "List UI elements (debugging)",
		Handler:     c.listElementsCommand,
	})

	c.Register(&Command{
		Name:        "click",
		Description: "Click on element by ID",
		Handler:     c.clickCommand,
	})

	c.Register(&Command{
		Name:        "validate-config",
		Description: "Validate configuration file",
		Handler:     c.validateConfigCommand,
	})

	c.Register(&Command{
		Name:        "help",
		Description: "Show help information",
		Handler:     c.helpCommand,
	})

	c.Register(&Command{
		Name:        "version",
		Description: "Show version information",
		Handler:     c.versionCommand,
	})
}

// Register registers a command
func (c *CLI) Register(cmd *Command) {
	c.commands[cmd.Name] = cmd
}

// Execute executes a command
func (c *CLI) Execute(args []string) error {
	if len(args) == 0 {
		return c.helpCommand(args)
	}

	cmdName := args[0]
	cmd, ok := c.commands[cmdName]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmdName)
	}

	return cmd.Handler(args[1:])
}

// Command handlers

func (c *CLI) hintCommand(args []string) error {
	fmt.Println("Activating hint mode...")
	// This would send a signal to the running GoVim instance
	// For now, just print a message
	return nil
}

func (c *CLI) scrollCommand(args []string) error {
	fmt.Println("Activating scroll mode...")
	return nil
}

func (c *CLI) exitCommand(args []string) error {
	fmt.Println("Exiting current mode...")
	return nil
}

func (c *CLI) statusCommand(args []string) error {
	fmt.Println("GoVim Status:")
	fmt.Println("  Mode: Idle")
	fmt.Println("  Config: " + config.GetConfigPath())
	return nil
}

func (c *CLI) reloadConfigCommand(args []string) error {
	fmt.Println("Reloading configuration...")
	
	newConfig, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	c.config = newConfig
	fmt.Println("Configuration reloaded successfully")
	return nil
}

func (c *CLI) listElementsCommand(args []string) error {
	appName := ""
	if len(args) > 0 {
		for i := 0; i < len(args); i++ {
			if args[i] == "--app" && i+1 < len(args) {
				appName = args[i+1]
				break
			}
		}
	}

	if appName != "" {
		fmt.Printf("Listing elements for app: %s\n", appName)
	} else {
		fmt.Println("Listing elements for frontmost window...")
	}

	// This would actually query accessibility elements
	fmt.Println("(Not implemented in CLI mode)")
	return nil
}

func (c *CLI) clickCommand(args []string) error {
	if len(args) < 2 || args[0] != "--element-id" {
		return fmt.Errorf("usage: govim click --element-id <id>")
	}

	elementID := args[1]
	fmt.Printf("Clicking element: %s\n", elementID)
	return nil
}

func (c *CLI) validateConfigCommand(args []string) error {
	configPath := config.GetConfigPath()
	if len(args) > 0 {
		configPath = args[0]
	}

	fmt.Printf("Validating configuration: %s\n", configPath)

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	fmt.Println("✓ Configuration is valid")
	return nil
}

func (c *CLI) helpCommand(args []string) error {
	fmt.Println("GoVim - Keyboard-driven navigation for macOS")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  govim [command] [options]")
	fmt.Println()
	fmt.Println("Available Commands:")

	for name, cmd := range c.commands {
		fmt.Printf("  %-20s %s\n", name, cmd.Description)
	}

	fmt.Println()
	fmt.Println("Run 'govim help [command]' for more information about a command.")
	return nil
}

func (c *CLI) versionCommand(args []string) error {
	fmt.Println("GoVim version 0.1.0")
	fmt.Println("Built with Go")
	return nil
}

// Run runs the CLI with the given arguments
func Run(args []string) {
	// Load config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Create CLI
	cli := NewCLI(cfg, logger)

	// Execute command
	if err := cli.Execute(args); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
