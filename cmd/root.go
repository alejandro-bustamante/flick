package cmd

import (
	"fmt"
	"os"

	"github.com/alejandro-bustamante/flick/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

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
	// Updated help string to be dynamic (dev vs prod)
	defaultPathDesc := "default is ./settings.toml (dev) or $HOME/.config/flick/settings.toml (release)"
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", fmt.Sprintf("Config file path (%s)", defaultPathDesc))
}

func initConfig() {
	settingsPath, _, err := config.GetConfigPaths(cfgFile)
	if err != nil {
		fmt.Printf("Error initializing config: %v\n", err)
		os.Exit(1)
	}

	viper.SetConfigFile(settingsPath)
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Error: Config file not found at %s. A default was created.\n", settingsPath)
			fmt.Println("Please edit it to add your TMDB API key and paths.")
		} else {
			fmt.Printf("Error reading config file: %v\n", err)
		}
	} else {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
