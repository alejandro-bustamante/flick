package cmd

import (
	"fmt"
	"os"
	"path/filepath"

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
	Short: "Gestiona el servicio de systemd de Flick",
}

var serviceInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Instala Flick como un servicio de usuario de systemd",
	Run: func(cmd *cobra.Command, args []string) {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Println("Error al obtener la ruta del ejecutable:", err)
			os.Exit(1)
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error al obtener el directorio home:", err)
			os.Exit(1)
		}

		serviceDir := filepath.Join(homeDir, ".config/systemd/user")
		servicePath := filepath.Join(serviceDir, "flick.service")

		serviceContent := fmt.Sprintf(serviceTemplate, exePath, homeDir)

		if err := os.MkdirAll(serviceDir, 0750); err != nil {
			fmt.Printf("Error al crear el directorio del servicio (%s): %v\n", serviceDir, err)
			os.Exit(1)
		}

		if err := os.WriteFile(servicePath, []byte(serviceContent), 0644); err != nil {
			fmt.Printf("Error al escribir el archivo de servicio (%s): %v\n", servicePath, err)
			os.Exit(1)
		}

		fmt.Printf("¡Archivo de servicio creado en: %s!\n\n", servicePath)
		fmt.Println("Para habilitar e iniciar el servicio, ejecuta:")
		fmt.Println("  systemctl --user daemon-reload")
		fmt.Println("  systemctl --user enable --now flick")
		fmt.Println("\nPara que el servicio se inicie automáticamente al arrancar (sin iniciar sesión):")
		fmt.Println("  loginctl enable-linger $(whoami)")
	},
}

var serviceUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Desinstala el servicio de usuario de Flick",
	Run: func(cmd *cobra.Command, args []string) {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error al obtener el directorio home:", err)
			os.Exit(1)
		}
		servicePath := filepath.Join(homeDir, ".config/systemd/user/flick.service")

		if _, err := os.Stat(servicePath); os.IsNotExist(err) {
			fmt.Println("El servicio no parece estar instalado.")
			return
		}

		if err := os.Remove(servicePath); err != nil {
			fmt.Printf("Error al eliminar el archivo de servicio: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("¡Archivo de servicio eliminado!")
		fmt.Println("Para detener y deshabilitar el servicio (si se está ejecutando):")
		fmt.Println("  systemctl --user disable --now flick")
		fmt.Println("  systemctl --user daemon-reload")
	},
}
