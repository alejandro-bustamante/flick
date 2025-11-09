package cmd

import (
	"log"
	"time"

	config "github.com/alejandro-bustamante/flick/internal/config"
	"github.com/alejandro-bustamante/flick/internal/core"
	finder "github.com/alejandro-bustamante/flick/internal/core/finder"
	parser "github.com/alejandro-bustamante/flick/internal/core/parser"
	"github.com/alejandro-bustamante/flick/internal/daemon"
	"github.com/alejandro-bustamante/flick/internal/utils"
	"github.com/alejandro-bustamante/flick/internal/watcher"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	RootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the Flick daemon in the foreground.",
	Long:  `Starts the main monitoring and organizing service.`,
	Run: func(cmd *cobra.Command, args []string) {

		_, patternsPath, err := config.GetConfigPaths(cfgFile)
		if err != nil {
			log.Fatalf("Error getting config paths: %v", err)
		}

		data, err := config.LoadPatterns(patternsPath) // Use LoadPatterns
		if err != nil {
			log.Fatalf("Error loading patterns.toml from %s: %v", patternsPath, err)
		}

		logger := utils.NewLogger("debug")
		p := parser.NewMediaParser(
			data.Tokenizer.Separators,
			data.Cleaner.JunkPatterns,
			data.Extractor.YearRange[:],
			logger,
		)

		apiKey := viper.GetString("secrets.tmdb_api_key")
		if apiKey == "" || apiKey == "YOUR_API_KEY_HERE" {
			log.Fatalf("Error: 'secrets.tmdb_api_key' not found or not set in %s", viper.ConfigFileUsed())
		}
		f := finder.NewTMDBFinder(apiKey, p)

		// --- Watcher ---
		watchDir := viper.GetString("directories.watch")
		if watchDir == "" {
			log.Fatalf("Error: 'directories.watch' not found in %s", viper.ConfigFileUsed())
		}
		watcherConfig := watcher.WatcherConfig{
			Path:           watchDir,
			StabilityDelay: 2 * time.Second,
			Recursive:      true,
		}
		folderWatcher, err := watcher.NewWatcher(watcherConfig)
		if err != nil {
			log.Fatalf("Error creating watcher: %v", err)
		}

		// --- Organizer ---
		organizer := core.NewOrganizer(
			p,
			f,
			folderWatcher,
			viper.GetString("directories.movies"),
			viper.GetString("directories.series"),
		)
		organizer.Run()

		// --- Daemon ---
		flickDaemon, err := daemon.NewDaemon(organizer)
		if err != nil {
			log.Fatalf("Could not start daemon: %v", err)
		}
		log.Println("Flick daemon started. Press Ctrl+C to stop.")
		flickDaemon.Start()

		log.Println("Flick has stopped.")
	},
}
