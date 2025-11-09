package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alejandro-bustamante/flick/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(serviceCmd)
	serviceCmd.AddCommand(serviceInstallCmd)
	serviceCmd.AddCommand(serviceUninstallCmd)
}

const serviceTemplate = `[Unit]
Description=Flick media organizer daemon
After=network-online.target

[Service]
ExecStart=%s start
WorkingDirectory=%s
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
`

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Manage the Flick systemd service",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install Flick as a systemd user service",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Println("Error getting executable path:", err)
			os.Exit(1)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			os.Exit(1)
		}

		serviceDir := filepath.Join(homeDir, ".config/systemd/user")
		servicePath := filepath.Join(serviceDir, "flick.service")

		serviceContent := fmt.Sprintf(serviceTemplate, exePath, homeDir)

		if err := os.MkdirAll(serviceDir, 0750); err != nil {
			fmt.Printf("Error creating service directory (%s): %v\n", serviceDir, err)
			os.Exit(1)
		}

		if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
			fmt.Printf("Error writing service file (%s): %v\n", servicePath, err)
			os.Exit(1)
		}

		fmt.Printf("Service file created at: %s!\n\n", servicePath)

		fmt.Println("IMPORTANT: You must configure Flick before starting the service.")

		settingsPath, _, err := config.GetConfigPaths("")
		if err != nil {
			fmt.Printf("Warning: Could not determine default config path: %v\n", err)
			fmt.Println("Please create your config file manually (e.g., $HOME/.config/flick/settings.toml).")
		} else {
			fmt.Printf("Please edit your config file at:\n%s\n", settingsPath)
			fmt.Println("You MUST add your 'tmdb_api_key' and set your 'directories'.")
		}

		fmt.Println("\nAfter editing the config, run these commands to start the service:")
		fmt.Println("  systemctl --user daemon-reload")
		fmt.Println("  systemctl --user enable --now flick")
		fmt.Println("\nTo have the service start automatically on boot (without logging in):")
		fmt.Println("  loginctl enable-linger $(whoami)")
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the Flick user service",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting home directory:", err)
			os.Exit(1)
		}
		servicePath := filepath.Join(homeDir, ".config/systemd/user/flick.service")

		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			fmt.Println("The service does not seem to be installed.")
			return
		}

		if err := os.Remove(servicePath); err != nil {
			fmt.Printf("Error removing service file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Service file removed!")
		fmt.Println("To stop and disable the service (if running):")
		fmt.Println("  systemctl --user disable --now flick")
		fmt.Println("  systemctl --user daemon-reload")
	},
}
