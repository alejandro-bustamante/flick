package cmd

import (
	"bufio"
	"fmt"
	"log"
	"net"

	"github.com/alejandro-bustamante/flick/internal/daemon"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(reloadCmd)
}

var reloadCmd = &cobra.Command{
	Use:   "reload",
	Short: "Reloads the configuration.",
	Long:  `Reloads the settings.toml and patterns.toml files.`,
	Run: func(cmd *cobra.Command, args []string) {
		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			log.Fatalf("Could not connect to Flick daemon. Is it running? (use 'flick start')\nError: %v", err)
		}
		defer conn.Close()

		_, err = conn.Write([]byte("RELOAD\n"))
		if err != nil {
			log.Fatalf("Error sending command to daemon: %v", err)
		}

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading response from daemon: %v", err)
		}

		fmt.Print(response)
	},
}
