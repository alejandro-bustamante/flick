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
	Short: "Muestra el estado actual del daemon de Flick.",
	Run: func(cmd *cobra.Command, args []string) {

		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			log.Fatalf("No se pudo conectar al daemon de Flick. ¿Está corriendo? (usa 'flick start')\nError: %v", err)
		}
		defer conn.Close()

		// "STATUS"
		_, err = conn.Write([]byte("STATUS\n"))
		if err != nil {
			log.Fatalf("Error al enviar comando al daemon: %v", err)
		}

		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}

		if err := scanner.Err(); err != nil {
			log.Fatalf("Error al leer respuesta del daemon: %v", err)
		}
	},
}
