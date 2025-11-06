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
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Muestra la versión de Flick.",
	Run: func(cmd *cobra.Command, args []string) {

		conn, err := net.Dial("unix", daemon.SocketPath)
		if err != nil {
			log.Fatalf("No se pudo conectar al daemon de Flick. ¿Está corriendo? (usa 'flick start')\nError: %v", err)
		}
		defer conn.Close()

		// "VERSION"
		_, err = conn.Write([]byte("VERSION\n"))
		if err != nil {
			log.Fatalf("Error al enviar comando al daemon: %v", err)
		}

		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Fatalf("Error al leer respuesta del daemon: %v", err)
		}

		fmt.Print(response)
	},
}
