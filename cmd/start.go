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
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var organizer *core.Organizer

func init() {
	RootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the Flick daemon in the foreground.",
	Long:  `Starts the main monitoring and organizing service.`,
	Run: func(cmd *cobra.Command, args []string) {

		settingsPath, patternsPath, err := config.GetConfigPaths(cfgFile)
		if err != nil {
			log.Fatalf("Error getting config paths: %v", err)
		}

		data, err := config.LoadPatterns(patternsPath)
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
		f := finder.NewTMDBFinder(apiKey, p)

		// --- Watcher ---
		watchDir := viper.GetString("directories.watch")
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
		organizer = core.NewOrganizer(
			p,
			f,
			folderWatcher,
			viper.GetString("directories.movies"),
			viper.GetString("directories.series"),
		)
		organizer.Run()

		// --- Config Watcher ---
		go watchConfig(settingsPath, patternsPath)

		// --- Daemon ---
		flickDaemon, err := daemon.NewDaemon(organizer, reloadConfig)
		if err != nil {
			log.Fatalf("Could not start daemon: %v", err)
		}
		log.Println("Flick daemon started. Press Ctrl+C to stop.")
		flickDaemon.Start()

		log.Println("Flick has stopped.")
	},
}

func watchConfig(settingsPath, patternsPath string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating config watcher: %v", err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(settingsPath)
	if err != nil {
		log.Printf("Error watching settings file %s: %v", settingsPath, err)
	}
	err = watcher.Add(patternsPath)
	if err != nil {
		log.Printf("Error watching patterns file %s: %v", patternsPath, err)
	}

	log.Println("Config watcher started. Monitoring:", settingsPath, "and", patternsPath)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				log.Println("Config watcher channel closed.")
				return
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
				log.Println("Config file modification detected:", event.Name, "Op:", event.Op)

				if event.Op&fsnotify.Rename != 0 {
					watcher.Remove(event.Name)
					watcher.Add(event.Name)
				}

				reloadConfig()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				log.Println("Config watcher error channel closed.")
				return
			}
			log.Println("Config watcher error:", err)
		}
	}
}

func reloadConfig() {
	settingsPath, patternsPath, err := config.GetConfigPaths(cfgFile)
	if err != nil {
		log.Printf("Error getting config paths on reload: %v", err)
		return
	}

	viper.SetConfigFile(settingsPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Error reading config file on reload (%s): %v", settingsPath, err)
		return
	}

	log.Println("Configuration reloaded successfully from:", viper.ConfigFileUsed())

	data, err := config.LoadPatterns(patternsPath)
	if err != nil {
		log.Printf("Error reloading patterns.toml: %v", err)
	} else {
		p := parser.NewMediaParser(
			data.Tokenizer.Separators,
			data.Cleaner.JunkPatterns,
			data.Extractor.YearRange[:],
			utils.NewLogger("debug"),
		)

		apiKey := viper.GetString("secrets.tmdb_api_key")
		moviesDir := viper.GetString("directories.movies")
		seriesDir := viper.GetString("directories.series")
		watchDir := viper.GetString("directories.watch")

		log.Printf("[Reload] API Key Set: %t", (apiKey != "" && apiKey != "YOUR_API_KEY_HERE"))
		log.Printf("[Reload] Movies Dir: '%s'", moviesDir)
		log.Printf("[Reload] Series Dir: '%s'", seriesDir)
		log.Printf("[Reload] Watch Dir: '%s'", watchDir)

		f := finder.NewTMDBFinder(apiKey, p)

		organizer.UpdateConfig(
			p,
			f,
			moviesDir,
			seriesDir,
		)
		log.Println("Organizer configuration has been updated.")
	}
}
