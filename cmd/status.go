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
	RootCmd.AddCommand(statusCmd)
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current status of the Flick daemon.",
	Run: func(cmd *cobra.Command, args []string) {

		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			log.Fatalf("Could not connect to Flick daemon. Is it running? (use 'flick start')\nError: %v", err)
		}
		defer conn.Close()

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
	},
}
