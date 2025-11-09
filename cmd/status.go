package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/alejandro-bustamante/flick/internal/daemon"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the Flick daemon.",
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Println("--- Config Status ---")
		apiKey := viper.GetString("secrets.tmdb_api_key")
		watchDir := viper.GetString("directories.watch")

		configOK := true
		if apiKey == "" || apiKey == "YOUR_API_KEY_HERE" {
			fmt.Println("API Key: NOT SET")
			configOK = false
		} else {
			fmt.Println("API Key: Set")
		}

		if watchDir == "" {
			fmt.Println("Directories: NOT SET")
			configOK = false
		} else {
			fmt.Println("Directories: Set")
		}

		if !configOK {
			fmt.Printf("\nConfig file is incomplete: %s\n", viper.ConfigFileUsed())
			fmt.Println("The organizer will remain idle until configured.")
		}
		fmt.Println("---------------------")

		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			log.Fatalf("Daemon: Not running. (Is it enabled? 'systemctl --user enable --now flick')\nError: %v", err)
		}
		defer conn.Close()

		fmt.Println("Daemon: Connected")
		fmt.Println("--- Daemon Status ---")

		// "STATUS"
		_, err = conn.Write([]byte("STATUS\n"))
		if err != nil {
			log.Fatalf("Error sending command to daemon: %v", err)
		}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Error reading response from daemon: %v", err)
		}
		fmt.Println("---------------------")
	},
}
