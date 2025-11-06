package cmd

import (
	"log"
	"time"

	// Imports de tu lógica principal
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
	Short: "Inicia el daemon de Flick en primer plano.",
	Long:  `Inicia el servicio principal de monitoreo y organización.`,
	Run: func(cmd *cobra.Command, args []string) {

		// patterns.toml is loaded separetely from settings.toml
		data, err := config.LoadData("./patterns.toml")
		if err != nil {
			log.Fatalf("Error al cargar patterns.toml: %v", err)
		}

		logger := utils.NewLogger("debug")
		p := parser.NewMediaParser(
			data.Tokenizer.Separators,
			data.Cleaner.JunkPatterns,
			data.Extractor.YearRange[:],
			logger,
		)

		apiKey := viper.GetString("secrets.tmdb_api_key")
		if apiKey == "" {
			log.Fatalf("Error: 'secrets.tmdb_api_key' no encontrada en settings.toml")
		}
		f := finder.NewTMDBFinder(apiKey, p)

		// --- Watcher ---
		watchDir := viper.GetString("directories.watch")
		if watchDir == "" {
			log.Fatalf("Error: 'directories.watch' no encontrada en settings.toml")
		}
		watcherConfig := watcher.WatcherConfig{
			Path:           watchDir,
			StabilityDelay: 2 * time.Second,
			Recursive:      true,
		}
		folderWatcher, err := watcher.NewWatcher(watcherConfig)
		if err != nil {
			log.Fatalf("Error al crear el watcher: %v", err)
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
			log.Fatalf("No se pudo iniciar el daemon: %v", err)
		}
		log.Println("Flick daemon iniciado. Presiona Ctrl+C para detener.")
		flickDaemon.Start()

		log.Println("Flick se ha detenido.")
	},
}
