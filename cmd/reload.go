package cmd

import (
	"log"

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
		if organizer == nil {
			log.Println("Error: Flick is not running. Please start it first.")
			return
		}
		log.Println("Reloading configuration...")
		reloadConfig()
	},
}
