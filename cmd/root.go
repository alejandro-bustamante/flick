package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd is the base command with no arguments
var RootCmd = &cobra.Command{
	Use:   "flick",
	Short: "Automated movies and series organizer.",
	Long:  `Flick watches your folders and organizes your movies and series files. It renames and moves them to a destination folder.`,
}

// Add all the sub commands to the root command
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Ruta al archivo de config (default es ./settings.toml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigName("settings")
		viper.SetConfigType("toml")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Usando archivo de config:", viper.ConfigFileUsed())
	}
}
