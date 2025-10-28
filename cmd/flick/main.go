package main

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
)

func main() {
	// --- Load config ---
	data, err := config.LoadData("/home/alejandro/nvme/Repositorios/Developer/flick/patterns.toml")
	if err != nil {
		log.Fatalf("Error al cargar patterns.toml: %v", err)
	}

	sttgs, err := config.LoadSettings("/home/alejandro/nvme/Repositorios/Developer/flick/settings.toml")
	if err != nil {
		log.Fatalf("Error al cargar settings.toml: %v", err)
	}

	// --- Initialize components ---
	logger := utils.NewLogger("debug")

	p := parser.NewMediaParser(
		data.Tokenizer.Separators,
		data.Cleaner.JunkPatterns,
		data.Extractor.YearRange[:],
		logger,
	)

	f := finder.NewTMDBFinder(sttgs.Secrets.TMDB_API_Key)
	o := core.NewOrganizer(p, f, sttgs.Directories.Watch, sttgs.Directories.Movies, sttgs.Directories.Series)

	watcherConfig := watcher.WatcherConfig{
		Path:           sttgs.Directories.Watch,
		StabilityDelay: 2 * time.Second,
		Recursive:      true,
	}

	folderWatcher, err := watcher.NewWatcher(watcherConfig, o)
	if err != nil {
		log.Fatalf("Error al crear el watcher: %v", err)
	}

	// --- Start watcher ---
	err = folderWatcher.Start()
	if err != nil {
		log.Fatalf("Error al iniciar el watcher: %v", err)
	}

	// --- Start the daemon (blocking function) ---
	flickDaemon, err := daemon.NewDaemon()
	if err != nil {
		log.Fatalf("No se pudo iniciar el daemon: %v", err)
	}

	flickDaemon.Start()

	log.Println("Flick se ha detenido.")
}
