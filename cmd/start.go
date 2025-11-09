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
		flickDaemon, err := daemon.NewDaemon(organizer)
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
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("Config file modified:", event.Name)
					reloadConfig()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(settingsPath)
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Add(patternsPath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func reloadConfig() {
	_, patternsPath, err := config.GetConfigPaths(cfgFile)
	if err != nil {
		log.Printf("Error getting config paths: %v", err)
		return
	}
	viper.ReadInConfig()
	data, err := config.LoadPatterns(patternsPath)
	if err != nil {
		log.Printf("Error reloading patterns.toml: %v", err)
	} else {
		// Update organizer
		p := parser.NewMediaParser(
			data.Tokenizer.Separators,
			data.Cleaner.JunkPatterns,
			data.Extractor.YearRange[:],
			utils.NewLogger("debug"),
		)
		apiKey := viper.GetString("secrets.tmdb_api_key")
		f := finder.NewTMDBFinder(apiKey, p)
		organizer.UpdateConfig(
			p,
			f,
			viper.GetString("directories.movies"),
			viper.GetString("directories.series"),
		)
	}
}
